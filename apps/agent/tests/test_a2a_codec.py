"""A2A codec parity tests — mirrors apps/backend/pkg/a2a/codec_test.go."""

import time

import pytest

from ecomatrix.a2a import (
    Action,
    Code,
    Currency,
    PROTOCOL_V,
    A2AError,
    decode_trade_payload,
    new_msg_id,
    validate_envelope,
    validate_msg_id,
)


def good_envelope() -> dict:
    return {
        "protocol_v": PROTOCOL_V,
        "msg_id": "tx_req_9948",
        "sender": "agent_miner_01",
        "action": Action.EXECUTE_TRADE.value,
        "payload": {
            "target_agent": "agent_merchant_03",
            "offer": {"currency_type": Currency.GOLD.value, "amount": 150},
            "reasoning": "vitality low",
        },
        "timestamp": int(time.time()),
    }


def test_validate_happy_path():
    validate_envelope(good_envelope())


def test_validate_protocol_mismatch():
    env = good_envelope()
    env["protocol_v"] = "1.0"
    with pytest.raises(A2AError) as exc:
        validate_envelope(env)
    assert exc.value.code == Code.PROTOCOL_MISMATCH


def test_validate_missing_msg_id():
    env = good_envelope()
    env["msg_id"] = "x"
    with pytest.raises(A2AError) as exc:
        validate_envelope(env)
    assert exc.value.code == Code.INVALID_ENVELOPE


def test_validate_unknown_action():
    env = good_envelope()
    env["action"] = "DROP_TABLE"
    with pytest.raises(A2AError) as exc:
        validate_envelope(env)
    assert exc.value.code == Code.UNKNOWN_ACTION


def test_validate_clock_skew():
    env = good_envelope()
    env["timestamp"] = int(time.time()) - 7200
    with pytest.raises(A2AError) as exc:
        validate_envelope(env)
    assert exc.value.code == Code.INVALID_ENVELOPE


def test_decode_trade_payload_negative_amount():
    payload = {"target_agent": "agent_merchant_03",
               "offer": {"currency_type": "GOLD", "amount": -10}}
    with pytest.raises(A2AError) as exc:
        decode_trade_payload(payload)
    assert exc.value.code == Code.INVALID_ENVELOPE


def test_decode_trade_payload_bad_currency():
    payload = {"target_agent": "agent_merchant_03",
               "offer": {"currency_type": "USD", "amount": 100}}
    with pytest.raises(A2AError) as exc:
        decode_trade_payload(payload)
    assert exc.value.code == Code.INVALID_ENVELOPE


def test_new_msg_id_format():
    mid = new_msg_id("trade")
    assert validate_msg_id(mid)


def test_decode_feed_payload_happy():
    from ecomatrix.a2a import decode_feed_payload
    p = decode_feed_payload({"content": "selling 10 GOLD of iron", "intent_type": "OFFER"})
    assert p.content == "selling 10 GOLD of iron"
    assert p.intent_type == "OFFER"


def test_decode_feed_payload_missing_content():
    import pytest
    from ecomatrix.a2a import decode_feed_payload, A2AError, Code
    with pytest.raises(A2AError) as exc:
        decode_feed_payload({"intent_type": "OFFER"})
    assert exc.value.code == Code.INVALID_ENVELOPE


def test_decode_feed_payload_bad_intent():
    import pytest
    from ecomatrix.a2a import decode_feed_payload, A2AError, Code
    with pytest.raises(A2AError) as exc:
        decode_feed_payload({"content": "hi", "intent_type": "BULLSHIT"})
    assert exc.value.code == Code.INVALID_ENVELOPE


def test_decode_feed_payload_too_long():
    import pytest
    from ecomatrix.a2a import decode_feed_payload, A2AError, Code
    with pytest.raises(A2AError) as exc:
        decode_feed_payload({"content": "a" * 501, "intent_type": "SOCIAL"})
    assert exc.value.code == Code.INVALID_ENVELOPE


def test_hmac_signing_matches_go_canonical_form():
    """Cross-language check: the Python signer must produce the same digest
    as the Go signer for the same inputs. We don't import the Go code; instead
    we reproduce the Go canonical form by hand and assert round-trip."""
    import hmac as _hmac
    import hashlib as _hashlib
    secret = b"super-secret"
    method = "POST"
    path = "/v1/trades"
    ts = 1713532588
    body = b'{"protocol_v":"1.1","msg_id":"tx_req_9948"}'
    body_hash = _hashlib.sha256(body).hexdigest()
    canonical = "\n".join([method, path, str(ts), body_hash]).encode("utf-8")
    expected = _hmac.new(secret, canonical, _hashlib.sha256).hexdigest()

    from ecomatrix.a2a import _sign
    got = _sign(secret, method, path, ts, body)
    assert got == expected
    assert len(got) == 64  # hex(sha256) is 64 chars


def test_hmac_signing_headers_includes_all_three_when_secret_set(monkeypatch):
    monkeypatch.setenv(
        "ECOMATRIX_AGENT_SECRETS",
        "agent_miner_01=s3cret-a,agent_merchant_01=s3cret-b",
    )
    from ecomatrix.a2a import _signing_headers
    headers = _signing_headers("agent_miner_01", "POST", "/v1/trades", b"{}")
    assert "X-Agent-Id" in headers
    assert "X-Agent-Timestamp" in headers
    assert "X-Agent-Signature" in headers
    assert headers["X-Agent-Id"] == "agent_miner_01"
    assert len(headers["X-Agent-Signature"]) == 64


def test_hmac_signing_headers_empty_when_no_secret(monkeypatch):
    monkeypatch.delenv("ECOMATRIX_AGENT_SECRETS", raising=False)
    from ecomatrix.a2a import _signing_headers
    assert _signing_headers("agent_miner_01", "POST", "/v1/trades", b"{}") == {}


def test_hmac_signing_headers_empty_for_unconfigured_agent(monkeypatch):
    monkeypatch.setenv(
        "ECOMATRIX_AGENT_SECRETS",
        "agent_miner_01=s3cret-a",
    )
    from ecomatrix.a2a import _signing_headers
    assert _signing_headers("agent_merchant_01", "POST", "/v1/trades", b"{}") == {}
