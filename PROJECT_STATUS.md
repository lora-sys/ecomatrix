# PROJECT_STATUS.md — Live Board

## Phase 1 — Physical Engine (Week 1)

| ID    | Title                                                | Owner     | Status      | Notes |
| ----- | ---------------------------------------------------- | --------- | ----------- | ----- |
| ISS-001 | Monorepo + Go backend skeleton                      | backend   | Done        | Bootstrap. |
| ISS-002 | A2A protocol codec + error envelope                 | backend   | Done        | Codec + 8 unit tests. |
| ISS-003 | Postgres connection + migrations runner             | database  | Done        | Raw-SQL migrations runner. |
| ISS-004 | DB schema (agents, transactions, social_feeds)      | database  | Done        | Embedded migrations + seed. |
| ISS-005 | Agent CRUD endpoints                                | backend   | Done        | + GET by-string-id added in Phase 2. |
| ISS-006 | Trade API with row-level lock + idempotency         | backend   | Done        | Row lock + DB-level idempotency. |
| ISS-007 | Concurrency test (50 goroutines racing)             | qa        | Done        | 33/17 split, invariant holds. |
| ISS-008 | WebSocket hub + `/v1/stream`                       | backend   | Done        | Backpressure + heartbeat. |
| ISS-009 | Health/Readiness + structured logging               | backend   | Done        | One JSON line per request. |
| ISS-010 | Seed script + Make targets + docker compose         | backend   | Done        | Make + compose + seed. |
| ISS-011 | Evidence pack + Phase 1 release prep               | release   | Done        | `docs/evidence/PHASE-1/`. |

## Phase 2 — Brain Onboarded (Week 2)

| ID    | Title | Owner | Status | Notes |
| ----- | ----- | ----- | ------ | ----- |
| ISS-012 | Python agent skeleton + LLM provider abstraction | agent | Done | uv venv + ruff clean. |
| ISS-013 | A2A client in Python (codec parity with Go)      | agent | Done | 8 parity tests. |
| ISS-014 | LangGraph state machines per job type             | agent | Done | miner / merchant / hacker / mediator graphs. |
| ISS-015 | Long-term memory migration + repo                | database | Done | Migration 0002 + PostgresLongTermMemory; dashboard renders. |
| ISS-016 | Two-agent end-to-end scenario (miner↔merchant)   | qa | Done | 5 + 10 tick runs; world conservation holds. |
| ISS-017 | Phase 2 evidence + `v0.2.0` tag                  | release | Done |
| ISS-024 | Social square: POST /v1/feeds + A2A POST_FEED    | backend | Done | Issue 024/025 collapsed into the multi-scenario. |
| ISS-025 | --scenario multi (parallel agents)              | agent   | Done | |
| ISS-026 | Dashboard social-feed panel + BFF proxy         | frontend| Done | |
| ISS-027 | Dashboard agent LTM panel + GET/PUT endpoints   | frontend/backend | Done | Phase 3.2 close-out. |

## Phase 3 — God's Eye (Week 3)

| ID    | Title | Owner | Status |
| ----- | ----- | ----- | ------ |
| ISS-018 | Next.js 15 + Tailwind + Aceternity scaffold      | frontend | Done        | apps/frontend scaffolded on port 3100. |
| ISS-019 | WS client hook + zustand store + value damping   | frontend | Done        | LiveProvider + reconnect backoff. |
| ISS-020 | KPI tiles + wealth chart + trade broadcast       | frontend | Done        | 4 tiles + wealth chart + feed + 3D cards. |
| ISS-021 | Agent detail panel + CoT trace viewer            | frontend | Done        | `/agents/[id]` with tracing-beam timeline. |
| ISS-022 | Playwright tests + desktop/mobile screenshots    | qa | Done        | 4/4 passing; screenshots in evidence. |
| ISS-023 | Phase 3 evidence + `v0.3.0` tag                  | release | Done        | docs/evidence/PHASE-3/. |

## Post-MVP Enhancements

| ID | Title | Owner | Status | Notes |
| -- | ----- | ----- | ------ | ----- |
| PHASE-4 | Metrics history + persistent HMAC + UX evidence | cross-cut | Done | `docs/evidence/PHASE-4-final/`. |
| PHASE-5 | LLM provider + tools + conversations + cache | cross-cut | Done | `docs/evidence/PHASE-5-ai/`. |
| PHASE-6 | Production agent contracts + ReAct + eval + traces | agent/backend | Done | `docs/evidence/PHASE-6-ai/`. |
| PHASE-7 | Live dashboard demo capture | qa | Done | `docs/evidence/PHASE-7-ai/`. |

## Phase 8 — Hierarchical Supervisor

| ID | Title | Owner | Status | Notes |
| -- | ----- | ----- | ------ | ----- |
| ISS-028 | Bounded supervisor workflow + CLI scenario | agent | Blocked | PR #2 open; first CI run red on baseline gofmt/ESLint gates, recovery fix prepared locally. |

## Blockers
- PR #2 CI run `29218912703` is red until the gofmt/ESLint recovery commit is pushed and rerun.

## Decisions Pending Human Input
(none)

## Recent Changes
- 2026-07-13 — Created private `lora-sys/ecomatrix`, Issue #1, and Draft PR #2; first CI run exposed baseline gofmt and interactive ESLint failures.
- 2026-07-13 — ISS-028 prepared as a local feature commit; authenticated GitHub account has no matching `ecomatrix` repository, so the board remains in Review.
- 2026-07-13 — ISS-028 implementation complete locally: 101 Python tests, Go race suite, Ruff, live seeded CLI, and two reviewer passes green; moved to Review pending PR/CI.
- 2026-07-13 — ISS-028 claimed: recovered the unfinished supervisor prototype; baseline is 80 Python tests passing and 3 supervisor tests failing.
- 2026-07-12 — Phase 6 shipped production agent contracts, ReAct, traces, cost controls, memory compression, and golden evals; Phase 7 captured live demo evidence.
- 2026-07-12 — Phase 4–5 shipped metrics history, persistent HMAC secrets, real LLM plumbing, tool calls, conversations, and cache.
- 2026-07-11 — Phase 3.1: social square (POST_FEED) + --scenario multi (13 agents, 3 ticks, 37 settled, 39 feed posts, world conservation holds).
- 2026-07-11 — Phase 2 shipped: Python LangGraph agent drives the Go backend end-to-end; 14 unit/integration tests + 5/10-tick two-agent scenarios pass with world-GOLD conservation.
- 2026-07-11 — Phase 1 shipped: Go backend live, 14 tests pass under `-race`, 50-goroutine concurrency proof recorded.
- 2026-07-11 — Bootstrap complete: Source-of-Truth docs, harness operating system, Phase 1–3 issue drafts.
