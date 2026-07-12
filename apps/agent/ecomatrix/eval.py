"""Evaluation framework — golden test cases + report.

Per the design spec: must test task completion, tool selection, and
failure recovery. Not just "can it answer". This module runs a battery
of scenarios and reports success rate, latency, cost.
"""
from __future__ import annotations

import time
import json
from dataclasses import dataclass, field
from typing import Any, Callable


@dataclass
class EvalCase:
    name: str
    description: str
    run: Callable[[], "EvalResult"]


@dataclass
class EvalResult:
    case: str
    passed: bool
    duration_ms: int
    notes: str = ""
    cost_tokens: int = 0


@dataclass
class EvalReport:
    total: int = 0
    passed: int = 0
    total_duration_ms: int = 0
    total_cost_tokens: int = 0
    cases: list[EvalResult] = field(default_factory=list)

    def pass_rate(self) -> float:
        return self.passed / self.total if self.total else 0.0

    def avg_duration_ms(self) -> float:
        return self.total_duration_ms / self.total if self.total else 0.0

    def to_dict(self) -> dict:
        return {
            "total": self.total,
            "passed": self.passed,
            "pass_rate": self.pass_rate(),
            "avg_duration_ms": self.avg_duration_ms(),
            "total_cost_tokens": self.total_cost_tokens,
            "cases": [
                {"case": c.case, "passed": c.passed, "duration_ms": c.duration_ms,
                 "notes": c.notes, "cost_tokens": c.cost_tokens}
                for c in self.cases
            ],
        }


def run_eval(cases: list[EvalCase]) -> EvalReport:
    report = EvalReport()
    report.total = len(cases)
    for c in cases:
        t0 = time.time()
        try:
            r = c.run()
        except Exception as e:
            r = EvalResult(case=c.name, passed=False, duration_ms=0, notes=f"raised: {e!r}")
        if r.duration_ms == 0:
            r.duration_ms = int((time.time() - t0) * 1000)
        report.cases.append(r)
        report.total_duration_ms += r.duration_ms
        report.total_cost_tokens += r.cost_tokens
        if r.passed:
            report.passed += 1
    return report


# === Default golden cases =====================================================


def case_stubllm_executes_trade():
    """Smoke: StubLLM makes a single trade in one tick."""
    from ecomatrix.llm import StubLLM
    from ecomatrix.workflows.react import run_react
    from tests.test_react import FakeA2A

    client = FakeA2A()
    llm = StubLLM()
    t0 = time.time()
    r = run_react(agent_id="agent_miner_01", job_type="miner", llm=llm, client=client,
                 initial_observation="balance=100", goal="trade")
    duration = int((time.time() - t0) * 1000)
    passed = r.final_receipt is not None and r.error is None
    return EvalResult(case="stubllm_executes_trade", passed=passed,
                      duration_ms=duration,
                      notes=f"receipt={r.final_receipt}", cost_tokens=200)


def case_handles_llm_failure():
    """Failure: an LLM that always fails should not crash the agent."""
    from ecomatrix.llm import MockLLMWithFailures, LLMRateLimitError
    from ecomatrix.workflows.react import run_react
    from tests.test_react import FakeA2A

    client = FakeA2A()
    llm = MockLLMWithFailures(failure_rate=1.0, failure_kind=LLMRateLimitError)
    t0 = time.time()
    r = run_react(agent_id="agent_miner_01", job_type="miner", llm=llm, client=client,
                 initial_observation="balance=100", goal="trade")
    duration = int((time.time() - t0) * 1000)
    passed = r.error is not None and r.final_receipt is None
    return EvalResult(case="handles_llm_failure", passed=passed,
                      duration_ms=duration,
                      notes=f"error={r.error}", cost_tokens=0)


def case_respects_max_iterations():
    """Loop LLM is bounded by max_iterations."""
    from ecomatrix.workflows.react import run_react
    from tests.test_react import FakeA2A, LoopLLM

    client = FakeA2A()
    t0 = time.time()
    r = run_react(agent_id="agent_miner_01", job_type="miner", llm=LoopLLM(), client=client,
                 initial_observation="x", goal="loop", max_iterations=3)
    duration = int((time.time() - t0) * 1000)
    tools = sum(1 for s in r.transcript if s.get("step") == "tool")
    passed = tools <= 3  # not unbounded
    return EvalResult(case="respects_max_iterations", passed=passed,
                      duration_ms=duration,
                      notes=f"tool_calls={tools}")


def case_contract_loadable():
    """All four agent types have a contract."""
    from ecomatrix.contracts import CONTRACTS
    required = {"miner", "merchant", "hacker", "mediator"}
    passed = required.issubset(set(CONTRACTS))
    return EvalResult(case="contract_loadable", passed=passed, duration_ms=0,
                      notes=f"contracts={list(CONTRACTS)}")


def case_human_approval_threshold():
    """Trades >= 100 GOLD require human approval."""
    from ecomatrix.cost import needs_human_approval
    return EvalResult(
        case="human_approval_threshold",
        passed=needs_human_approval(150) and not needs_human_approval(50),
        duration_ms=0,
        notes="100 GOLD threshold",
    )


DEFAULT_CASES = [
    EvalCase(name=c.__name__, description=c.__doc__ or "", run=c)
    for c in (case_stubllm_executes_trade, case_handles_llm_failure,
             case_respects_max_iterations, case_contract_loadable,
             case_human_approval_threshold)
]
