# Changelog

All notable changes to EcoMatrix are recorded here. Format: [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

## [0.3.0] — 2026-07-11 — Phase 3 God's Eye

### Added
- Next.js 15 (App Router) + React 19 + Tailwind + Framer Motion + zustand dashboard at `apps/frontend`.
- Routes: `/` (dashboard with KPI tiles, wealth chart, trade broadcast, agent list, job-type 3D cards), `/agents/[id]` (vitals + recent trades via tracing-beam).
- `LiveProvider` client: WebSocket reconnect with exponential backoff, periodic `/v1/metrics` polling, value damping for KPI tiles.
- Aceternity-style local components: `GlowingCard`, `TracingBeam`, `ThreeDCard`.
- i18n strings (`zh-CN` default) — pre-wired, swap-in for English via `next-intl`.
- Playwright e2e covering dashboard and agent detail on desktop + mobile.
- Go: `GET /v1/metrics` (`MetricsService`); CORS middleware for dashboard.
- Metrics tests: empty DB, jobs breakdown, QPS window.

### Verified
- Playwright 4/4 green on desktop + mobile.
- Next.js first-load JS: 151 KB (budget 250 KB).
- Backend `/v1/metrics` shows live `recent_qps` and `last_trade_at` after a trade burst.

[0.3.0]: https://github.com/ecomatrix/ecomatrix/compare/v0.2.0...v0.3.0

## [0.3.1] — 2026-07-11 — Social Square + Multi-Agent Scenario

### Added
- A2A `POST_FEED` action + `POST /v1/feeds` + `GET /v1/feeds?limit=N`.
- `FeedRepo` (GORM) + 3 codec tests.
- Python `A2AClient.post_feed()` / `list_feeds()` with 4 parity tests.
- `--scenario multi` runner: spawns every seeded agent in parallel (ThreadPoolExecutor), drives N ticks, asserts world-GOLD conservation.
- Graph `act` node emits a feed post every tick (intent = SOCIAL/OFFER/REQUEST depending on decision).
- Dashboard `SocialFeed` panel + BFF proxy `/api/proxy/feeds` and `/api/proxy/metrics`.

### Verified
- Multi-scenario (3 ticks, 13 agents): 37 settled, 39 feed posts, 0 errors, world 2560 → 2560.
- `/v1/metrics` QPS rises from 0 to 16 during the burst.
- Playwright 4/4 still green after dashboard layout change.
- Dashboard desktop screenshot shows mixed [REQUEST]/[SOCIAL]/[OFFER] posts in the new panel.

[0.3.1]: https://github.com/ecomatrix/ecomatrix/compare/v0.3.0...v0.3.1

## [0.3.2] — 2026-07-11 — Long-Term Memory

### Added
- Migration `0002_agents_long_term_memory`: JSONB column with GIN index on `agents`.
- `LongTermMemory` domain struct + `AgentRepo.GetLongTermMemory` / `SetLongTermMemory`.
- Endpoints: `GET / PUT /v1/agents/by-string-id/{sid}/long-term-memory`.
- Python `PostgresLongTermMemory` (HTTP-backed); `FileLongTermMemory` kept as the offline default.
- Runner uses Postgres LTM by default; switch with `ECOMATRIX_AGENT_LTM=file`.
- Dashboard agent detail page renders an LTM card with summary + last 8 facts.

### Verified
- 50-goroutine concurrency test still green after the column addition.
- 19 Python tests pass; 21 Go tests pass.
- Multi-scenario (3 ticks) ends with conservation intact and 3 new `settled tx_…` facts appended to `agent_miner_01`'s LTM.

[0.3.2]: https://github.com/ecomatrix/ecomatrix/compare/v0.3.1...v0.3.2

## [Unreleased]

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
