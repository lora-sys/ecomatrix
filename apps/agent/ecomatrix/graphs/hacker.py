"""Hacker agent: opportunistic; attempts many small trades."""

from __future__ import annotations

from .base import make_graph, run_one_tick
from ..a2a import A2AClient
from ..llm import LLM

STRATEGY = (
    "You are an EcoMatrix hacker. You probe for opportunities with 1-5 GOLD trades. "
    "If three consecutive trades are rejected, switch to SKIP for two ticks."
)


def build(*, llm: LLM, client: A2AClient):
    return make_graph("hacker", llm=llm, client=client, strategy_prompt=STRATEGY)


def tick(graph, *, agent_id: str):
    return run_one_tick(graph, agent_id=agent_id, job_type="hacker")
