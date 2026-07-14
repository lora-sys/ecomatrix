# ISS-030 Verification

## Live `curl` probes

```
$ curl -sS http://localhost:8080/v1/metrics
{"agent_count":12,"total_gold":1660,"jobs_breakdown":{"hacker":2,"mediator":1,
 "merchant":3,"miner":6},"recent_qps":0,"ws_connections":0,
 "supervisor_runs_count":2,
 "supervisor_last_run_at":"2026-07-13T13:42:39Z",
 "generated_at":"..."}

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

- `cd apps/backend && go test -race -count=1 ./...` — all packages green
  with the new schema-aware supervisor http suite and the new repo
  `wipeRepo` extension.
- `cd apps/agent && uv run pytest -q` — **105 passed** (4 new tests in
  `test_supervisor_detail.py`).
- `cd apps/frontend && npx tsc --noEmit && npm run lint && npm run build`
  — clean; `/supervisor/[id]` First Load 149 kB.

## CI (PR #4 run `29300898334`)

| Job                          | Status | Duration |
| ---------------------------- | ------ | -------- |
| `agent (Python pytest)`      | pass   | ~11 s    |
| `backend (Go -race)`         | pass   | ~34 s    |
| `frontend (Next.js + Playwright)` | pass | ~62 s   |
| `e2e (Playwright)`           | pass   | ~2 m 37 s |

All four required CI jobs are green and the PR is `MERGEABLE`.

## Playwright (desktop + mobile)

| Spec                                                | Tests | Status                                  |
| --------------------------------------------------- | ----- | --------------------------------------- |
| `dashboard.spec.ts`                                 | 5     | all green on both viewports             |
| `supervisor.spec.ts` (ISS-030)                      | 2     | green on both viewports; seeds a run via `POST /v1/supervisor/runs` before exercising the dashboard link and the agent-page section. |
| `ai-thought-screenshot.spec.ts`                     | 1     | skipped (manual repro)                  |
| `demo-video.spec.ts`                                | 1     | green on both viewports                 |
| `final-video.spec.ts`                               | 1     | green on both viewports                 |

Screenshots (desktop + mobile) saved at:

- `docs/evidence/ISSUE-030/supervisor-detail-desktop.png`
- `docs/evidence/ISSUE-030/supervisor-detail-mobile.png`
- `docs/evidence/ISSUE-030/agent-supervisor-history-desktop.png`
- `docs/evidence/ISSUE-030/agent-supervisor-history-mobile.png`

## PR

- Branch: `feature/ISSUE-030-supervisor-detail`
- PR: https://github.com/lora-sys/ecomatrix/pull/4 (run `29300898334`)
