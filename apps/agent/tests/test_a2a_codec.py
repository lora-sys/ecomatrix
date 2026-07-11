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
