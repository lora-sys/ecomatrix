# ISSUE-030 — Supervisor detail and agent participation in the dashboard

## Context

ISS-029 surfaces supervisor runs as a flat log on the dashboard. Operators can
see the latest run's cost and status, but they cannot:

- Open a run to inspect its subtasks and per-worker receipts.
- See which agent participated in which run, or compare token cost across
  agents within the same run.
- Aggregate supervisor cost into the live metrics tiles.

The Phase 6 design audit flagged the supervisor as the long-term multi-agent
control surface, so the dashboard must grow from a single summary into a
browseable history.

## Goal

- Every supervisor run is reachable as a detail page in the dashboard.
- Agent detail pages show that agent's recent supervisor participations.
- Live metrics tiles expose supervisor cost and last-run timestamp.
- The log panel still behaves as before; new affordances are additive.

## Non-Goals

- Editing supervisor runs or replaying them.
- A new visual theme for the dashboard.
- Backend changes to the supervisor decomposition algorithm.

## Implementation Plan

1. Backend: `SupervisorRunsRepo.ByID`, `ByAgent` aggregation helper.
2. New endpoints:
   - `GET /v1/supervisor/runs/{id}` — single run with full JSONB columns.
   - `GET /v1/agents/by-string-id/{sid}/supervisor-runs?limit=N` — recent runs
     that mention the agent in their `worker_results` JSONB.
3. Extend `MetricsService.Snapshot` to include `supervisor_runs_count` and
   `supervisor_last_run_at`; the existing snapshot endpoint is updated to
   return the new fields.
4. Python `A2AClient` exposes `fetch_supervisor_run` and
   `fetch_agent_supervisor_runs`.
5. Frontend: new `/supervisor/[id]` page reusing the existing store; the
   `<SupervisorLog>` recent run list links to it. `agents/[id]` adds a small
   "Recent supervisor runs" section fed by SSR + the new endpoint.
6. E2E:
   - Dashboard renders the new supervisor link and follows it.
   - Agent detail renders the recent supervisor section without errors.
   - Existing dashboard test continues to pass.

## Acceptance Criteria

- [ ] `GET /v1/supervisor/runs/{id}` returns the full run including subtasks,
      worker results, and warnings.
- [ ] `GET /v1/agents/by-string-id/{sid}/supervisor-runs` returns at most
      `limit` runs that mention the agent.
- [ ] `/v1/metrics` returns `supervisor_runs_count` and `supervisor_last_run_at`.
- [ ] Dashboard `/supervisor/[id]` renders the run detail and links back to
      the dashboard.
- [ ] `agents/[id]` includes a "Recent supervisor runs" section when the
      agent has participated.
- [ ] Python tests + Go race + frontend typecheck/lint/build all pass.
- [ ] All four required CI jobs on the PR are green.

## Evidence Requirements

- `docs/evidence/ISSUE-030/change-summary.md`
- `docs/evidence/ISSUE-030/verification.md` with PR run id and per-job times.
- Bug Hunter and Architecture Reviewer re-reviews recorded in
  `docs/evidence/ISSUE-030/review-report.md`.

## Reviewer Requirements

- [ ] bug-hunter
- [ ] architecture-reviewer

## Related Docs

- `docs/evidence/PHASE-6-ai/design-audit.md`
- `docs/architecture/agent.md`
- `docs/architecture/frontend.md`

## Allowed Files

- `apps/backend/internal/repo/supervisor_runs_repo.go`
- `apps/backend/internal/transport/http/router.go`
- `apps/backend/internal/service/metrics.go`
- `apps/backend/cmd/server/main.go`
- `apps/agent/ecomatrix/a2a.py`
- `apps/agent/tests/test_runner.py`
- `apps/frontend/lib/api.ts`
- `apps/frontend/lib/types.ts`
- `apps/frontend/hooks/store.ts`
- `apps/frontend/app/page.tsx`
- `apps/frontend/app/agents/[id]/page.tsx`
- `apps/frontend/app/supervisor/[id]/page.tsx`
- `apps/frontend/components/supervisor-log.tsx`
- `apps/frontend/components/supervisor-list.tsx`
- `apps/frontend/components/agent-supervisor-history.tsx`
- `apps/frontend/e2e/dashboard.spec.ts`
- `apps/frontend/e2e/agent.spec.ts`
- `docs/evidence/ISSUE-030/**`
- `sessions/**`
- `tasks/ISSUE-030.md`
- `PROJECT_STATUS.md`
- `docs/product/roadmap.md`

## Linked Roadmap Item

- Phase 10, "Supervisor detail and agent participation in the dashboard".
