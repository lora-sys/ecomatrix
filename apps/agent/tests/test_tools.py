"""Tool registry tests."""
import json
from ecomatrix.tools import TOOL_SCHEMA, execute_tool, execute_tools_in_sequence, ToolResult


class FakeClient:
    def __init__(self):
        self.calls = []

    def get_agent(self, agent_id):
        return {"StringID": agent_id, "Balance": 100, "Vitality": 50, "CreditScore": 70}

    def execute_trade(self, *, sender, target_agent, amount, reasoning):
        self.calls.append(("execute_trade", sender, target_agent, amount, reasoning))
        from ecomatrix.a2a import Receipt
        return Receipt(
            tx_id="tx_test", from_=sender, to=target_agent,
            amount=amount, currency_type="GOLD",
            balance_after_from=100 - amount, balance_after_to=200,
        )

    def post_feed(self, *, sender, content, intent_type):
        self.calls.append(("post_feed", sender, content, intent_type))
        return 42


def test_schema_exposes_three_tools():
    names = {t["name"] for t in TOOL_SCHEMA}
    assert names == {"get_agent_state", "execute_trade", "post_feed"}


def test_get_agent_state():
    c = FakeClient()
    r = execute_tool("get_agent_state", {}, client=c, sender="agent_miner_01")
    assert r.ok
    assert r.result["balance"] == 100
    assert r.result["vitality"] == 50


def test_execute_trade_happy_path():
    c = FakeClient()
    r = execute_tool("execute_trade", {"target_agent": "agent_merchant_03", "amount": 7, "reasoning": "test"},
                    client=c, sender="agent_miner_01")
    assert r.ok
    assert r.result["tx_id"] == "tx_test"
    assert r.result["from_balance_after"] == 93
    assert c.calls[0][0] == "execute_trade"


def test_execute_trade_invalid_amount():
    c = FakeClient()
    r = execute_tool("execute_trade", {"target_agent": "agent_x", "amount": 0}, client=c, sender="a")
    assert not r.ok
    assert "amount" in r.error


def test_post_feed_happy():
    c = FakeClient()
    r = execute_tool("post_feed", {"content": "hello", "intent_type": "SOCIAL"},
                    client=c, sender="agent_miner_01")
    assert r.ok
    assert r.result["post_id"] == 42


def test_post_feed_bad_intent():
    c = FakeClient()
    r = execute_tool("post_feed", {"content": "hi", "intent_type": "ROAR"},
                    client=c, sender="agent_miner_01")
    assert not r.ok
    assert "intent_type" in r.error


def test_post_feed_too_long():
    c = FakeClient()
    r = execute_tool("post_feed", {"content": "x" * 501, "intent_type": "SOCIAL"},
                    client=c, sender="agent_miner_01")
    assert not r.ok


def test_unknown_tool_returns_error():
    c = FakeClient()
    r = execute_tool("does_not_exist", {}, client=c, sender="a")
    assert not r.ok
    assert "unknown tool" in r.error


def test_execute_tools_in_sequence():
    c = FakeClient()
    out = execute_tools_in_sequence([
        {"name": "get_agent_state", "arguments": {}},
        {"name": "execute_trade", "arguments": {"target_agent": "agent_x", "amount": 5}},
        {"name": "post_feed", "arguments": {"content": "hi", "intent_type": "SOCIAL"}},
        {"name": "bad_tool", "arguments": {}},
    ], client=c, sender="agent_miner_01")
    assert len(out) == 4
    assert out[0].ok
    assert out[1].ok
    assert out[2].ok
    assert not out[3].ok
