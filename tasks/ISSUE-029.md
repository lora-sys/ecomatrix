# ISSUE-029 — Supervisor task stream in the dashboard

## Context

ISS-028 shipped the bounded hierarchical supervisor as a Python CLI scenario,
but its runs are only visible via stdout. Operators currently have to invoke
`--scenario supervisor` to see the cost budget, subtasks, and per-worker
receipts; there is no way to inspect historical runs or react to a new one
from the dashboard.

The Phase 6 design audit flagged "supervisor over free chat" as the long-term
shape of multi-agent control. Exposing the supervisor loop in the same
LiveProvider that already drives trades, metrics, and social feeds closes the
gap between back-office and front-office observability.

## Goal

Stream every supervisor run to the dashboard as a live, paginated task log
with a small cost chart, and persist enough state to render history after a
restart.

## Non-Goals

- Multi-user concurrent run coordination; one supervisor at a time is enough
  for the demo economy.
- Changing the supervisor decomposition algorithm.
- A new "supervisor control panel" beyond the log: the run is still
  triggered from the CLI in this issue.

## Implementation Plan

1. Migration `0007_supervisor_runs`: `id`, `goal`, `subtask_count`, `error`,
   `warnings`, `started_at`, `finished_at`, `tokens_used`, `tokens_budget`,
   JSONB `workers` (array of agent_id, action, receipt, error).
2. `SupervisorRunsRepo` (GORM) with `Insert` and `Recent(limit)`.
3. New endpoint `POST /v1/supervisor/runs` (admin token) and
   `GET /v1/supervisor/runs?limit=N`.
4. WebSocket event `supervisor.run.{started, finished}` carrying the run
   summary; LiveProvider applies it to the store.
5. Python `--scenario supervisor` now POSTs the `SupervisorResult` back to
   the backend after a successful run.
6. Dashboard `SupervisorLog` component, two-column on desktop (recent runs
   list + current cost bar), live region for the latest finished run.
7. Hook the panel into the SSR feed on the dashboard so first paint shows
   the last persisted run.
8. Playwright e2e: dashboard renders the supervisor log panel; with the
   supervisor disabled, the empty state is shown.

## Acceptance Criteria

- [ ] Each `python -m ecomatrix.runner --scenario supervisor` run is recorded
  in `supervisor_runs` and appears in `GET /v1/supervisor/runs`.
- [ ] A live `supervisor.run.finished` event drives a `supervisor` slice update
  on every connected dashboard within 1 s.
- [ ] The dashboard SSR payload includes the last run so first paint shows
  the cost bar with the previous run's number.
- [ ] `apps/agent` and `apps/backend` build cleanly; full Python and Go suites
  pass; dashboard `tsc`/`lint`/`build` clean.
- [ ] Playwright dashboard test asserts the supervisor log section is
  visible.
- [ ] All GitHub Actions jobs on the PR are green.

## Evidence Requirements

- `docs/evidence/ISSUE-029/change-summary.md`
- `docs/evidence/ISSUE-029/verification.md` with PR run id and per-job times.
- Bug Hunter and Architecture Reviewer reports in
  `docs/evidence/ISSUE-029/review-report.md`.

## Reviewer Requirements

- [ ] bug-hunter
- [ ] architecture-reviewer

## Related Docs

- `docs/evidence/PHASE-6-ai/design-audit.md`
- `docs/architecture/agent.md`
- `apps/frontend/hooks/store.ts` (existing LiveProvider pattern)
- `apps/agent/ecomatrix/supervisor.py`

## Allowed Files

- `apps/backend/internal/repo/migrations_fs/0007_supervisor_runs.up.sql`
- `apps/backend/internal/repo/migrations_fs/0007_supervisor_runs.down.sql`
- `apps/backend/internal/repo/supervisor_runs_repo.go`
- `apps/backend/internal/domain/supervisor.go`
- `apps/backend/internal/domain/supervisor_test.go`
- `apps/backend/internal/transport/http/router.go`
- `apps/backend/cmd/server/main.go`
- `apps/agent/ecomatrix/supervisor.py`
- `apps/agent/ecomatrix/runner.py`
- `apps/agent/ecomatrix/a2a.py`
- `apps/agent/ecomatrix/types.py`
- `apps/agent/tests/test_supervisor.py`
- `apps/agent/tests/test_runner.py`
- `apps/frontend/hooks/store.ts`
- `apps/frontend/lib/types.ts`
- `apps/frontend/lib/api.ts`
- `apps/frontend/components/supervisor-log.tsx`
- `apps/frontend/components/dashboard-client.tsx`
- `apps/frontend/app/page.tsx`
- `apps/frontend/e2e/dashboard.spec.ts`
- `apps/backend/migrations/0007_supervisor_runs.up.sql`
- `apps/backend/migrations/0007_supervisor_runs.down.sql`
- `docs/product/roadmap.md`
- `docs/architecture/agent.md`
- `docs/evidence/ISSUE-029/**`
- `sessions/**`
- `tasks/ISSUE-029.md`
- `PROJECT_STATUS.md`
- `CHANGELOG.md`

## Linked Roadmap Item

- Phase 9, "Supervisor stream in the dashboard".
