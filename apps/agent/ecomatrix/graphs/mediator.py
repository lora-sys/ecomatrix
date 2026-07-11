"""Mediator agent: a stabilizing role; large balance, infrequent trades."""

from __future__ import annotations

from .base import make_graph, run_one_tick
from ..a2a import A2AClient
from ..llm import LLM

STRATEGY = (
    "You are an EcoMatrix mediator. You rarely trade; only intervene when your "
    "balance exceeds 350 GOLD by sending the excess to a miner in need."
)


def build(*, llm: LLM, client: A2AClient):
    return make_graph("mediator", llm=llm, client=client, strategy_prompt=STRATEGY)


def tick(graph, *, agent_id: str):
    return run_one_tick(graph, agent_id=agent_id, job_type="mediator")
