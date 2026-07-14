# Session — ISS-029 supervisor task stream

- Started: 2026-07-13T12:05:00Z
- Coordinator: Codex
- Phase: 9
- Branch: `feature/ISSUE-029-supervisor-tasks`

## Plan

1. Persist supervisor runs on the backend and expose REST + WS endpoints.
2. Mirror migrations and Python client helper to POST summaries.
3. Dashboard panel with SSR hydration and live updates.

## Outcomes

- Migration `0007_supervisor_runs` adds a `supervisor_runs` table with
  `subtasks`, `worker_results`, and `warnings` as JSONB plus the cost counters
  and timestamps.
- `SupervisorRunsRepo` (GORM) handles insert and recent reads; both
  `apps/backend/internal/repo/migrations_fs/0007_*.sql` and
  `apps/backend/migrations/0007_*.sql` carry the migration.
- HTTP endpoints `GET/POST /v1/supervisor/runs` added to the `Server` and
  wired through `cmd/server/main.go`.
- Python `A2AClient` exposes `post_supervisor_run` and `list_supervisor_runs`;
  the supervisor CLI scenario now forwards every run to the backend.
- Dashboard `SupervisorLog` component merges SSR-fetched runs and live WS
  deltas via a new `supervisorRuns`/`supervisorLatest` slice in the store.
- `apps/frontend/e2e/dashboard.spec.ts` asserts the supervisor log live region
  is visible on first paint.
- Local gates: Python 101/101, Go race suite green, frontend `tsc`/`lint`/
  `build` clean (First Load 255 KB).

## Handoff

- Branch is committed; PR opening is the final step.
- Evidence: `docs/evidence/ISSUE-029/` (to be filled with the PR run id and
  per-job times).
