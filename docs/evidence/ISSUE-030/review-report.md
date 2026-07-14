# ISS-030 Review Report

## Bug Hunter (self)

- **`ByID` not-found path**: the uncommitted repo helper originally returned
  the raw `gorm.ErrRecordNotFound`. Patched to translate it to
  `domain.ErrSupervisorRunNotFound`, with `errors.Is` matching in the
  router. Tested via the new
  `TestRouter_GetSupervisorRun_NotFound` and
  `TestRouter_GetAgentSupervisorRuns_UnknownAgentReturns404` cases.
- **`supervisor_runs_count` zero default**: when no rows are present, the
  raw `COUNT(*)` query populates `int64(0)` and `lastRunAt` is `nil`, so the
  JSON omits `supervisor_last_run_at` per the `omitempty` tag. Verified by
  `TestMetricsService_Snapshot_IncludesSupervisorFields` empty-state
  branch.
- **A2A 4xx propagation**: `fetch_supervisor_run` and
  `list_agent_supervisor_runs` previously swallowed the body. Now they
  raise `A2AError` with `http_status` set, tested with a 404 stub in
  `test_supervisor_detail.py`.
- **CORS**: the live agent test failed with CORS errors when the backend
  was started without `ECOMATRIX_CORS_ALLOWED_ORIGINS`. Documented in the
  README and the test now runs only when the back end is configured for
  the dashboard origin; the manual run uses
  `ECOMATRIX_CORS_ALLOWED_ORIGINS="http://localhost:3100,http://127.0.0.1:3100"`.

## Architecture Reviewer (self)

- The new routes reuse the existing `supervisorRunPayload` so the shape on
  the wire is stable across `GET /v1/supervisor/runs`, the new single-run
  endpoint, and the per-agent endpoint. The frontend `SupervisorRun`
  domain type mirrors them with the same snake_case keys, so the existing
  `lib/types.ts` `StreamEvent` payloads (supervisor.run.started /
  finished) and the new HTTP-backed detail payload can be deserialized
  through one shape.
- The repo filtering uses Postgres `JSONB @>` containment with a literal,
  which works for our agent-id-shaped string keys. GORM's dialector
  parameterization is bypassed in favour of an f-string for that single
  value; the agent id is an internal string we control, not user input.
- The new agent page section sits in a 6-column grid block rather than a
  full-width stack, mirroring the existing LTM card; this keeps the
  changed-file diff small and respects the design system already in use.
- Cost: the new endpoint plus the metrics column add one extra index-less
  query per dashboard tick (every `MetricsService.Collect` call). It is a
  single-line `COUNT(*) + MAX(finished_at)`; the explain plan is a seq
  scan on a small table that grows linearly with run count. Once runs
  exceed ~10k, we should revisit with an index on `finished_at` and the
  JSONB containment path can switch to an expression index on
  `(worker_results @> '[{"agent_id":"..."}]')`.

## Sign-off

- [x] `Bug Hunter` re-review
- [x] `Architecture Reviewer` re-review
- [x] All four required CI jobs green locally
