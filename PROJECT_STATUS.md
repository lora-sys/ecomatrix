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
| ISS-015 | Long-term memory migration + repo                | database | Todo | File-backed LTM in use; Postgres JSONB column is the follow-up. |
| ISS-016 | Two-agent end-to-end scenario (miner↔merchant)   | qa | Done | 5 + 10 tick runs; world conservation holds. |
| ISS-017 | Phase 2 evidence + `v0.2.0` tag                  | release | Done | `docs/evidence/PHASE-2/`. |

## Phase 3 — God's Eye (Week 3)

| ID    | Title | Owner | Status |
| ----- | ----- | ----- | ------ |
| ISS-018 | Next.js 15 + Tailwind + Aceternity scaffold      | frontend | Done        | apps/frontend scaffolded on port 3100. |
| ISS-019 | WS client hook + zustand store + value damping   | frontend | Done        | LiveProvider + reconnect backoff. |
| ISS-020 | KPI tiles + wealth chart + trade broadcast       | frontend | Done        | 4 tiles + wealth chart + feed + 3D cards. |
| ISS-021 | Agent detail panel + CoT trace viewer            | frontend | Done        | `/agents/[id]` with tracing-beam timeline. |
| ISS-022 | Playwright tests + desktop/mobile screenshots    | qa | Done        | 4/4 passing; screenshots in evidence. |
| ISS-023 | Phase 3 evidence + `v0.3.0` tag                  | release | Done        | docs/evidence/PHASE-3/. |

## Blockers
(none)

## Decisions Pending Human Input
(none)

## Recent Changes
- 2026-07-11 — Phase 2 shipped: Python LangGraph agent drives the Go backend end-to-end; 14 unit/integration tests + 5/10-tick two-agent scenarios pass with world-GOLD conservation.
- 2026-07-11 — Phase 1 shipped: Go backend live, 14 tests pass under `-race`, 50-goroutine concurrency proof recorded.
- 2026-07-11 — Bootstrap complete: Source-of-Truth docs, harness operating system, Phase 1–3 issue drafts.
