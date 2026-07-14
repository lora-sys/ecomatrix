# ISS-030 Change Summary

## Shipped

### Backend (Go)

- `SupervisorRunsRepo.ByID(ctx, id)` returns the full supervisor run
  including JSONB columns (`subtasks`, `worker_results`, `warnings`); missing
  rows translate `gorm.ErrRecordNotFound` to the new domain sentinel
  `ErrSupervisorRunNotFound` and surface as HTTP 404.
- `SupervisorRunsRepo.ByAgent(ctx, agentID, limit)` filters the runs whose
  `worker_results` JSONB mentions the given agent id (newest first).
- `GET /v1/supervisor/runs/{id}` — returns the canonical
  `supervisorRunPayload` for one run.
- `GET /v1/agents/by-string-id/{sid}/supervisor-runs?limit=N` — returns recent
  runs that mention the agent. Returns 404 when the agent does not exist.
- `MetricsService.Snapshot` gains `supervisor_runs_count` (int64) and
  `supervisor_last_run_at` (RFC3339Nano, optional).
- Tests added:
  - `internal/repo/supervisor_runs_repo_test.go` — roundtrip + not-found +
    agent filter.
  - `internal/transport/http/router_supervisor_test.go` — Fiber-level
    coverage for the two new endpoints (OK / 404 / not-found run / unknown
    agent).
  - `internal/service/metrics_test.go` —
    `TestMetricsService_Snapshot_IncludesSupervisorFields` exercises the
    empty-state zeros and the populated snapshot, including
    `supervisor_last_run_at` parsing.

### Python agent

- `A2AClient.fetch_supervisor_run(run_id)` and
  `A2AClient.list_agent_supervisor_runs(agent_id, limit)` mirror the new
  backend routes. Surfaces 4xx via `A2AError` with the HTTP status
  preserved.
- New test file `apps/agent/tests/test_supervisor_detail.py` covers happy
  path, 404 propagation, query parameter wiring, and the empty `runs` list
  fallback.

### Frontend (Next.js)

- `lib/types.ts` — `MetricsSnapshot` now declares `supervisor_runs_count` and
  `supervisor_last_run_at?`.
- `lib/api.ts` — `fetchSupervisorRun(id)` and `fetchAgentSupervisorRuns(...)`.
- `components/supervisor-log.tsx` — wraps the latest run goal in a
  `/supervisor/[id]` link and the historical list items click through to the
  detail page using `data-supervisor-link` selectors (so Playwright can target
  them). Recent runs header shows “详情 →” so the affordance is visible even
  before any click.
- `components/agent-supervisor-history.tsx` — pre-existing list component is
  now wired into the agent detail page; each entry links to
  `/supervisor/[id]`.
- `app/supervisor/[id]/page.tsx` — new SSR page; fetches the run via the new
  BFF-less `fetchSupervisorRun` helper, maps the payload, and renders the
  existing `SupervisorRunDetail` component inside a `GlowingCard`. Returns
  `notFound()` for unknown ids.
- `app/agents/[id]/page.tsx` — adds a second grid section that fetches
  `fetchAgentSupervisorRuns` in parallel and renders
  `AgentSupervisorHistory`. Empty state shows the “还没有被 Supervisor 调度过”
  copy.
- E2E: new `apps/frontend/e2e/supervisor.spec.ts` covers (a) clicking the
  dashboard link lands on `/supervisor/{id}` with the title and the back
  link, (b) the agent page exposes the new section without console errors.
  The fallback branch navigates to `/supervisor/999999` when the world has no
  runs and still confirms the not-found page is well-formed.

## Verification

- Local: `cd apps/backend && go test -race -count=1 ./...` — green across all
  packages.
- Local: `cd apps/agent && uv run pytest -q` — 105 passed.
- Local: `cd apps/frontend && npm run typecheck && npm run lint &&
  npm run build` — clean; `tsc` 0 errors, `next lint` 0 warnings,
  `next build` emits First-Load JS of 149 kB on `/supervisor/[id]`.
- E2E (Playwright, desktop + mobile): **20 passed, 2 skipped**; the two new
  supervisor tests both green on each viewport, with screenshots saved under
  `docs/evidence/ISSUE-030/`.
