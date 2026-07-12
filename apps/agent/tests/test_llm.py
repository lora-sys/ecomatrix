"""LLM provider tests: stub, mock-with-failures, error hierarchy, parse_action_json."""
import json
import os
import pytest

from ecomatrix.llm import (
    StubLLM,
    MockLLMWithFailures,
    OpenAICompatibleLLM,
    parse_action_json,
    LLMTimeoutError,
    LLMRateLimitError,
    LLMProviderError,
    LLMMalformedResponseError,
    LLMRefusalError,
    LLMError,
)


# --- Stub ----------------------------------------------------------------------


def test_stub_returns_valid_json():
    raw = StubLLM().complete([
        {"role": "user", "content": "target=agent_merchant_01 amount=10"}
    ])
    obj = parse_action_json(raw)
    assert obj["action"] == "EXECUTE_TRADE"
    assert obj["target_agent"] == "agent_merchant_01"
    assert obj["amount"] == 10


def test_stub_avoids_self_trade():
    raw = StubLLM().complete([
        {"role": "user", "content": "agent_id=agent_merchant_01 target=agent_merchant_01 amount=5"}
    ])
    obj = parse_action_json(raw)
    assert obj["target_agent"] != "agent_merchant_01"


# --- Mock with failures ----------------------------------------------------------


def test_mock_failures_default_no_failures():
    m = MockLLMWithFailures(failure_rate=0.0)
    for _ in range(10):
        raw = m.complete([{"role": "user", "content": "x"}])
        parse_action_json(raw)
    assert m.call_count == 10


def test_mock_failures_always_fail():
    m = MockLLMWithFailures(failure_rate=1.0)
    for _ in range(5):
        with pytest.raises(LLMError):
            m.complete([{"role": "user", "content": "x"}])


def test_mock_failures_specific_kind(monkeypatch):
    monkeypatch.setenv("ECOMATRIX_AGENT_LLM_FAILURE_KIND", "timeout")
    m = MockLLMWithFailures(failure_rate=1.0)
    with pytest.raises(LLMTimeoutError):
        m.complete([{"role": "user", "content": "x"}])


def test_mock_failures_probabilistic_distribution():
    """Over many trials, the failure rate should approximate the configured rate."""
    import random
    m = MockLLMWithFailures(failure_rate=0.5, rng=random.Random(42))
    fail = 0
    n = 200
    for _ in range(n):
        try:
            m.complete([{"role": "user", "content": "x"}])
        except LLMError:
            fail += 1
    # Should be within 20% of 50%.
    assert 0.30 <= fail / n <= 0.70, f"failures={fail}/{n}"


# --- OpenAICompatibleLLM --------------------------------------------------------


def test_openai_missing_api_key_raises_provider_error():
    m = OpenAICompatibleLLM(api_key="")
    with pytest.raises(LLMProviderError) as exc:
        m.complete([{"role": "user", "content": "x"}])
    assert not exc.value.retryable  # not retryable — env misconfig


def test_openai_retry_on_timeout(monkeypatch):
    """Two timeouts then a success: should retry with backoff and return."""
    import httpx
    from unittest.mock import patch, MagicMock

    class FakeResp:
        def __init__(self, payload): self._p = payload
        def json(self): return self._p
        status_code = 200

    calls = {"n": 0}

    def fake_post(url, headers=None, json=None, timeout=None):
        calls["n"] += 1
        if calls["n"] <= 2:
            raise httpx.TimeoutException("boom")
        return FakeResp({
            "choices": [{"message": {"content": '{"action":"EXECUTE_TRADE","target_agent":"agent_x","amount":3,"reasoning":"r"}'}}]
        })

    m = OpenAICompatibleLLM(
        api_key="k", max_retries=2, retry_backoff_seconds=0.0, timeout_seconds=1.0,
    )
    with patch("httpx.post", side_effect=fake_post):
        raw = m.complete([{"role": "user", "content": "x"}])
    obj = parse_action_json(raw)
    assert obj["action"] == "EXECUTE_TRADE"
    assert calls["n"] == 3  # 2 failures + 1 success


def test_openai_non_retryable_4xx(monkeypatch):
    from unittest.mock import patch
    class FakeResp:
        status_code = 401
        text = "bad key"
        def json(self): return {"error": "bad"}
    m = OpenAICompatibleLLM(api_key="k", max_retries=2, retry_backoff_seconds=0.0)
    with patch("httpx.post", return_value=FakeResp()):
        with pytest.raises(LLMProviderError) as exc:
            m.complete([{"role": "user", "content": "x"}])
    assert not exc.value.retryable


# --- parse_action_json ---------------------------------------------------------


def test_parse_action_json_extracts_from_fence():
    raw = "```json\n{\"action\":\"X\",\"target_agent\":\"a\",\"amount\":1,\"reasoning\":\"r\"}\n```"
    obj = parse_action_json(raw)
    assert obj["action"] == "X"


def test_parse_action_json_no_json_raises():
    with pytest.raises(LLMMalformedResponseError):
        parse_action_json("just text, no braces")


def test_parse_action_json_empty_raises():
    with pytest.raises(LLMMalformedResponseError):
        parse_action_json("")


def test_parse_action_json_invalid_json_raises():
    with pytest.raises(LLMMalformedResponseError):
        parse_action_json("{not valid json}")


# --- Error hierarchy ---------------------------------------------------------


def test_error_hierarchy():
    for cls in (LLMTimeoutError, LLMRateLimitError, LLMProviderError, LLMMalformedResponseError, LLMRefusalError):
        assert issubclass(cls, LLMError)
        assert hasattr(cls, "retryable")
    assert LLMTimeoutError.retryable is True
