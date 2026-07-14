# ISS-030 Verification

## Live `curl` probes

```
$ curl -sS http://localhost:8080/v1/metrics
{"agent_count":12,"total_gold":1660,"jobs_breakdown":{"hacker":2,"mediator":1,
 "merchant":3,"miner":6},"recent_qps":0,"ws_connections":0,
 "supervisor_runs_count":2,
 "supervisor_last_run_at":"2026-07-13T13:42:39Z",
 "generated_at":"2026-07-14T01:29:44.475501072Z"}

$ curl -sS http://localhost:8080/v1/supervisor/runs/2 | jq .
{ "id": 2, "goal": "ISS-030 detail verification", "status": "finished",
  "subtasks": [...], "worker_results": [...], "warnings": [...],
  "tokens_used": 3499, "tokens_budget": 12000, "duration_ms": 36, ... }

$ curl -sS -o /dev/null -w "%{http_code}\n" \
    http://localhost:8080/v1/supervisor/runs/9999
404

$ curl -sS "http://localhost:8080/v1/agents/by-string-id/agent_miner_01/supervisor-runs?limit=2"
{ "runs": [ { "id": 1, "goal": "trade", "worker_results": [...], ... } ] }

$ curl -sS -o /dev/null -w "%{http_code}\n" \
    "http://localhost:8080/v1/agents/by-string-id/agent_ghost/supervisor-runs"
404
```

## Local Gates

- `cd apps/backend && go test -race -count=1 ./...` — all packages green,
  including the new repo+http+service tests.
- `cd apps/agent && uv run pytest -q` — **105 passed** (4 new
  tests in `test_supervisor_detail.py`, 101 prior).
- `cd apps/frontend && npm run typecheck && npm run lint && npm run build`
  — `tsc` clean, `next lint` clean, `next build` produces the
  `/supervisor/[id]` route at First Load 149 kB.

## Playwright (desktop + mobile)

| Spec                                                | Tests | Status                          |
| --------------------------------------------------- | ----- | ------------------------------- |
| `dashboard.spec.ts`                                 | 5     | all green                       |
| `supervisor.spec.ts` (new)                          | 2     | all green on both viewports     |
| `ai-thought-screenshot.spec.ts`                     | 1     | green (skipped if no backend)   |
| `demo-video.spec.ts`                                | 1     | green                           |
| `final-video.spec.ts`                               | 1     | green                           |
| Total                                               | 18 + 2 skipped                   | **18 passed**, 2 skipped        |

Screenshots (desktop + mobile) saved at:

- `docs/evidence/ISSUE-030/dashboard-desktop.png`
- `docs/evidence/ISSUE-030/dashboard-mobile.png`
- `docs/evidence/ISSUE-030/supervisor-detail-desktop.png`
- `docs/evidence/ISSUE-030/supervisor-detail-mobile.png`
- `docs/evidence/ISSUE-030/agent-supervisor-history-desktop.png`
- `docs/evidence/ISSUE-030/agent-supervisor-history-mobile.png`

## PR

- Branch: `feature/ISSUE-030-supervisor-detail`
- PR run to be filled in after CI; all four required jobs are green locally
  (agent pytest, backend Go -race, frontend typecheck/lint/build,
  Playwright e2e).
