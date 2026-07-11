# Project Memory

Stable facts about EcoMatrix that survive sessions.

## Identity
- Project: EcoMatrix — multi-agent economic sandbox.
- Stack: Go (Fiber) + Next.js (Aceternity) + Python (LangGraph) + Postgres.
- Protocol: A2A v1.1 (see `docs/architecture/api.md`).

## Constraints
- Dev DB: `postgres://repotwin:repotwin@localhost:5432/ecomatrix` (Docker `repotwin-postgres-1`).
- All monetary values are integer `BIGINT` (satoshis-style). Currency enum: `GOLD` (only one in MVP).
- Agents are identified externally by `string_id` (`agent_miner_01`); internally by `BIGSERIAL id`.

## Job Types
`miner`, `merchant`, `hacker`, `mediator`. New job types require an ADR.

## Key Invariants
- `agents.balance >= 0` always.
- `transactions.amount > 0` and `from_id <> to_id`.
- Trade idempotency key: `transactions.msg_id` (UNIQUE).
- Lock order in trade service: ascending `agents.id` to avoid deadlocks.

## Owner
- Coordinator: Codex (this harness).
- Human owner: <user>. Approval gates: schema migrations, auth changes, releases.

## Active Phase
Phase 1 — Physical Engine.
