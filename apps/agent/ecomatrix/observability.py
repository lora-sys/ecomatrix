"""Observability — client for writing traces to the EcoMatrix backend.

The trace is the unit of observability: every decision, tool call, and
error is logged with latency + (optional) token counts + cost.
"""

from __future__ import annotations

import os
import time
import urllib.request
import urllib.error
from dataclasses import dataclass, field
from typing import Any


@dataclass
class TraceClient:
    base_url: str = "http://localhost:8080"
    agent_id: str = ""
    enabled: bool = True

    @classmethod
    def from_env(cls, agent_id: str) -> "TraceClient":
        base = os.environ.get("ECOMATRIX_AGENT_BACKEND_URL", "http://localhost:8080")
        enabled = os.environ.get("ECOMATRIX_AGENT_TRACES", "1") != "0"
        return cls(base_url=base, agent_id=agent_id, enabled=enabled)

    def _post(self, path: str, body: dict) -> None:
        if not self.enabled:
            return
        try:
            req = urllib.request.Request(
                f"{self.base_url}{path}",
                data=__import__("json").dumps(body).encode(),
                headers={"Content-Type": "application/json"},
                method="POST",
            )
            urllib.request.urlopen(req, timeout=2).read()
        except Exception:
            # Never let a tracing failure crash the agent.
            pass

    def plan(self, plan_text: str, latency_ms: int = 0) -> None:
        self._post("/v1/traces", {
            "agent_id": self.agent_id,
            "kind": "plan",
            "content": plan_text,
            "latency_ms": latency_ms,
        })

    def decision(self, decision_json: str, latency_ms: int = 0,
                 tokens_in: int = 0, tokens_out: int = 0) -> None:
        self._post("/v1/traces", {
            "agent_id": self.agent_id,
            "kind": "decision",
            "content": decision_json,
            "latency_ms": latency_ms,
            "tokens_in": tokens_in,
            "tokens_out": tokens_out,
        })

    def tool_call(self, name: str, args: dict, latency_ms: int = 0,
                  parent_id: int | None = None) -> None:
        import json as _json
        self._post("/v1/traces", {
            "agent_id": self.agent_id,
            "kind": "tool_call",
            "content": f"{name}({_json.dumps(args, ensure_ascii=False)})",
            "latency_ms": latency_ms,
            "tool_name": name,
            "tool_input": _json.dumps(args).encode(),
            "parent_id": parent_id,
        })

    def tool_result(self, name: str, result: dict, ok: bool, latency_ms: int = 0,
                    error_code: str = "", parent_id: int | None = None) -> None:
        import json as _json
        self._post("/v1/traces", {
            "agent_id": self.agent_id,
            "kind": "tool_result",
            "content": f"{name} ok={ok}",
            "latency_ms": latency_ms,
            "tool_name": name,
            "tool_output": _json.dumps(result).encode(),
            "error_code": error_code,
            "parent_id": parent_id,
        })

    def observation(self, text: str) -> None:
        self._post("/v1/traces", {
            "agent_id": self.agent_id,
            "kind": "observation",
            "content": text,
        })

    def error(self, message: str, code: str = "") -> None:
        self._post("/v1/traces", {
            "agent_id": self.agent_id,
            "kind": "error",
            "content": message,
            "error_code": code,
        })

    def reflection(self, text: str) -> None:
        self._post("/v1/traces", {
            "agent_id": self.agent_id,
            "kind": "reflection",
            "content": text,
        })
