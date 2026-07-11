# PROJECT_STATUS.md — Live Board

> Source of truth for "what's next". Update this when you start/finish work.

## Legend
- `Todo` — accepted, not started.
- `Planning` — being scoped into an Issue.
- `Implementing` — Issue claimed, branch active.
- `Review` — PR open, reviewers running.
- `Testing` — QA + Evidence collection.
- `Blocked` — needs human input (see note).
- `Done` — merged, evidence complete.

## Phase 1 — Physical Engine (Week 1)

| ID    | Title                                                | Owner     | Status      | Branch                          | Notes |
| ----- | ---------------------------------------------------- | --------- | ----------- | ------------------------------- | ----- |
| ISS-001 | Monorepo + Go backend skeleton                      | backend   | Done        | `feature/ISSUE-001-monorepo`    | This bootstrap. |
| ISS-002 | A2A protocol codec + error envelope                 | backend   | Done        | `feature/ISSUE-002-a2a-codec`   | Codec + 8 unit tests. |
| ISS-003 | Postgres connection + migrations runner             | database  | Done        | `feature/ISSUE-003-db-migrations` | Raw-SQL migrations runner. |
| ISS-004 | DB schema (agents, transactions, social_feeds)      | database  | Done        | `feature/ISSUE-004-db-schema`   | Embedded migrations + seed. |
| ISS-005 | Agent CRUD endpoints                                | backend   | Done        | `feature/ISSUE-005-agent-crud`  | Admin-gated create + GET list. |
| ISS-006 | Trade API with row-level lock + idempotency         | backend   | Done        | `feature/ISSUE-006-trade-api`   | Row lock + DB-level idempotency. |
| ISS-007 | Concurrency test (50 goroutines racing)             | qa        | Done        | `feature/ISSUE-007-concurrency` | 33/17 split, invariant holds. |
| ISS-008 | WebSocket hub + `/v1/stream`                       | backend   | Done        | `feature/ISSUE-008-ws-hub`      | Backpressure + heartbeat. |
| ISS-009 | Health/Readiness + structured logging               | backend   | Done        | `feature/ISSUE-009-observability` | One JSON line per request. |
| ISS-010 | Seed script + Make targets + docker compose         | backend   | Done        | `feature/ISSUE-010-dev-ergonomics` | Make + compose + seed. |
| ISS-011 | Evidence pack + Phase 1 release prep               | release   | Done        | `feature/ISSUE-011-release-v0.1.0` | `docs/evidence/PHASE-1/`. |

## Phase 2 — Brain Onboarded (Week 2)

| ID    | Title | Owner | Status |
| ----- | ----- | ----- | ------ |
| ISS-012 | Python agent skeleton + LLM provider abstraction | agent | Todo |
| ISS-013 | A2A client in Python (codec parity with Go)      | agent | Todo |
| ISS-014 | LangGraph state machines per job type             | agent | Todo |
| ISS-015 | Long-term memory migration + repo                | database | Todo |
| ISS-016 | Two-agent end-to-end scenario (miner↔merchant)   | qa | Todo |
| ISS-017 | Phase 2 evidence + `v0.2.0` tag                  | release | Todo |

## Phase 3 — God's Eye (Week 3)

| ID    | Title | Owner | Status |
| ----- | ----- | ----- | ------ |
| ISS-018 | Next.js 15 + Tailwind + Aceternity scaffold      | frontend | Todo |
| ISS-019 | WS client hook + zustand store + value damping   | frontend | Todo |
| ISS-020 | KPI tiles + wealth chart + trade broadcast       | frontend | Todo |
| ISS-021 | Agent detail panel + CoT trace viewer            | frontend | Todo |
| ISS-022 | Playwright tests + desktop/mobile screenshots    | qa | Todo |
| ISS-023 | Phase 3 evidence + `v0.3.0` tag                  | release | Todo |

## Blockers
(none)

## Decisions Pending Human Input
(none)

## Recent Changes
- 2026-07-11 — Phase 1 shipped: Go backend live, 14 tests pass under `-race`, 50-goroutine concurrency proof recorded.
- 2026-07-11 — Bootstrap complete: Source-of-Truth docs, harness operating system, Phase 1–3 issue drafts.
