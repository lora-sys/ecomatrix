"""Common graph helpers.

The base graph wires the LLM and tool executor into the act node. Errors
are recovered by falling back to a deterministic action; the error is
logged to the conversation table so the dashboard can show what happened.
"""

from __future__ import annotations

import json
import re
import time
from dataclasses import dataclass
from typing import Any, TypedDict

from langgraph.graph import END, StateGraph

from ..a2a import A2AClient, A2AError, Receipt
from ..llm import (
    LLM,
    LLMError,
    parse_action_json,
)
from ..tools import TOOL_SCHEMA, execute_tools_in_sequence


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
    last_feed: dict[str, Any]
    # Cached latest LLM result (for the dashboard's "AI Thought Trace").
    last_llm_raw: str
    last_llm_thought: str
    tool_calls: list[dict[str, Any]]
    tool_results: list[str]
    # Conversation log entries (for the dashboard) — one per LLM call.
    conversation: list[dict[str, Any]]


@dataclass
class GraphResult:
    decision: dict[str, Any]
    receipt: Receipt | None
    error: LLMError | str | None
    feed_post_id: int | None = None


def _parse_decision(raw: str) -> dict[str, Any]:
    m = re.search(r"\{.*\}", raw, flags=re.DOTALL)
    if not m:
        raise ValueError(f"LLM returned no JSON: {raw!r}")
    return json.loads(m.group(0))


def _fallback_action(state: AgentState, reason: str) -> dict[str, Any]:
    """Deterministic SKIP action when the LLM is unavailable.

    The agent records *why* it skipped so the dashboard can show "LLM
    timeout" in the AI Thought Trace.
    """
    return {
        "action": "SKIP",
        "target_agent": "",
        "amount": 0,
        "reasoning": f"fallback: {reason}",
    }


def make_graph(
    job_type: str,
    *,
    llm: LLM,
    client: A2AClient,
    strategy_prompt: str,
    on_event: callable | None = None,
) -> Any:
    """Build a LangGraph compiled graph for a given job type.

    ``on_event`` is an optional hook called with (event_name, payload)
    on every LLM call / tool execution. The dashboard uses this to keep
    the conversation log live.
    """

    def observe(state: AgentState) -> dict[str, Any]:
        try:
            a = client.get_agent(state["agent_id"])
            balance = int(a.get("Balance", a.get("balance", 0)))
        except A2AError as e:
            return {"balance": 0, "observations": [f"observe failed: {e.message}"]}
        obs = list(state.get("observations", []))
        obs.append(f"tick: balance={balance} GOLD, job={state['job_type']}")
        return {"balance": balance, "observations": obs}

    def think(state: AgentState) -> dict[str, Any]:
        # Build the system prompt with strategy + tool schema.
        sys = (
            f"{strategy_prompt}\n\n"
            f"Current agent_id={state['agent_id']} balance={state['balance']} job_type={state['job_type']}\n"
            f"You may call tools to read state, execute trades, or post to the social feed. "
            f"Available tools: {json.dumps(TOOL_SCHEMA, ensure_ascii=False)}\n\n"
            f"Respond with a single JSON object. If you want to call tools, include a "
            f"'tool_calls' field as a list of {{\"name\": ..., \"arguments\": ...}} objects. "
            f"Otherwise set action='SKIP' or 'EXECUTE_TRADE' with target_agent, amount, reasoning."
        )
        user = (
            "Decide your next action. Pick a target_agent and amount when relevant. "
            "Use a target=... and amount=... hint in your reasoning text."
        )
        messages = [
            {"role": "system", "content": sys},
            {"role": "user", "content": user},
        ]
        t0 = time.time()
        try:
            raw = llm.complete(messages, temperature=0.4)
        except LLMError as e:
            elapsed = int((time.time() - t0) * 1000)
            entry = {
                "role": "error",
                "content": str(e),
                "error_code": type(e).__name__,
                "latency_ms": elapsed,
            }
            if on_event:
                try: on_event("llm_error", entry)
                except Exception: pass
            return {
                "decision": _fallback_action(state, str(e)),
                "last_error": str(e),
                "last_llm_raw": "",
                "last_llm_thought": f"LLM error: {e}",
                "conversation": [entry],
            }
        elapsed = int((time.time() - t0) * 1000)
        try:
            decision = parse_action_json(raw)
        except (ValueError, LLMError) as e:
            entry = {
                "role": "error",
                "content": f"parse failed: {e}",
                "error_code": "PARSE",
                "latency_ms": elapsed,
            }
            if on_event:
                try: on_event("llm_error", entry)
                except Exception: pass
            return {
                "decision": _fallback_action(state, f"parse: {e}"),
                "last_error": f"parse: {e}",
                "last_llm_raw": raw,
                "last_llm_thought": f"parse failed: {e}",
                "conversation": [entry],
            }
        # Success: log the conversation.
        thought = decision.get("thought") or decision.get("reasoning") or ""
        entry = {
            "role": "assistant",
            "content": json.dumps(decision, ensure_ascii=False),
            "latency_ms": elapsed,
        }
        if on_event:
            try: on_event("llm_ok", entry)
            except Exception: pass
        return {
            "decision": decision,
            "last_llm_raw": raw,
            "last_llm_thought": thought,
            "conversation": [entry],
        }

    def act(state: AgentState) -> dict[str, Any]:
        d = state.get("decision") or {}
        action = str(d.get("action", "SKIP")).upper()
        out: dict[str, Any] = {"last_action": d}

        # Optional tool calls first.
        tool_calls = d.get("tool_calls") or []
        if tool_calls:
            results = execute_tools_in_sequence(tool_calls, client=client, sender=state["agent_id"])
            out["tool_results"] = [r.to_json() for r in results]
            out["tool_calls"] = tool_calls
            if on_event:
                try:
                    on_event("tools", [{"name": r.name, "ok": r.ok} for r in results])
                except Exception: pass

        if action == "EXECUTE_TRADE":
            target = str(d.get("target_agent", ""))
            amount = int(d.get("amount", 0))
            reasoning = str(d.get("reasoning", ""))
            try:
                receipt = client.execute_trade(
                    sender=state["agent_id"],
                    target_agent=target, amount=amount, reasoning=reasoning,
                )
                receipts = list(state.get("last_receipts", []))
                receipts.append({
                    "tx_id": receipt.tx_id, "to": receipt.to, "amount": receipt.amount,
                })
                out["last_receipt"] = {
                    "tx_id": receipt.tx_id, "to": receipt.to, "amount": receipt.amount,
                }
                out["last_receipts"] = receipts
            except A2AError as e:
                out["last_error"] = f"trade: {e.code.value}: {e.message}"
            return out

        if action == "POST_FEED":
            content = str(d.get("content", ""))
            intent = str(d.get("intent_type", "SOCIAL"))
            try:
                post_id = client.post_feed(sender=state["agent_id"], content=content, intent_type=intent)
                out["last_feed"] = {"post_id": post_id, "content": content, "intent_type": intent}
            except A2AError as e:
                out["last_error"] = f"feed: {e.code.value}: {e.message}"
            return out

        # SKIP or unknown action.
        return out

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
    err = final.get("last_error") or None
    receipt = None
    feed = None
    if final.get("last_receipt"):
        r = final["last_receipt"]
        receipt = Receipt(
            tx_id=r["tx_id"], from_=agent_id, to=r["to"], amount=r["amount"],
            currency_type="GOLD", balance_after_from=0, balance_after_to=0,
        )
    if final.get("last_feed"):
        feed = final["last_feed"]["post_id"]
    return GraphResult(decision=decision, receipt=receipt, error=err, feed_post_id=feed)
