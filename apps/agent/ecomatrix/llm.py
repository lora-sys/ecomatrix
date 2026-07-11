"""LLM provider abstraction.

Default providers:
- ``stub``: deterministic in-process echo with strategy-like structure.
  Use for tests and for running the demo without network access.
- ``openai``: OpenAI-compatible chat completion. Configure via env vars.

Any provider must satisfy the simple ``complete(messages) -> str`` interface
so the LangGraph nodes remain provider-agnostic.
"""

from __future__ import annotations

import json
import os
import re
from dataclasses import dataclass
from typing import Protocol

_SENDER_RE = re.compile(r"agent_id=([A-Za-z0-9_]+)")
_TARGET_RE = re.compile(r"target=([A-Za-z0-9_]+)")
_AMOUNT_RE = re.compile(r"amount=(\d+)")


class LLM(Protocol):
    def complete(self, messages: list[dict[str, str]], *, temperature: float = 0.4) -> str: ...


@dataclass
class StubLLM:
    """Deterministic LLM. Useful for tests and offline demos.

    Picks a structured response based on the most recent user message.
    """

    def complete(self, messages: list[dict[str, str]], *, temperature: float = 0.4) -> str:
        last_user = next((m["content"] for m in reversed(messages) if m.get("role") == "user"), "")
        target_m = _TARGET_RE.search(last_user)
        amount_m = _AMOUNT_RE.search(last_user)
        sender_m = _SENDER_RE.search(last_user)
        target = target_m.group(1) if target_m else "agent_merchant_01"
        amount = int(amount_m.group(1)) if amount_m else 10
        sender = sender_m.group(1) if sender_m else ""
        if target == sender:
            target = "agent_merchant_02" if sender.startswith("agent_merchant") else "agent_merchant_01"
        return json.dumps({
            "thought": "stub: deterministic trade",
            "action": "EXECUTE_TRADE",
            "target_agent": target,
            "amount": amount,
            "reasoning": "stub provider",
        })


@dataclass
class OpenAICompatibleLLM:
    """OpenAI-compatible chat completion client.

    Reads ``ECOMATRIX_AGENT_OPENAI_API_KEY`` and ``ECOMATRIX_AGENT_OPENAI_BASE_URL``.
    Uses the standard ``/chat/completions`` endpoint.
    """

    api_key: str
    base_url: str = "https://api.openai.com/v1"
    model: str = "gpt-4o-mini"

    def complete(self, messages: list[dict[str, str]], *, temperature: float = 0.4) -> str:
        import httpx
        if not self.api_key:
            raise RuntimeError("ECOMATRIX_AGENT_OPENAI_API_KEY is not set")
        r = httpx.post(
            f"{self.base_url.rstrip('/')}/chat/completions",
            headers={"Authorization": f"Bearer {self.api_key}"},
            json={"model": self.model, "messages": messages, "temperature": temperature},
            timeout=30.0,
        )
        r.raise_for_status()
        return str(r.json()["choices"][0]["message"]["content"])


def get_default_llm() -> LLM:
    provider = os.environ.get("ECOMATRIX_AGENT_LLM_PROVIDER", "stub").lower()
    if provider == "stub":
        return StubLLM()
    if provider in {"openai", "openai-compatible"}:
        return OpenAICompatibleLLM(
            api_key=os.environ.get("ECOMATRIX_AGENT_OPENAI_API_KEY", ""),
            base_url=os.environ.get("ECOMATRIX_AGENT_OPENAI_BASE_URL", "https://api.openai.com/v1"),
            model=os.environ.get("ECOMATRIX_AGENT_LLM_MODEL", "gpt-4o-mini"),
        )
    raise ValueError(f"unknown LLM provider: {provider}")
