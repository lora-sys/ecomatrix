"""LLM provider abstraction with error types, retry, and mock-with-failures.

Default providers:
- ``stub``: deterministic in-process echo with strategy-like structure.
- ``mock_failures``: deterministic JSON output but a configurable failure
  rate. Used by the failure-mode test suite to verify graceful degradation.
- ``openai``: OpenAI-compatible chat completion with JSON-mode forcing,
  response validation, and timeout-aware retries.

Any provider must satisfy the ``LLM`` protocol so the LangGraph nodes
remain provider-agnostic.
"""

from __future__ import annotations

import json
import os
import random
import re
import time
from dataclasses import dataclass, field
from typing import Any, Protocol

_SENDER_RE = re.compile(r"agent_id=([A-Za-z0-9_]+)")
_TARGET_RE = re.compile(r"target=([A-Za-z0-9_]+)")
_AMOUNT_RE = re.compile(r"amount=(\d+)")
_HASH_RE = re.compile(r"hash=([0-9a-f]+)")


# --- Error hierarchy ----------------------------------------------------------


class LLMError(Exception):
    """Base class for all LLM provider errors.

    Subclasses carry a ``retryable`` flag so the agent can decide whether
    to retry, fall back, or surface the error to the user.
    """

    retryable: bool = False

    def __init__(self, message: str, *, retryable: bool | None = None) -> None:
        super().__init__(message)
        if retryable is not None:
            self.retryable = retryable


class LLMTimeoutError(LLMError):
    retryable = True


class LLMRateLimitError(LLMError):
    retryable = True


class LLMProviderError(LLMError):
    """Upstream provider returned a non-2xx or transport error."""

    retryable = True


class LLMMalformedResponseError(LLMError):
    """The response wasn't parseable JSON / didn't match the action schema."""

    retryable = False


class LLMRefusalError(LLMError):
    """The model refused the request (safety / policy)."""

    retryable = False


# --- Protocol ------------------------------------------------------------------


class LLM(Protocol):
    def complete(self, messages: list[dict[str, str]], *, temperature: float = 0.4) -> str: ...


# --- Stub (deterministic) ------------------------------------------------------


@dataclass
class StubLLM:
    """Deterministic LLM. Returns a structured JSON action.

    Picks a target from the prompt and an amount based on the sender role.
    Use for tests + offline demos.
    """

    def complete(self, messages: list[dict[str, str]], *, temperature: float = 0.4) -> str:
        last_user = next(
            (m["content"] for m in reversed(messages) if m.get("role") == "user"),
            "",
        )
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


# --- Mock with failures (for failure-mode tests) -------------------------------


@dataclass
class MockLLMWithFailures:
    """Like StubLLM but raises LLMError with a configurable probability.

    Use ``ECOMATRIX_AGENT_LLM_FAILURE_RATE=0.5`` to flip half the calls.
    """

    failure_rate: float = 0.0
    rng: random.Random = field(default_factory=random.Random)
    call_count: int = 0
    failure_kind: type[LLMError] = LLMRateLimitError
    base: StubLLM = field(default_factory=StubLLM)

    def __post_init__(self) -> None:
        if "ECOMATRIX_AGENT_LLM_FAILURE_RATE" in os.environ:
            self.failure_rate = float(os.environ["ECOMATRIX_AGENT_LLM_FAILURE_RATE"])
        if "ECOMATRIX_AGENT_LLM_FAILURE_KIND" in os.environ:
            kind = os.environ["ECOMATRIX_AGENT_LLM_FAILURE_KIND"]
            self.failure_kind = {
                "timeout": LLMTimeoutError,
                "rate_limit": LLMRateLimitError,
                "provider": LLMProviderError,
                "malformed": LLMMalformedResponseError,
                "refusal": LLMRefusalError,
            }.get(kind, LLMRateLimitError)

    def complete(self, messages: list[dict[str, str]], *, temperature: float = 0.4) -> str:
        self.call_count += 1
        if self.rng.random() < self.failure_rate:
            kind = self.failure_kind
            msg = f"mock failure #{self.call_count} ({kind.__name__})"
            raise kind(msg)
        return self.base.complete(messages, temperature=temperature)


# --- OpenAI-compatible provider ----------------------------------------------


