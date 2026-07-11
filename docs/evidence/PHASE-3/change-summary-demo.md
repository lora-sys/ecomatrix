# Phase 3.5 — One-shot demo onboarding

## What
Added a single command that brings up the entire stack end-to-end:
- Postgres (via docker compose, reuses an existing `ecomatrix-postgres` container if present).
- Backend migrations + seed (11 deterministic agents).
- Go backend on `:8080`.
- Next.js dashboard on `:3100`.
- Python multi-agent scenario in the background, ticking continuously.

## Why
The harness has produced 9 commits and a fully runnable system, but until now a new contributor had to read four READMEs and start four terminals. `make demo` collapses that to one command.

## Files

```
Makefile                              # NEW: root orchestrator with help/db/seed/backend/frontend/agent/demo/test
scripts/demo.sh                       # NEW: bring-everything-up; SIGINT-safe
README.md                             # rewritten: "Quickstart — one command" header + manual fallback
docs/evidence/PHASE-3/demo/demo-smoke-output.txt   # NEW: end-to-end run record
docs/evidence/PHASE-3/screenshots/dashboard-desktop.png  # live QPS=33 during burst
```

## Verified (smoke run)

```
[1/5] Postgres: /var/run/postgresql:5432 - accepting connections
[2/5] building backend...
{"time":"...","level":"INFO","msg":"seed complete","agents":11}
[3/5] starting backend... backend ready after 2s
[4/5] starting frontend... frontend ready after 7s
[5/5] running multi-agent for 5 ticks...
{
  "agents": 13,
  "ticks": 5,
  "settled": 61,
  "rejected": 4,
  "feeds_posted": 50,
  "errors": [],
  "world_initial": 2560,
  "world_final": 2560,
  "conservation": true
}
```

Backend `/v1/metrics` after the burst:

```json
{
  "agent_count": 13,
  "total_gold": 2560,
  "recent_qps": 33,
  "ws_connections": 1,
  "last_trade_at": "2026-07-11T07:21:15.883198238Z"
}
```

Playwright 4/4 green on desktop + mobile against the running stack. Screenshot at `docs/evidence/PHASE-3/screenshots/dashboard-desktop.png` shows the dashboard with **QPS: 33.00** and a populated social square, captured during the multi-agent scenario.

## How a new contributor uses it

```bash
git clone <this repo>
cd ecomatrix
make demo
# → open http://localhost:3100
# → Ctrl-C to stop
```

`make test` runs the full suite (Go `-race`, agent pytest, frontend tsc + Playwright).
