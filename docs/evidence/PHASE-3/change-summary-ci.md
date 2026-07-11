# Phase 3.8 — CI Workflow Updated

## What
`.github/workflows/ci.yml` was stale (referenced `pnpm` while the frontend uses `npm`; skipped the auth tests; missed Playwright; used `pip` directly instead of `uv`). Brought it in line with the actual repo: 4 jobs (`backend`, `agent`, `frontend`, `e2e`) all matching the local commands documented in `Makefile`.

## Why
CI is the durable evidence that the harness works for any new contributor. After 12 commits with HMAC, CORS, LTM, social square, recharts, and multi-agent, the workflow needed to reflect the current shape of the three apps.

## Files

```
.github/workflows/ci.yml                   # rewrite: 4 jobs, parallel where possible
docs/evidence/PHASE-3/change-summary-ci.md # NEW
apps/backend/                              # gofmt -s applied to 5 files flagged by `gofmt -l`
```

## Verified locally (the same commands CI will run)

| Job       | Command                                              | Result |
| --------- | ---------------------------------------------------- | ------ |
| backend   | `go vet ./...`                                        | clean  |
| backend   | `gofmt -l .`                                          | empty  |
| backend   | `go test -race -count=1 ./...`                        | 36 tests, all pass under `-race` |
| agent     | `uv run pytest -q`                                    | 23 passed in 0.24s |
| frontend  | `npx tsc --noEmit`                                    | clean  |
| frontend  | `npm run build`                                       | 253 KB First-Load JS (within budget) |

The `e2e` job runs Playwright against a real Postgres + backend + frontend; the same flow that `make demo` exercises locally.
