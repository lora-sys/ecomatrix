"""Common graph helpers.

Each job graph follows the same three-node shape:

    observe -> think -> act -> END

* ``observe`` reads the agent's current balance, recent receipts, and the
  social feed (when present in Phase 2 follow-ups).
* ``think`` calls the LLM with a strategy prompt and parses a structured
  decision (JSON).
* ``act`` either submits an A2A trade or skips (no-op) the tick.
"""

from __future__ import annotations

import json
import re
from dataclasses import dataclass
from typing import Any, TypedDict

from langgraph.graph import END, StateGraph

from ..a2a import A2AClient, A2AError, Receipt
from ..llm import LLM


class AgentState(TypedDict, total=False):
    agent_id: str
    job_type: str
    balance: int
    observations: list[str]
    last_receipts: list[dict[str, Any]]
    decision: dict[str, Any]
    last_action: dict[str, Any]
    last_error: str
    last_receipt: dict[str, Any]


@dataclass
class GraphResult:
    decision: dict[str, Any]
    receipt: Receipt | None
    error: A2AError | None


def _parse_decision(raw: str) -> dict[str, Any]:
    """Tolerate LLMs that wrap JSON in markdown fences or add preamble."""
    m = re.search(r"\{.*\}", raw, flags=re.DOTALL)
    if not m:
        raise ValueError(f"LLM returned no JSON: {raw!r}")
    return json.loads(m.group(0))


def make_graph(job_type: str, *, llm: LLM, client: A2AClient,
               strategy_prompt: str) -> Any:
    """Build a LangGraph compiled graph for a given job type."""

    def observe(state: AgentState) -> dict[str, Any]:
        agent = client.get_agent(state["agent_id"])
        balance = int(agent.get("Balance", agent.get("balance", 0)))
        obs = list(state.get("observations", []))
        obs.append(f"tick: balance={balance} GOLD, job={state['job_type']}")
        return {"balance": balance, "observations": obs}

    def think(state: AgentState) -> dict[str, Any]:
        prompt = (
            f"{strategy_prompt}\n"
            f"agent_id={state['agent_id']} balance={state['balance']} "
            f"job_type={state['job_type']} target=agent_merchant_01 amount=10"
        )
        raw = llm.complete(
            [{"role": "system", "content": "You are an EcoMatrix agent. "
             "Reply with a single JSON object."},
             {"role": "user", "content": prompt}])
        try:
            decision = _parse_decision(raw)
        except ValueError as e:
            decision = {"action": "SKIP", "reason": str(e)}
        return {"decision": decision}

    def act(state: AgentState) -> dict[str, Any]:
        d = state.get("decision") or {}
        action = str(d.get("action", "SKIP")).upper()
        if action != "EXECUTE_TRADE":
            return {"last_action": d}
        target = str(d.get("target_agent", ""))
        amount = int(d.get("amount", 0))
        reasoning = str(d.get("reasoning", ""))
        try:
            receipt = client.execute_trade(
                sender=state["agent_id"],
                target_agent=target,
                amount=amount,
                reasoning=reasoning,
            )
            receipts = list(state.get("last_receipts", []))
            receipts.append({
                "tx_id": receipt.tx_id,
                "to": receipt.to,
                "amount": receipt.amount,
            })
            return {
                "last_action": d,
                "last_receipt": {
                    "tx_id": receipt.tx_id, "to": receipt.to, "amount": receipt.amount,
                },
                "last_receipts": receipts,
            }
        except A2AError as e:
            return {"last_action": d, "last_error": f"{e.code.value}:{e.message}"}

    g = StateGraph(AgentState)
    g.add_node("observe", observe)
    g.add_node("think", think)
    g.add_node("act", act)
    g.set_entry_point("observe")
    g.add_edge("observe", "think")
    g.add_edge("think", "act")
    g.add_edge("act", END)
    return g.compile()


def run_one_tick(graph: Any, *, agent_id: str, job_type: str) -> GraphResult:
    init: AgentState = {"agent_id": agent_id, "job_type": job_type}
    final = graph.invoke(init)
    decision = final.get("decision") or {}
    last_receipt = final.get("last_receipt")
    last_error = final.get("last_error")

    receipt = None
    if last_receipt:
        receipt = Receipt(
            tx_id=last_receipt["tx_id"],
            from_=agent_id,
            to=last_receipt["to"],
            amount=int(last_receipt["amount"]),
            currency_type="GOLD",
            balance_after_from=0,
            balance_after_to=0,
        )

    error = None
    if last_error:
        from ..a2a import Code
        try:
            code, _, msg = last_error.partition(":")
            error = A2AError(Code(code), msg)
        except ValueError:
            error = A2AError(Code.INTERNAL, last_error)

    return GraphResult(decision=decision, receipt=receipt, error=error)
