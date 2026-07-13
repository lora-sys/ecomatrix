# ISSUE-028 — Hierarchical Multi-Agent Supervisor

## Context

Phase 6 delivered per-agent contracts, ReAct execution, traces, cost controls,
and golden evaluations. Its design audit explicitly deferred hierarchical
multi-agent supervision because the current `multi` scenario runs agents as a
flat peer set.

An initial untracked supervisor prototype exists, but it is not CLI-reachable,
fails three tests during result aggregation, and does not validate LLM-produced
subtasks or an empty worker registry.

Live CLI verification also exposed a Phase 5 regression: Go `Receipt` JSON
tags were dropped, so `/v1/trades` emitted PascalCase receipt fields that the
authoritative A2A v1.1 Python client could not parse.

The required parallel Go race gate then exposed repo/service test binaries
truncating the same public schema. Service tests need an isolated ephemeral
schema so the documented command is reliable.

## Goal

Ship a bounded supervisor workflow that decomposes one goal, routes validated
subtasks to specialized ReAct workers, aggregates their receipts, and can be
run from the existing agent CLI.

## Non-Goals

- Parallel subtask execution or dependency graphs.
- A new LangGraph supervisor graph.
- MCP exposure, vector memory, or LLM-as-judge evaluation.
- Production database schema or frontend changes.

## Implementation Plan

1. Validate the supervisor input and worker registry.
2. Normalize and cap LLM-produced subtasks before dispatch.
3. Route by explicit agent, then job type, then worker weight.
4. Execute each subtask through the existing bounded ReAct workflow.
5. Enforce one shared `CostLedger` budget across supervisor and worker LLM calls.
6. Aggregate compact worker results with graceful LLM failure handling.
7. Add `--scenario supervisor --goal ...` to the existing CLI.
8. Add unit and CLI-dispatch tests, then capture evidence.
9. Restore the A2A receipt's snake_case wire shape and prove it live.
10. Isolate service tests from repo-package database cleanup.

## Acceptance Criteria

- [x] Empty goals, empty workers, and invalid bounds return a structured error without raising.
- [x] Malformed or empty decomposition falls back to one bounded subtask.
- [x] More than `max_subtasks` generated tasks are never dispatched.
- [x] Worker routing is deterministic and respects explicit agent/job targeting.
- [x] Every dispatched subtask runs through `run_react`; one worker failure does not stop later workers.
- [x] Worker contracts inform decomposition and all LLM calls share one bounded token ledger.
- [x] Aggregation returns a non-empty summary, including when its LLM call fails.
- [x] `python -m ecomatrix.runner --scenario supervisor --goal ...` uses seeded backend agents and emits JSON.
- [x] Existing `two_agent` and `multi` CLI behavior remains unchanged.
- [x] CLI exits non-zero when any delegated worker fails, while still returning all worker results.
- [x] Full Python pytest passes; touched Python files pass Ruff.
- [x] Go receipt JSON matches A2A v1.1; `go test -race ./...` passes.

## Evidence Requirements

- `docs/evidence/ISSUE-028/change-summary.md`
- `docs/evidence/ISSUE-028/verification.md`
- Full Python pytest output and touched-file Ruff output recorded in verification.
- Backend `go test -race ./...` output and live CLI trace recorded in verification.
- Bug Hunter and Architecture Reviewer reports recorded in evidence.

## Reviewer Requirements

- [x] bug-hunter
- [x] architecture-reviewer

## Related Docs

- `docs/evidence/PHASE-6-ai/design-audit.md`
- `docs/evidence/PHASE-6-ai/REPORT.md`
- `docs/architecture/agent.md`
- `ENGINEERING.md`
- `TESTING.md`

## Allowed Files

- `apps/agent/ecomatrix/supervisor.py`
- `apps/agent/ecomatrix/runner.py`
- `apps/agent/tests/test_supervisor.py`
- `apps/agent/tests/test_runner.py`
- `apps/backend/internal/domain/transaction.go`
- `apps/backend/internal/domain/transaction_test.go`
- `apps/backend/internal/service/trade_test.go`
- `tasks/ISSUE-028.md`
- `docs/product/roadmap.md`
- `docs/evidence/ISSUE-028/**`
- `sessions/**`
- `PROJECT_STATUS.md`

## Linked Roadmap Item

- Phase 8, "Hierarchical supervisor workflow".
