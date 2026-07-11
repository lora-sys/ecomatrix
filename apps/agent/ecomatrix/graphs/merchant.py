"""Merchant agent: trade goods for gold; balanced buy/sell behavior."""

from __future__ import annotations

from .base import make_graph, run_one_tick
from ..a2a import A2AClient
from ..llm import LLM

STRATEGY = (
    "You are an EcoMatrix merchant. You accept gold from miners in exchange for goods. "
    "Periodically rebalance by sending 5-15 GOLD to other merchants you trust."
)


def build(*, llm: LLM, client: A2AClient):
    return make_graph("merchant", llm=llm, client=client, strategy_prompt=STRATEGY)


def tick(graph, *, agent_id: str):
    return run_one_tick(graph, agent_id=agent_id, job_type="merchant")
