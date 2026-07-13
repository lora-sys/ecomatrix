"""Bounded hierarchical orchestration for specialized EcoMatrix agents."""

from __future__ import annotations

import json
import math
import time
from dataclasses import asdict, dataclass, field
from typing import Any

from .a2a import A2AClient
from .cost import CostLedger
from .llm import LLM, LLMError, LLMProviderError, parse_action_json
from .observability import TraceClient
from .workflows.react import run_react


MAX_SUBTASKS = 4
# Supervisor subtasks are narrower than free-form ReAct goals.
WORKER_MAX_ITERATIONS = 3
SUPERVISOR_TOKEN_BUDGET = 12_000
LLM_RESPONSE_TOKEN_RESERVE = 512


@dataclass(frozen=True)
class WorkerSpec:
    """A specialized worker available to the supervisor."""

    agent_id: str
    job_type: str
    description: str
    capabilities: tuple[str, ...] = ()
    limitations: tuple[str, ...] = ()
    weight: float = 1.0


@dataclass
class SupervisorResult:
    goal: str
    subtasks: list[dict[str, Any]]
    worker_results: list[dict[str, Any]]
    final_summary: str
    duration_ms: int
    error: str | None = None
    warnings: list[str] = field(default_factory=list)
    cost: dict[str, int] = field(default_factory=dict)

    def to_dict(self) -> dict[str, Any]:
        return asdict(self)


def _elapsed_ms(started_at: float) -> int:
    return max(0, int((time.monotonic() - started_at) * 1000))


def _invalid_result(
    goal: str,
    started_at: float,
    message: str,
    *,
    cost: dict[str, int] | None = None,
) -> SupervisorResult:
    return SupervisorResult(
        goal=goal,
        subtasks=[],
        worker_results=[],
        final_summary="",
        duration_ms=_elapsed_ms(started_at),
        error=message,
        cost=cost or {},
    )


def _validate_inputs(
    goal: str, workers: list[WorkerSpec], max_subtasks: int
) -> str | None:
    if not goal.strip():
        return "goal must not be empty"
    if not workers:
        return "at least one worker is required"
    if not 1 <= max_subtasks <= MAX_SUBTASKS:
        return f"max_subtasks must be between 1 and {MAX_SUBTASKS}"

    seen: set[str] = set()
    for worker in workers:
        if not worker.agent_id.strip() or not worker.job_type.strip():
            return "every worker requires a non-empty agent_id and job_type"
        if worker.agent_id in seen:
            return f"duplicate worker agent_id: {worker.agent_id}"
        if not math.isfinite(worker.weight) or worker.weight <= 0:
            return f"worker weight must be positive: {worker.agent_id}"
        seen.add(worker.agent_id)
    return None


def _fallback_subtask(goal: str, workers: list[WorkerSpec]) -> dict[str, Any]:
    worker = max(workers, key=lambda candidate: candidate.weight)
    return {
        "subtask": goal,
        "target_job_type": worker.job_type,
        "target_agent": worker.agent_id,
        "reasoning": "fallback after empty or malformed decomposition",
    }


def _normalize_subtasks(
    raw_subtasks: Any,
    *,
    goal: str,
    workers: list[WorkerSpec],
    max_subtasks: int,
) -> list[dict[str, Any]]:
    normalized: list[dict[str, Any]] = []
    if isinstance(raw_subtasks, list):
        for raw in raw_subtasks:
            if len(normalized) >= max_subtasks:
                break
            if not isinstance(raw, dict):
                continue
            subtask = str(raw.get("subtask") or "").strip()
            if not subtask:
                continue
            normalized.append(
                {
                    "subtask": subtask,
                    "target_job_type": str(raw.get("target_job_type") or "").strip(),
                    "target_agent": str(raw.get("target_agent") or "").strip(),
                    "reasoning": str(raw.get("reasoning") or "").strip(),
                }
            )
    return normalized or [_fallback_subtask(goal, workers)]


