# Changelog

All notable changes to EcoMatrix are recorded here. Format: [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

## [Unreleased]

### Phase 3 — God's Eye (queued)
- Next.js dashboard with Aceternity UI
- WebSocket-driven KPI tiles, trade broadcast, agent detail

## [0.2.0] — 2026-07-11 — Phase 2 Brain Onboarded

### Added
- Python 3.12 agent runtime at `apps/agent` (uv-managed venv).
- A2A v1.1 client (`ecomatrix/a2a.py`) — Python mirror of `pkg/a2a`; 8 parity tests.
- LLM provider abstraction: `StubLLM` (deterministic, default) + `OpenAICompatibleLLM`.
- LangGraph state machines for miner, merchant, hacker, mediator (observe → think → act).
- Short-term memory (graph state) + file-backed long-term memory.
- Runner CLI: `--scenario two_agent` drives two agents for N ticks; emits a JSON summary with world-GOLD conservation check.
- New Go endpoint `GET /v1/agents/by-string-id/{sid}` so Python agents can resolve by `string_id`.
- 14 pytest tests pass; ruff clean.

### Verified
- 5-tick scenario: 10 settled, 0 rejected, 0 errors, world 2560 → 2560.
- 10-tick scenario: 20 settled, 0 rejected, 0 errors, world 2560 → 2560 (miner drains to 0).

[0.2.0]: https://github.com/ecomatrix/ecomatrix/compare/v0.1.0...v0.2.0

## [0.1.0] — 2026-07-11 — Phase 1 Physical Engine

## [0.1.0] — 2026-07-11 — Phase 1 Physical Engine

### Added
- Go backend (Fiber v2) on `:8080` with structured `slog` JSON logging.
- A2A v1.1 protocol codec (`pkg/a2a`) with 8 unit tests.
- Postgres migrations runner with embedded SQL files.
- Schema: `agents`, `transactions`, `social_feeds`, `schema_migrations`.
- Endpoints:
  - `GET /healthz`, `GET /readyz`
  - `GET /v1/agents`, `GET /v1/agents/{id}`
  - `POST /v1/agents` (admin-gated)
  - `POST /v1/trades` (A2A `EXECUTE_TRADE`)
  - `GET /v1/transactions`
  - `GET /v1/stream` (WebSocket fan-out)
- Trade service with `SELECT ... FOR UPDATE` row-level locks (ascending id order to avoid deadlock).
- Idempotency by `msg_id` (UNIQUE constraint + service-level short-circuit).
- WebSocket hub with per-connection buffered channel + 20s heartbeat + slow-consumer drop.
- 50-goroutine concurrency test: settled 33 / rejected 17, balance never negative.
- Seed script (11 deterministic agents).
- Docker Compose for local Postgres.
- Make targets: `build`, `test`, `test-race`, `vet`, `tidy`, `run`, `seed`, `db-up`, `db-down`, `lint`, `fmt`.
- CI workflow (`.github/workflows/ci.yml`).
- Harness operating system: `CLAUDE.md`, `AGENTS.md`, `DESIGN.md`, `ENGINEERING.md`, `TESTING.md`, `CONTRIBUTING.md`, `PROJECT_STATUS.md`, `docs/INDEX.md`, ADRs, evidence pack.

### Security
- Shared `ECOMATRIX_ADMIN_TOKEN` for admin endpoints (dev only).
- A2A payload `reasoning` not logged; balances above 10 000 masked.

[Unreleased]: https://github.com/ecomatrix/ecomatrix/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/ecomatrix/ecomatrix/releases/tag/v0.1.0