@dataclass
class OpenAICompatibleLLM:
    """OpenAI-compatible chat completion client with retries.

    Reads from env: ``ECOMATRIX_AGENT_OPENAI_API_KEY``,
    ``ECOMATRIX_AGENT_OPENAI_BASE_URL``, ``ECOMATRIX_AGENT_LLM_MODEL``.
    Forces JSON mode (response_format=json_object).
    """

    api_key: str = ""
    base_url: str = "https://api.openai.com/v1"
    model: str = "gpt-4o-mini"
    timeout_seconds: float = 15.0
    max_retries: int = 2
    retry_backoff_seconds: float = 0.5

    def __post_init__(self) -> None:
        if not self.api_key:
            self.api_key = os.environ.get("ECOMATRIX_AGENT_OPENAI_API_KEY", "")
        if "ECOMATRIX_AGENT_OPENAI_BASE_URL" in os.environ:
            self.base_url = os.environ["ECOMATRIX_AGENT_OPENAI_BASE_URL"]
        if "ECOMATRIX_AGENT_LLM_MODEL" in os.environ:
            self.model = os.environ["ECOMATRIX_AGENT_LLM_MODEL"]

    def complete(self, messages: list[dict[str, str]], *, temperature: float = 0.4) -> str:
        if not self.api_key:
            raise LLMProviderError(
                "ECOMATRIX_AGENT_OPENAI_API_KEY is not set; cannot call LLM",
                retryable=False,
            )
        import httpx
        last_err: Exception | None = None
        for attempt in range(self.max_retries + 1):
            try:
                r = httpx.post(
                    f"{self.base_url.rstrip('/')}/chat/completions",
                    headers={"Authorization": f"Bearer {self.api_key}"},
                    json={
                        "model": self.model,
                        "messages": messages,
                        "temperature": temperature,
                        "response_format": {"type": "json_object"},
                    },
                    timeout=self.timeout_seconds,
                )
                if r.status_code == 429:
                    raise LLMRateLimitError(f"upstream 429: {r.text[:200]}")
                if r.status_code >= 500:
                    raise LLMProviderError(f"upstream {r.status_code}: {r.text[:200]}")
                if r.status_code >= 400:
                    raise LLMProviderError(
                        f"upstream {r.status_code}: {r.text[:200]}",
                        retryable=False,
                    )
                body = r.json()
                content = body["choices"][0]["message"]["content"]
                if not content or not content.strip():
                    raise LLMMalformedResponseError("upstream returned empty content")
                return content
            except httpx.TimeoutException as e:
                last_err = LLMTimeoutError(str(e)[:200])
            except httpx.HTTPError as e:
                last_err = LLMProviderError(str(e)[:200])
            except (LLMRateLimitError, LLMProviderError) as e:
                last_err = e
                if not e.retryable:
                    raise
            if attempt < self.max_retries:
                time.sleep(self.retry_backoff_seconds * (2 ** attempt))
        raise last_err or LLMProviderError("LLM call failed after retries", retryable=False)


# --- JSON parsing + retry (shared) ------------------------------------------


def parse_action_json(raw: str) -> dict:
    """Tolerate LLMs that wrap JSON in markdown fences or add preamble.

    Raises LLMMalformedResponseError if no JSON object is found.
    """
    if not raw:
        raise LLMMalformedResponseError("empty LLM response")
    m = re.search(r"\{.*\}", raw, flags=re.DOTALL)
    if not m:
        raise LLMMalformedResponseError(f"no JSON in LLM response: {raw[:200]}")
    try:
        return json.loads(m.group(0))
    except json.JSONDecodeError as e:
        raise LLMMalformedResponseError(f"invalid JSON: {e}; raw={raw[:200]}")


# --- Factory -----------------------------------------------------------------


def get_default_llm() -> LLM:
    provider = os.environ.get("ECOMATRIX_AGENT_LLM_PROVIDER", "stub").lower()
    if provider == "stub":
        return StubLLM()
    if provider == "mock_failures":
        return MockLLMWithFailures()
    if provider in {"openai", "openai-compatible"}:
        return OpenAICompatibleLLM()
    raise ValueError(f"unknown LLM provider: {provider}")
