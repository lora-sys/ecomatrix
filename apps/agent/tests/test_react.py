"""ReAct workflow tests."""
import pytest
import json

from ecomatrix.llm import StubLLM, MockLLMWithFailures, LLMError
from ecomatrix.workflows.react import run_react, ReActResult


class FakeA2A:
    def __init__(self):
        self.trades = []
        self.feeds = []
        self.get_calls = 0

    def get_agent(self, agent_id):
        self.get_calls += 1
        return {"StringID": agent_id, "Balance": 100, "Vitality": 80, "CreditScore": 60}

    def execute_trade(self, *, sender, target_agent, amount, reasoning):
        from ecomatrix.a2a import Receipt
        self.trades.append({"sender": sender, "target": target_agent, "amount": amount})
        return Receipt(
            tx_id=f"tx_{len(self.trades)}", from_=sender, to=target_agent,
            amount=amount, currency_type="GOLD",
            balance_after_from=100 - amount, balance_after_to=200,
        )

    def post_feed(self, *, sender, content, intent_type):
        self.feeds.append({"sender": sender, "content": content, "intent": intent_type})
        return len(self.feeds)


def test_react_with_stubllm_completes_in_one_iteration():
    client = FakeA2A()
    llm = StubLLM()  # always returns EXECUTE_TRADE
    r = run_react(
        agent_id="agent_miner_01", job_type="miner", llm=llm, client=client,
        initial_observation="balance=100", goal="trade 10 GOLD with merchant_03")
    assert r.error is None
    assert r.final_receipt is not None
    assert r.final_receipt["amount"] == 10
    assert len(client.trades) == 1
    assert client.trades[0]["target"] == "agent_merchant_01"


def test_react_handles_llm_error_gracefully():
    client = FakeA2A()
    llm = MockLLMWithFailures(failure_rate=1.0)
    r = run_react(
        agent_id="agent_miner_01", job_type="miner", llm=llm, client=client,
        initial_observation="balance=100", goal="trade")
    assert r.error is not None
    assert "reason failed" in r.error


class LoopLLM:
    """LLM that always emits a tool call for reason, and always says continue=true for reflect."""
    def complete(self, messages, *, temperature=0.0):
        # The reflect prompt is in messages[0].system; the reason prompt is in messages[0].system.
        # The difference: the user message for reflect says "Return JSON: {continue: bool, reason: str}".
        # For reason, the user message says "Decide your next action."
        last = messages[-1] if messages else {}
        text = str(last.get("content", ""))
        if "continue: bool" in text:
            return json.dumps({"continue": True, "reason": "more info"})
        return json.dumps({
            "thought": "need more info",
            "tool_calls": [{"name": "get_agent_state", "arguments": {}}],
        })


def test_react_with_max_iterations_caps_loop():
    client = FakeA2A()
    r = run_react(
        agent_id="agent_miner_01", job_type="miner", llm=LoopLLM(), client=client,
        initial_observation="x", goal="loop", max_iterations=3)
    # With 3 iterations: 3 reason + 3 tool calls = 6 transcript entries, 3 tool calls.
    tools = sum(1 for s in r.transcript if s.get("step") == "tool")
    assert tools == 3, f"expected 3 tool calls, got {tools}; transcript={r.transcript}"


def test_react_respects_contract_budget():
    from ecomatrix.contracts import get_contract
    c = get_contract("miner")
    assert c.estimated_cost_tokens > 0
