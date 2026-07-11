# Phase 1 — Change Summary

## What
EcoMatrix backend is operational. A single Go binary serves the A2A v1.1 trade protocol, an admin-protected agent CRUD surface, a real-time WebSocket hub, and liveness/readiness probes — all backed by Postgres with row-level locks preventing double-spend.

## Why
PRD §7 calls for the **Physical Engine** in Week 1: a backend that can absorb thousands of concurrent A2A trades without double-spending, expose state changes live to the dashboard, and survive restart via a deterministic seed.

## Notable Decisions
- **Fiber v2 + GORM + raw `SELECT … FOR UPDATE` in LockPair** — GORM's `gorm:query_option` does not reliably append `FOR UPDATE`; raw SQL inside the tx is the only safe choice. (See `internal/repo/agent_repo.go::LockPair`.)
- **Idempotency at the DB layer** — `transactions.msg_id` carries a `UNIQUE` constraint; the service treats a duplicate as a settled replay or a rejection replay (no side-effects).
- **Lock order: ascending id** — every state-mutating endpoint that touches two agent rows locks them in the same order to avoid deadlock under contention. The 50-goroutine test exercises this.
- **WebSocket backpressure** — per-connection buffered channel; slow consumers are dropped, the publisher never blocks.
- **Money as `BIGINT` + CHECK constraint** — defense in depth: even if the service-level guard were bypassed, the DB would refuse to go negative.

## Out of Scope
- Phase 2 (Python LangGraph agent) and Phase 3 (Next.js dashboard) — both have issue drafts in `tasks/`.
- Per-agent HMAC tokens (still using shared admin token in dev).
- `/v1/metrics` Prometheus endpoint.

## Files Shipped (Phase 1)

```
apps/backend/
├── cmd/server/main.go
├── cmd/seed/main.go
├── pkg/a2a/{envelope,codec,errors,codec_test}.go
├── internal/
│   ├── config/config.go
│   ├── domain/{agent,transaction,errors}.go
│   ├── repo/{db,agent_repo,tx_repo,tx_repo_list}.go
│   ├── repo/migrations_fs/0001_init.{up,down}.sql
│   ├── service/trade.go + trade_test.go
│   ├── transport/http/router.go
│   ├── transport/ws/hub.go
│   └── observability/log.go
├── migrations/0001_init.{up,down}.sql
├── Makefile
├── README.md
└── .env.example
```

Plus `docker-compose.yml`, `docs/evidence/PHASE-1/*`, `docs/decisions/0001-stack-and-architecture.md`, and the harness operating system (`CLAUDE.md`, `AGENTS.md`, `DESIGN.md`, `ENGINEERING.md`, `TESTING.md`, `CONTRIBUTING.md`, `PROJECT_STATUS.md`, `docs/INDEX.md`, `tasks/ISSUE-002..011.md`).
