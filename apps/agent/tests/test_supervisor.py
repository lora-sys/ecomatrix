"""Tests for bounded hierarchical supervisor orchestration."""

import json

import pytest

import ecomatrix.supervisor as supervisor_module
from ecomatrix.cost import CostLedger
from ecomatrix.llm import LLMRateLimitError, MockLLMWithFailures, StubLLM
from ecomatrix.observability import TraceClient
from ecomatrix.supervisor import MAX_SUBTASKS, WorkerSpec, run_supervisor
from ecomatrix.workflows.react import ReActResult
from tests.test_react import FakeA2A


def disabled_trace() -> TraceClient:
    return TraceClient(agent_id="supervisor", enabled=False)


def workers() -> list[WorkerSpec]:
    return [
        WorkerSpec(
            agent_id="agent_miner_01",
            job_type="miner",
            description="trade and earn",
        ),
        WorkerSpec(
            agent_id="agent_merchant_01",
            job_type="merchant",
            description="buy and rebalance",
            weight=2.0,
        ),
    ]


class SupervisorLLM:
    def __init__(
        self,
        subtasks,
        *,
        aggregate_error: bool = False,
        aggregate_response: dict | None = None,
    ):
        self.subtasks = subtasks
        self.aggregate_error = aggregate_error
        self.aggregate_response = aggregate_response
        self.stub = StubLLM()
        self.decompose_prompt = ""

    def complete(self, messages, *, temperature=0.0):
        request = messages[-1].get("content", "")
        if request == "Decompose the goal.":
            self.decompose_prompt = messages[0]["content"]
            return json.dumps({"subtasks": self.subtasks})
        if request == "Aggregate the results.":
            if self.aggregate_error:
                raise LLMRateLimitError("aggregate unavailable")
            if self.aggregate_response is not None:
                return json.dumps(self.aggregate_response)
            return json.dumps({"summary": "Workers completed the delegated goal."})
        return self.stub.complete(messages, temperature=temperature)


def test_supervisor_falls_back_when_llm_does_not_return_subtasks():
    result = run_supervisor(
        goal="increase total gold by 10",
        workers=workers(),
        llm=StubLLM(),
        client=FakeA2A(),
        trace=disabled_trace(),
    )

    assert result.error is None
    assert len(result.subtasks) == 1
    assert result.subtasks[0]["target_agent"] == "agent_merchant_01"
    assert len(result.worker_results) == 1
    assert result.final_summary


def test_supervisor_handles_decompose_failure():
    result = run_supervisor(
        goal="trade",
        workers=workers(),
        llm=MockLLMWithFailures(failure_rate=1.0),
        client=FakeA2A(),
        trace=disabled_trace(),
    )

    assert result.error is not None
    assert "decompose" in result.error.lower()
    assert result.worker_results == []


def test_supervisor_routes_by_agent_then_job_type():
    llm = SupervisorLLM(
        [
            {
                "subtask": "miner task",
                "target_job_type": "merchant",
                "target_agent": "agent_miner_01",
            },
            {
                "subtask": "merchant task",
                "target_job_type": "merchant",
                "target_agent": "missing",
            },
        ]
    )

    result = run_supervisor(
        goal="rebalance",
        workers=workers(),
        llm=llm,
        client=FakeA2A(),
        trace=disabled_trace(),
    )

    assert [item["agent_id"] for item in result.worker_results] == [
        "agent_miner_01",
        "agent_merchant_01",
    ]


def test_supervisor_uses_contract_constraints_and_stable_tie_break():
    registered = [
        WorkerSpec(
            agent_id="agent_miner_02",
            job_type="miner",
            description="mine",
            capabilities=("execute_trade",),
            limitations=("never self trade",),
        ),
        WorkerSpec(
            agent_id="agent_miner_01",
            job_type="miner",
            description="mine",
            capabilities=("execute_trade",),
            limitations=("never self trade",),
        ),
    ]
    llm = SupervisorLLM([{"subtask": "mine", "target_job_type": "miner"}])

    result = run_supervisor(
        goal="mine",
        workers=registered,
        llm=llm,
        client=FakeA2A(),
        trace=disabled_trace(),
    )

    assert result.worker_results[0]["agent_id"] == "agent_miner_01"
    assert "Capabilities: execute_trade" in llm.decompose_prompt
    assert "Limitations: never self trade" in llm.decompose_prompt


