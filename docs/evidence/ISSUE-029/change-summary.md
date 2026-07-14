# ISS-029 Change Summary

## Shipped

- Backend persistence: migration `0007_supervisor_runs`, GORM
  `SupervisorRunsRepo`, and `GET`/`POST /v1/supervisor/runs` endpoints.
- Python client helper `A2AClient.post_supervisor_run` /
  `list_supervisor_runs`; `run_supervisor_scenario` forwards every run to
  the backend (best-effort).
- Dashboard `<SupervisorLog>` with SSR hydration and live updates; merges
  initial runs and live deltas through the store.
- E2E: dashboard test asserts the supervisor log live region is visible on
  first paint.

## Verification

- Local: `go test -race -count=1 ./...` green, Python 101/101, frontend
  `tsc`/`lint`/`build` clean.
- Live smoke (post-merge): pending PR run; will be appended to
  `verification.md` after CI.