def _route_worker(subtask: dict[str, Any], workers: list[WorkerSpec]) -> WorkerSpec:
    target_agent = subtask.get("target_agent")
    for worker in workers:
        if worker.agent_id == target_agent:
            return worker

    target_job = subtask.get("target_job_type")
    matching_job = [worker for worker in workers if worker.job_type == target_job]
    candidates = matching_job or workers
    return sorted(candidates, key=lambda candidate: (-candidate.weight, candidate.agent_id))[0]


def _fallback_summary(worker_results: list[dict[str, Any]]) -> str:
    failed = sum(1 for result in worker_results if result["error"])
    receipts = sum(
        1 for result in worker_results if result["final_receipt"] is not None
    )
    succeeded = len(worker_results) - failed
    return (
        f"Completed {len(worker_results)} subtask(s): {succeeded} succeeded, "
        f"{failed} failed, and {receipts} produced a receipt."
    )


def _worker_trace(parent: TraceClient, agent_id: str) -> TraceClient:
    return TraceClient(
        base_url=parent.base_url, agent_id=agent_id, enabled=parent.enabled
    )


def _estimate_tokens(messages: list[dict[str, str]]) -> int:
    characters = sum(len(message.get("content", "")) for message in messages)
    return max(1, (characters + 3) // 4)


@dataclass
class _BudgetedLLM:
    delegate: LLM
    ledger: CostLedger

    def complete(
        self, messages: list[dict[str, str]], *, temperature: float = 0.4
    ) -> str:
        input_tokens = _estimate_tokens(messages)
        reserved_tokens = input_tokens + LLM_RESPONSE_TOKEN_RESERVE
        if not self.ledger.spend(reserved_tokens):
            raise LLMProviderError("supervisor token budget exhausted", retryable=False)

        raw = self.delegate.complete(messages, temperature=temperature)
        actual_tokens = input_tokens + max(1, (len(raw) + 3) // 4)
        extra_tokens = max(0, actual_tokens - reserved_tokens)
        if extra_tokens and not self.ledger.spend(extra_tokens):
            raise LLMProviderError("supervisor token budget exhausted", retryable=False)
        return raw


def run_supervisor(
    *,
    goal: str,
    workers: list[WorkerSpec],
    llm: LLM,
    client: A2AClient,
    trace: TraceClient | None = None,
    max_subtasks: int = MAX_SUBTASKS,
    cost_ledger: CostLedger | None = None,
) -> SupervisorResult:
    """Decompose one goal, run bounded ReAct workers, and aggregate results."""

    started_at = time.monotonic()
    if cost_ledger is None:
        cost_ledger = CostLedger(
            budget_per_tick=SUPERVISOR_TOKEN_BUDGET,
            budget_cumulative=SUPERVISOR_TOKEN_BUDGET,
        )
    validation_error = _validate_inputs(goal, workers, max_subtasks)
    if validation_error:
        return _invalid_result(
            goal, started_at, validation_error, cost=cost_ledger.report()
        )

    if trace is None:
        trace = TraceClient.from_env("supervisor")
    trace.plan(f"supervisor goal={goal}")

    workers_prompt = "\n".join([
        f"- {worker.agent_id} ({worker.job_type}, weight={worker.weight:g})\n"
        f"  Mission: {worker.description}\n"
        f"  Capabilities: {', '.join(worker.capabilities) or '(none declared)'}\n"
        f"  Limitations: {', '.join(worker.limitations) or '(none declared)'}"
        for worker in workers
    ])
    budgeted_llm = _BudgetedLLM(llm, cost_ledger)
    decompose_prompt = f"""\
You are the EcoMatrix supervisor. Decompose the goal into 1-{max_subtasks}
independent subtasks. Assign each subtask to exactly one available worker.
Do not invent workers or job types.

## Goal
{goal}

## Available workers
{workers_prompt}

## Output
Return JSON with this shape:
{{"subtasks": [{{"subtask": "...", "target_job_type": "...", "target_agent": "...", "reasoning": "..."}}]}}
"""
    try:
        raw = budgeted_llm.complete(
            [
                {"role": "system", "content": decompose_prompt},
                {"role": "user", "content": "Decompose the goal."},
            ],
            temperature=0.0,
        )
        decomposed = parse_action_json(raw)
    except LLMError as exc:
        trace.error(f"decomposition failed: {exc}", code=type(exc).__name__)
        return _invalid_result(
            goal,
            started_at,
            f"decompose failed: {exc}",
            cost=cost_ledger.report(),
        )

    subtasks = _normalize_subtasks(
        decomposed.get("subtasks"),
        goal=goal,
        workers=workers,
        max_subtasks=max_subtasks,
    )

    worker_results: list[dict[str, Any]] = []
    for index, subtask in enumerate(subtasks, start=1):
        worker = _route_worker(subtask, workers)
        trace.decision(
            f"dispatch {index}/{len(subtasks)} to {worker.agent_id}: "
            f"{subtask['subtask'][:120]}"
        )
        try:
            result = run_react(
                agent_id=worker.agent_id,
                job_type=worker.job_type,
                llm=budgeted_llm,
                client=client,
                initial_observation=f"delegated subtask: {subtask['subtask']}",
                goal=subtask["subtask"],
                max_iterations=WORKER_MAX_ITERATIONS,
                trace=_worker_trace(trace, worker.agent_id),
            )
            action_errors = [
                str(step["error"])
                for step in result.transcript
                if step.get("step") == "act" and step.get("error")
            ]
            worker_error = result.error
            if worker_error is None and action_errors:
                worker_error = f"action failed: {action_errors[-1]}"
            worker_results.append(
                {
                    "subtask": subtask["subtask"],
                    "agent_id": worker.agent_id,
                    "iterations": result.iterations,
                    "final_action": result.final_action,
                    "final_receipt": result.final_receipt,
                    "transcript": result.transcript,
                    "error": worker_error,
                }
            )
        except Exception as exc:
            message = f"worker failed: {type(exc).__name__}: {exc}"
            trace.error(message, code="WORKER_ERROR")
            worker_results.append(
                {
                    "subtask": subtask["subtask"],
                    "agent_id": worker.agent_id,
                    "iterations": 0,
                    "final_action": None,
                    "final_receipt": None,
                    "transcript": [],
                    "error": message,
                }
            )

    compact_results = [
        {
            "subtask": result["subtask"],
            "agent": result["agent_id"],
            "action": result["final_action"],
            "receipt": result["final_receipt"],
            "error": result["error"],
        }
        for result in worker_results
    ]
    aggregate_prompt = f"""\
Summarize the worker results against the original goal. State completed work,
failures, and concrete receipts. Do not claim success for failed work.

## Original goal
{goal}

## Worker results
{json.dumps(compact_results, ensure_ascii=False, indent=2)}

## Output
Return JSON: {{"summary": "...", "key_findings": ["..."]}}
"""
    warnings: list[str] = []
    try:
        raw_aggregate = budgeted_llm.complete(
            [
                {"role": "system", "content": aggregate_prompt},
                {"role": "user", "content": "Aggregate the results."},
            ],
            temperature=0.0,
        )
        aggregated = parse_action_json(raw_aggregate)
        final_summary = str(aggregated.get("summary") or "").strip()
        if not final_summary:
            warnings.append("aggregation returned an empty summary")
            final_summary = _fallback_summary(worker_results)
    except LLMError as exc:
        warnings.append(f"aggregation failed: {exc}")
        final_summary = _fallback_summary(worker_results)

    trace.reflection(f"supervisor complete: {len(worker_results)} subtasks")
    return SupervisorResult(
        goal=goal,
        subtasks=subtasks,
        worker_results=worker_results,
        final_summary=final_summary,
        duration_ms=_elapsed_ms(started_at),
        warnings=warnings,
        cost=cost_ledger.report(),
    )