def test_supervisor_skips_malformed_subtasks_and_caps_dispatch():
    generated = [
        None,
        {"subtask": ""},
        *[
            {"subtask": f"task {index}", "target_job_type": "miner"}
            for index in range(MAX_SUBTASKS + 3)
        ],
    ]

    result = run_supervisor(
        goal="bounded work",
        workers=workers(),
        llm=SupervisorLLM(generated),
        client=FakeA2A(),
        trace=disabled_trace(),
    )

    assert len(result.subtasks) == MAX_SUBTASKS
    assert len(result.worker_results) == MAX_SUBTASKS
    assert [subtask["subtask"] for subtask in result.subtasks] == [
        f"task {index}" for index in range(MAX_SUBTASKS)
    ]


@pytest.mark.parametrize(
    ("goal", "registered_workers", "max_subtasks", "expected"),
    [
        ("", workers(), 1, "goal"),
        ("trade", [], 1, "worker"),
        ("trade", workers(), 0, "max_subtasks"),
        ("trade", workers(), MAX_SUBTASKS + 1, "max_subtasks"),
    ],
)
def test_supervisor_returns_structured_validation_errors(
    goal, registered_workers, max_subtasks, expected
):
    result = run_supervisor(
        goal=goal,
        workers=registered_workers,
        llm=StubLLM(),
        client=FakeA2A(),
        trace=disabled_trace(),
        max_subtasks=max_subtasks,
    )

    assert result.error is not None
    assert expected in result.error
    assert result.to_dict()["error"] == result.error


def test_supervisor_continues_after_worker_exception(monkeypatch):
    calls = []

    def fake_run_react(**kwargs):
        calls.append(kwargs["agent_id"])
        if len(calls) == 1:
            raise RuntimeError("worker crashed")
        return ReActResult(1, {"action": "SKIP"}, None, [])

    monkeypatch.setattr(supervisor_module, "run_react", fake_run_react)
    llm = SupervisorLLM(
        [
            {"subtask": "first", "target_agent": "agent_miner_01"},
            {"subtask": "second", "target_agent": "agent_merchant_01"},
        ]
    )

    result = run_supervisor(
        goal="run both",
        workers=workers(),
        llm=llm,
        client=FakeA2A(),
        trace=disabled_trace(),
    )

    assert calls == ["agent_miner_01", "agent_merchant_01"]
    assert "worker failed" in result.worker_results[0]["error"]
    assert result.worker_results[1]["error"] is None


def test_supervisor_surfaces_react_action_failure(monkeypatch):
    def fake_run_react(**kwargs):
        return ReActResult(
            1,
            {"action": "EXECUTE_TRADE"},
            None,
            [{"step": "act", "error": "insufficient funds"}],
        )

    monkeypatch.setattr(supervisor_module, "run_react", fake_run_react)
    result = run_supervisor(
        goal="trade",
        workers=workers(),
        llm=SupervisorLLM(
            [{"subtask": "trade", "target_agent": "agent_miner_01"}]
        ),
        client=FakeA2A(),
        trace=disabled_trace(),
    )

    assert result.worker_results[0]["error"] == "action failed: insufficient funds"


def test_supervisor_uses_fallback_summary_when_aggregation_fails():
    result = run_supervisor(
        goal="earn 5",
        workers=workers(),
        llm=SupervisorLLM(
            [{"subtask": "trade", "target_job_type": "miner"}],
            aggregate_error=True,
        ),
        client=FakeA2A(),
        trace=disabled_trace(),
    )

    assert result.error is None
    assert result.final_summary.startswith("Completed 1 subtask")
    assert result.warnings == ["aggregation failed: aggregate unavailable"]


@pytest.mark.parametrize("aggregate_response", [{"summary": ""}, {"foo": "bar"}])
def test_supervisor_uses_fallback_summary_when_aggregation_is_empty(
    aggregate_response,
):
    result = run_supervisor(
        goal="earn 5",
        workers=workers(),
        llm=SupervisorLLM(
            [{"subtask": "trade", "target_job_type": "miner"}],
            aggregate_response=aggregate_response,
        ),
        client=FakeA2A(),
        trace=disabled_trace(),
    )

    assert result.final_summary.startswith("Completed 1 subtask")
    assert result.warnings == ["aggregation returned an empty summary"]


def test_supervisor_enforces_shared_token_budget():
    result = run_supervisor(
        goal="trade",
        workers=workers(),
        llm=StubLLM(),
        client=FakeA2A(),
        trace=disabled_trace(),
        cost_ledger=CostLedger(budget_per_tick=1, budget_cumulative=1),
    )

    assert result.error == "decompose failed: supervisor token budget exhausted"
    assert result.cost["tick_used"] == 0
