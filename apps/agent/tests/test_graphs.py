"""Per-job graph tests against a mocked A2A backend."""
from dataclasses import dataclass, field
from typing import Any

from ecomatrix.graphs import hacker, mediator, merchant, miner
from ecomatrix.graphs.base import run_one_tick
from ecomatrix.llm import StubLLM


@dataclass
class FakeReceipt:
    tx_id: str = "tx_fake"
    from_: str = ""
    to: str = ""
    amount: int = 0
    currency_type: str = "GOLD"
    balance_after_from: int = 0
    balance_after_to: int = 0


@dataclass
class FakeClient:
    agents: dict[str, dict[str, Any]] = field(default_factory=lambda: {
        "agent_miner_01": {"Balance": 200},
        "agent_merchant_01": {"Balance": 300},
        "agent_merchant_02": {"Balance": 250},
        "agent_hacker_01": {"Balance": 100},
        "agent_mediator_01": {"Balance": 400},
    })
    trades: list[dict[str, Any]] = field(default_factory=list)
    feeds: list[dict[str, Any]] = field(default_factory=list)

    def get_agent(self, agent_id):
        return self.agents[agent_id]

    def execute_trade(self, sender, target_agent, amount, reasoning="", msg_id=None):
        self.trades.append({
            "sender": sender, "target": target_agent, "amount": amount,
            "reasoning": reasoning,
        })
        self.agents[sender]["Balance"] -= amount
        self.agents[target_agent]["Balance"] += amount
        return FakeReceipt(from_=sender, to=target_agent, amount=amount)

    def post_feed(self, sender, content, intent_type="SOCIAL", msg_id=None):
        self.feeds.append({"sender": sender, "content": content, "intent_type": intent_type})
        return len(self.feeds)

    def list_feeds(self, limit=50):
        return list(self.feeds)[-limit:]


def test_miner_graph_emits_trade():
    c = FakeClient()
    g = miner.build(llm=StubLLM(), client=c)
    result = miner.tick(g, agent_id="agent_miner_01")
    assert result.receipt is not None
    assert result.error is None
    assert c.trades and c.trades[0]["sender"] == "agent_miner_01"


def test_merchant_graph_emits_trade():
    c = FakeClient()
    g = merchant.build(llm=StubLLM(), client=c)
    result = merchant.tick(g, agent_id="agent_merchant_01")
    assert result.receipt is not None
    assert c.trades[0]["sender"] == "agent_merchant_01"


def test_all_four_job_types_run():
    c = FakeClient()
    llm = StubLLM()
    cases = [
        (miner.build, "agent_miner_01", "miner"),
        (merchant.build, "agent_merchant_01", "merchant"),
        (hacker.build, "agent_hacker_01", "hacker"),
        (mediator.build, "agent_mediator_01", "mediator"),
    ]
    for builder, agent_id, job_type in cases:
        g = builder(llm=llm, client=c)
        r = run_one_tick(g, agent_id=agent_id, job_type=job_type)
        assert r is not None
