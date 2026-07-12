"""Cost control + token budget.

Per the design spec: production agents must respect token budgets.
This module:
- Tracks per-tick budget against the contract's estimated_cost_tokens.
- Tracks cumulative cost across ticks.
- Rejects LLM calls when the budget is exhausted.
"""
from __future__ import annotations

import time
from dataclasses import dataclass, field


@dataclass
class CostLedger:
    budget_per_tick: int = 1000
    budget_cumulative: int = 100_000  # 100k tokens
    cumulative_used: int = 0
    tick_used: int = 0
    tick_start: float = field(default_factory=time.time)

    def reset_tick(self) -> None:
        self.tick_used = 0
        self.tick_start = time.time()

    def can_spend(self, tokens: int) -> bool:
        return (self.tick_used + tokens) <= self.budget_per_tick \
            and (self.cumulative_used + tokens) <= self.budget_cumulative

    def spend(self, tokens: int) -> bool:
        if not self.can_spend(tokens):
            return False
        self.tick_used += tokens
        self.cumulative_used += tokens
        return True

    def report(self) -> dict:
        return {
            "tick_used": self.tick_used,
            "tick_budget": self.budget_per_tick,
            "cumulative_used": self.cumulative_used,
            "cumulative_budget": self.budget_cumulative,
        }


# P7: human-in-the-loop
HUMAN_APPROVAL_THRESHOLD = 100  # trades >= this amount require admin approval


def needs_human_approval(amount: int) -> bool:
    return amount >= HUMAN_APPROVAL_THRESHOLD
