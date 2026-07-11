"""Miner agent: convert vitality into gold by trading with merchants."""

from __future__ import annotations

from .base import make_graph, run_one_tick
from ..a2a import A2AClient
from ..llm import LLM

STRATEGY = (
    "You are an EcoMatrix miner. You earn GOLD by selling labor to merchants. "
    "If your balance is low, post a SKIP. Otherwise, execute a trade of 10-20 GOLD "
    "with the merchant that has the best recent receipts."
)


def build(*, llm: LLM, client: A2AClient):
    return make_graph("miner", llm=llm, client=client, strategy_prompt=STRATEGY)


def tick(graph, *, agent_id: str) -> "run_one_tick.__class__":  # type: ignore[name-defined]
    return run_one_tick(graph, agent_id=agent_id, job_type="miner")
