"""Tool registry and executor for the agent.

Each tool is a callable that takes a dict of arguments and returns a JSON-serialisable
result. The agent graph can invoke tools based on the LLM's decision.

Tools:
- ``get_agent_state``: returns the agent's current balance and recent activity.
- ``execute_trade``: submits a trade to the A2A gateway.
- ``post_feed``: publishes a social-feed post.

The tool executor returns a ``ToolResult`` with the result + a status.
"""

from __future__ import annotations

import json
from dataclasses import dataclass
from typing import Any, Callable

from .a2a import A2AClient, A2AError


@dataclass
class ToolResult:
    name: str
    ok: bool
    result: Any
    error: str | None = None

    def to_json(self) -> str:
        d = {"name": self.name, "ok": self.ok, "result": self.result}
        if self.error:
            d["error"] = self.error
        return json.dumps(d, ensure_ascii=False)


def _err(name: str, msg: str) -> ToolResult:
    return ToolResult(name=name, ok=False, result=None, error=msg)


# Public schema — exposed to the LLM in the system prompt so it knows what tools exist.
TOOL_SCHEMA = [
    {
        "name": "get_agent_state",
        "description": "Read the agent's current balance, recent trade count, and social feed activity.",
        "parameters": {"agent_id": "string"},
    },
    {
        "name": "execute_trade",
        "description": "Submit a trade to the A2A gateway. Returns the receipt or an error code.",
        "parameters": {
            "target_agent": "string",
            "amount": "int (positive)",
            "reasoning": "string (optional, <=200 chars)",
        },
    },
    {
        "name": "post_feed",
        "description": "Publish a social-feed post visible to the dashboard and other agents.",
        "parameters": {
            "content": "string (<=500 chars)",
            "intent_type": "OFFER | REQUEST | SOCIAL | META",
        },
    },
]


def execute_tool(name: str, args: dict, *, client: A2AClient, sender: str) -> ToolResult:
    """Dispatch a tool call. Errors are caught and returned as failed ToolResults.

    Never raises — the agent graph must be able to continue after a tool failure.
    """
    try:
        if name == "get_agent_state":
            a = client.get_agent(sender)
            return ToolResult(name=name, ok=True, result={
                "string_id": a.get("StringID", sender),
                "balance": a.get("Balance", 0),
                "vitality": a.get("Vitality", 0),
                "credit_score": a.get("CreditScore", 0),
            })
        if name == "execute_trade":
            target = str(args.get("target_agent", ""))
            amount = int(args.get("amount", 0))
            reasoning = str(args.get("reasoning", ""))
            if amount <= 0:
                return _err(name, "amount must be > 0")
            receipt = client.execute_trade(
                sender=sender, target_agent=target, amount=amount, reasoning=reasoning[:200],
            )
            return ToolResult(name=name, ok=True, result={
                "tx_id": receipt.tx_id,
                "from_balance_after": receipt.balance_after_from,
                "to_balance_after": receipt.balance_after_to,
            })
        if name == "post_feed":
            content = str(args.get("content", ""))
            intent = str(args.get("intent_type", "SOCIAL"))
            if intent not in ("OFFER", "REQUEST", "SOCIAL", "META"):
                return _err(name, f"intent_type {intent!r} not in OFFER|REQUEST|SOCIAL|META")
            if not content or len(content) > 500:
                return _err(name, "content must be 1..500 chars")
            post_id = client.post_feed(sender=sender, content=content, intent_type=intent)
            return ToolResult(name=name, ok=True, result={"post_id": post_id})
        return _err(name, f"unknown tool: {name}")
    except A2AError as e:
        return _err(name, f"a2a {e.code.value}: {e.message}")
    except Exception as e:
        return _err(name, f"unexpected: {e!r}")


def execute_tools_in_sequence(
    calls: list[dict], *, client: A2AClient, sender: str,
) -> list[ToolResult]:
    """Run a list of tool calls in order. Each must be {name, arguments}."""
    out: list[ToolResult] = []
    for c in calls:
        name = str(c.get("name", ""))
        args = c.get("arguments") or {}
        if not isinstance(args, dict):
            args = {}
        out.append(execute_tool(name, args, client=client, sender=sender))
    return out
