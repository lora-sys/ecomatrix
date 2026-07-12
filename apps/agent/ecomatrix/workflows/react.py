"""ReAct workflow: Reason -> Act -> Observe -> Reflect -> repeat or stop."""
from __future__ import annotations

import json
from dataclasses import dataclass
from typing import Any

from ..a2a import A2AClient
from ..contracts import get_contract
from ..llm import LLM, LLMError, parse_action_json
from ..observability import TraceClient
from ..tools import TOOL_SCHEMA, execute_tool


REACT_MAX_ITERATIONS = 5


@dataclass
class ReActResult:
    iterations: int
    final_action: dict | None
    final_receipt: dict | None
    transcript: list[dict]
    error: str | None = None


def _now() -> int:
    import time as _t
    return int(_t.time() * 1000)


def _build_reason_prompt(contract, goal, observation, transcript):
    transcript_str = "\n".join(
        f"- {s.get('step')}: {json.dumps({k: v for k, v in s.items() if k != 'latency_ms'}, ensure_ascii=False)[:300]}"
        for s in transcript[-5:]
    )
    return (
        f"{contract.system_prompt_section()}\n\n"
        f"## Your goal\n{goal}\n\n"
        f"## Tools available\n{json.dumps(TOOL_SCHEMA, ensure_ascii=False)}\n\n"
        f"## Initial observation\n{observation}\n\n"
        f"## Transcript (most recent)\n{transcript_str or '(none yet)'}\n\n"
        f"Decide your next action. Output JSON:\n"
        f'- {{ "thought": str, "action": "EXECUTE_TRADE" | "POST_FEED" | "SKIP", "target_agent": str, "amount": int, "reasoning": str, "tool_calls": [{{"name": str, "arguments": {{...}}}}] }}\n'
        f"- Use tool_calls to gather information before acting.\n"
        f"- If you have enough, output the final action."
    )


def _build_reflect_prompt(transcript):
    return (
        f"Reflect on your recent work.\n\n"
        f"## Transcript\n{json.dumps([{k: v for k, v in s.items() if k != 'latency_ms'} for s in transcript[-5:]], ensure_ascii=False, indent=2)}\n\n"
        f'Output JSON: {{ "continue": bool, "reason": str }}'
    )


def run_react(
    *,
    agent_id: str,
    job_type: str,
    llm: LLM,
    client: A2AClient,
    initial_observation: str,
    goal: str,
    max_iterations: int = REACT_MAX_ITERATIONS,
    trace: TraceClient | None = None,
) -> ReActResult:
    contract = get_contract(job_type)
    if trace is None:
        trace = TraceClient.from_env(agent_id)
    transcript: list[dict] = []
    final_action: dict | None = None
    final_receipt: dict | None = None
    error: str | None = None
    trace.plan(f"goal={goal}; initial={initial_observation}", latency_ms=0)
    for i in range(max_iterations):
        reason_prompt = _build_reason_prompt(contract, goal, initial_observation, transcript)
        t0 = _now()
        try:
            raw = llm.complete([
                {"role": "system", "content": reason_prompt},
                {"role": "user", "content": "Decide your next action."},
            ], temperature=0.3)
        except LLMError as e:
            error = f"reason failed: {e}"
            trace.error(error, code=type(e).__name__)
            return ReActResult(i + 1, None, None, transcript, error)
        reason_ms = _now() - t0
        try:
            decision = parse_action_json(raw)
        except (ValueError, LLMError) as e:
            error = f"reason parse: {e}"
            trace.error(error, code="PARSE")
            return ReActResult(i + 1, None, None, transcript, error)
        trace.decision(json.dumps(decision, ensure_ascii=False), latency_ms=reason_ms)
        transcript.append({"step": "reason", "iteration": i, "decision": decision, "latency_ms": reason_ms})

        for call in decision.get("tool_calls", []) or []:
            name = str(call.get("name", ""))
            args = call.get("arguments") or {}
            if not isinstance(args, dict):
                args = {}
            t0 = _now()
            trace.tool_call(name, args, latency_ms=0)
            result = execute_tool(name, args, client=client, sender=agent_id)
            tool_ms = _now() - t0
            trace.tool_result(name, {"ok": result.ok, "result": result.result, "error": result.error},
                             ok=result.ok, latency_ms=tool_ms,
                             error_code="" if result.ok else "TOOL_ERROR")
            transcript.append({"step": "tool", "name": name, "ok": result.ok, "result": result.result, "error": result.error, "latency_ms": tool_ms})

        action = str(decision.get("action", "SKIP")).upper()
        if action == "EXECUTE_TRADE":
            target = str(decision.get("target_agent", ""))
            amount = int(decision.get("amount", 0))
            reasoning = str(decision.get("reasoning", ""))
            try:
                receipt = client.execute_trade(sender=agent_id, target_agent=target, amount=amount, reasoning=reasoning)
                final_receipt = {"tx_id": receipt.tx_id, "to": receipt.to, "amount": receipt.amount}
                transcript.append({"step": "act", "action": "EXECUTE_TRADE", "receipt": final_receipt})
                trace.observation(f"trade settled: {final_receipt}")
            except Exception as e:
                trace.error(f"trade failed: {e}", code="A2A")
                transcript.append({"step": "act", "action": "EXECUTE_TRADE", "error": str(e)})
            final_action = decision
            break
        if action == "POST_FEED":
            content = str(decision.get("content", ""))
            intent = str(decision.get("intent_type", "SOCIAL"))
            try:
                post_id = client.post_feed(sender=agent_id, content=content, intent_type=intent)
                final_action = {**decision, "post_id": post_id}
                transcript.append({"step": "act", "action": "POST_FEED", "post_id": post_id})
                trace.observation(f"feed posted: {post_id}")
            except Exception as e:
                trace.error(f"feed failed: {e}", code="A2A")
                transcript.append({"step": "act", "action": "POST_FEED", "error": str(e)})
            final_action = decision
            break

        if i == max_iterations - 1:
            final_action = decision
            trace.reflection("max iterations reached; stopping")
            break
        reflect_prompt = _build_reflect_prompt(transcript)
        try:
            reflect_raw = llm.complete(
                [{"role": "system", "content": reflect_prompt},
                 {"role": "user", "content": "Return JSON: {continue: bool, reason: str}"}],
                temperature=0.0)
            reflect = parse_action_json(reflect_raw)
        except LLMError as e:
            trace.error(f"reflect failed: {e}", code=type(e).__name__)
            final_action = decision
            break
        if not reflect.get("continue", False):
            final_action = decision
            trace.reflection(f"agent stopped: {reflect.get('reason', '')}")
            break
        trace.reflection(f"agent continued: {reflect.get('reason', '')}")
        initial_observation = "reflection: continue with next step"

    return ReActResult(
        iterations=len(transcript),
        final_action=final_action,
        final_receipt=final_receipt,
        transcript=transcript,
        error=error,
    )
