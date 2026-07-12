"""Agent Contract — formal schema for every agent in the system.

Per the design spec:
- Input:  what the agent accepts.
- Output: what the agent produces.
- Capability: which tools it can use.
- Limitation: what it must NOT do.
- Example: a concrete input/output pair.
- Cost:   estimated tokens per tick.

The contract is the unit of governance:
- Tests assert the agent honours its contract.
- The dashboard shows the contract so operators understand what the agent does.
- Future LLM-as-judge eval can grade outputs against the contract.
"""
from __future__ import annotations

import json
from dataclasses import dataclass, field, asdict
from typing import Any


@dataclass
class AgentContract:
    name: str                           # e.g. "miner"
    job_type: str                       # matches domain.JobType
    goal: str                           # one-sentence mission
    inputs: list[str] = field(default_factory=list)        # accepted input shapes (described)
    outputs: list[str] = field(default_factory=list)       # produced output shapes (described)
    capabilities: list[str] = field(default_factory=list)  # which tools the agent may call
    limitations: list[str] = field(default_factory=list)   # what the agent must NOT do
    example_input: dict[str, Any] = field(default_factory=dict)
    example_output: dict[str, Any] = field(default_factory=dict)
    estimated_cost_tokens: int = 800   # single-tick budget for prompt + completion

    def to_dict(self) -> dict:
        return asdict(self)

    def to_json(self) -> str:
        return json.dumps(self.to_dict(), ensure_ascii=False, indent=2)

    def system_prompt_section(self) -> str:
        """Render the contract as a section of the system prompt.

        Used by every workflow to enforce the contract.
        """
        lines = [
            f"## Agent Contract — {self.name}",
            f"Goal: {self.goal}",
            "",
            "### You ACCEPT the following input shapes:",
            *[f"- {i}" for i in self.inputs],
            "",
            "### You MUST PRODUCE one of the following output shapes:",
            *[f"- {o}" for o in self.outputs],
            "",
            "### You MAY call the following tools:",
            *[f"- {c}" for c in self.capabilities],
            "",
            "### You MUST NOT:",
            *[f"- {l}" for l in self.limitations],
            "",
            f"### Cost budget: ~{self.estimated_cost_tokens} tokens per tick. Stay under it.",
        ]
        return "\n".join(lines)


# --- Concrete contracts per job type -----------------------------------------


MINER_CONTRACT = AgentContract(
    name="miner",
    job_type="miner",
    goal="Convert labour into GOLD by trading with merchants at fair rates.",
    inputs=[
        "agent_id (string, default agent_miner_*)",
        "current balance (int GOLD)",
        "vitality (int 0..100, drives urgency)",
    ],
    outputs=[
        "EXECUTE_TRADE action with target_agent, amount, reasoning",
        "POST_FEED action with content and intent (REQUEST preferred when low on funds)",
        "SKIP action when no profitable move is visible",
    ],
    capabilities=[
        "get_agent_state: read own balance, vitality, and recent activity",
        "execute_trade: post a trade to the A2A gateway",
        "post_feed: publish a status to the social square",
    ],
    limitations=[
        "NEVER trade with yourself",
        "NEVER exceed your current balance",
        "NEVER exceed the per-tick token budget (no infinite reasoning loops)",
        "NEVER use a tool you don't need (skip empty tool_calls when SKIP is the right answer)",
    ],
    example_input={"agent_id": "agent_miner_01", "balance": 100, "vitality": 80},
    example_output={
        "thought": "balance 100, want to trade with merchant_03",
        "action": "EXECUTE_TRADE",
        "target_agent": "agent_merchant_03",
        "amount": 7,
        "reasoning": "low vitality — buy food",
    },
    estimated_cost_tokens=600,
)


MERCHANT_CONTRACT = AgentContract(
    name="merchant",
    job_type="merchant",
    goal="Accept trades that turn a profit and balance the book across counterparties.",
    inputs=[
        "agent_id (string, default agent_merchant_*)",
        "current balance (int GOLD)",
        "vitality (int)",
    ],
    outputs=[
        "EXECUTE_TRADE: rebalance to another merchant or buy from miner",
        "POST_FEED: post an OFFER when holding inventory",
    ],
    capabilities=[
        "get_agent_state",
        "execute_trade",
        "post_feed",
    ],
    limitations=[
        "NEVER trade with yourself",
        "NEVER trade below your cost basis without a reason",
    ],
    example_input={"agent_id": "agent_merchant_01", "balance": 200},
    example_output={"action": "EXECUTE_TRADE", "target_agent": "agent_merchant_02", "amount": 5, "reasoning": "rebalance"},
    estimated_cost_tokens=600,
)


HACKER_CONTRACT = AgentContract(
    name="hacker",
    job_type="hacker",
    goal="Probe counterparties for asymmetric opportunities (1-5 GOLD probes).",
    inputs=["agent_id", "balance", "vitality"],
    outputs=[
        "Small (1-5 GOLD) probe trades",
        "POST_FEED: SOCIAL post about findings",
    ],
    capabilities=["get_agent_state", "execute_trade", "post_feed"],
    limitations=[
        "NEVER trade with yourself",
        "NEVER exceed 5 GOLD on a single probe (too risky)",
        "If 3 consecutive probes are REJECTED, SKIP for 2 ticks",
    ],
    example_input={"agent_id": "agent_hacker_01", "balance": 80},
    example_output={"action": "EXECUTE_TRADE", "target_agent": "agent_merchant_01", "amount": 1},
    estimated_cost_tokens=500,
)


MEDIATOR_CONTRACT = AgentContract(
    name="mediator",
    job_type="mediator",
    goal="Stabilize the system: intervene only when balance > 350 by redistributing to a miner in need.",
    inputs=["agent_id", "balance", "vitality"],
    outputs=[
        "Single transfer: balance - excess, target = a miner",
        "POST_FEED: META post about the intervention",
    ],
    capabilities=["get_agent_state", "execute_trade", "post_feed"],
    limitations=[
        "NEVER trade with yourself",
        "NEVER act if balance <= 350 (stay below the threshold)",
        "NEVER execute more than 1 transfer per tick",
    ],
    example_input={"agent_id": "agent_mediator_01", "balance": 400},
    example_output={"action": "EXECUTE_TRADE", "target_agent": "agent_miner_01", "amount": 50, "reasoning": "stabilize"},
    estimated_cost_tokens=400,
)


# --- Registry ----------------------------------------------------------------


CONTRACTS: dict[str, AgentContract] = {
    c.job_type: c for c in (MINER_CONTRACT, MERCHANT_CONTRACT, HACKER_CONTRACT, MEDIATOR_CONTRACT)
}


def get_contract(job_type: str) -> AgentContract:
    if job_type not in CONTRACTS:
        raise KeyError(f"unknown job_type {job_type!r}; known: {list(CONTRACTS)}")
    return CONTRACTS[job_type]
