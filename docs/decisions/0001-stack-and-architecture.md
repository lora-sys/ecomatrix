# ADR-0001: Stack and Architecture Selection

- **Status:** Accepted
- **Date:** 2026-07-11
- **Deciders:** Coordinator (Codex)

## Context
PRD requires: high-concurrency trade sync, real-time front-end rendering, AI agents with strategy generation. Constraints: no double-spend, fast iteration, single-team scope.

## Decision
- **Backend:** Go 1.26 + Fiber v2 + GORM + `gofiber/contrib/websocket`.
- **Database:** PostgreSQL 16 with row-level locks (`SELECT ... FOR UPDATE`) for trade atomicity.
- **Frontend:** Next.js 15 App Router + Aceternity UI + Tailwind + Framer Motion.
- **AI:** Python 3.12 + LangGraph + pluggable LLM provider.

## Consequences
Positive:
- Single source of truth (Postgres) makes double-spend impossible by construction.
- Go goroutines handle thousands of concurrent A2A calls with low memory.
- RSC-first Next.js minimizes client JS for the dashboard.

Negative:
- Three-language monorepo increases onboarding.
- Aceternity UI upstream changes require wrapping discipline.

Neutral:
- Vendor lock-in is low; the protocol (A2A) is the contract, not the stack.

## Alternatives Considered
- **Event-sourced ledger (Kafka + projections):** overkill for MVP; deferred.
- **SQLite + Litestream:** insufficient for true row-level locking under load.
- **Single Node.js + Fastify backend:** rejected to keep AI/ledger responsibilities split and to use Go's concurrency model for the hot path.

## References
- `docs/product/prd.md` §2
- `docs/architecture/system.md`
