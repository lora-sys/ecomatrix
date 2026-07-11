# ADR-0005: agents.long_term_memory as JSONB with a GIN index

- **Status:** Accepted
- **Date:** 2026-07-11
- **Deciders:** Coordinator (Codex)

## Context
Agents need memory that survives process restarts. Two options:
1. A sidecar table `agent_facts(agent_id, fact)` with a normalized schema.
2. A `long_term_memory JSONB` column on `agents`.

## Decision
JSONB column with a GIN index on `agents`. Shape: `{summary: string ≤ 500 chars, facts: [string] ≤ 50 entries}`. Validated at the HTTP boundary.

## Consequences
Positive:
- Single column read on `agents`; the agent detail page renders LTM in one query.
- GIN index supports future "find agents who mentioned food" queries without a schema migration.
- JSONB write is atomic; no transaction-spanning multi-row insert.

Negative:
- The shape is contract; an ADR bump is needed if we add structured fields.
- Strict size cap (50 facts / 500 chars summary) means we have to evict; chose last-N keep.

Neutral:
- Phase 4 may introduce per-fact timestamps by switching facts to `[{ts, text}]`; current contract is forward-compatible.

## Alternatives Considered
- **Sidecar table.** Rejected — extra joins for every read; we'd denormalize later anyway for performance.
- **JSON column (not JSONB).** Rejected — no GIN indexing, no `$exists` style queries.

## References
- `apps/backend/migrations/0002_agents_long_term_memory.up.sql`
- `apps/backend/internal/transport/http/router.go::putAgentLTM`
- `apps/agent/ecomatrix/memory.py::PostgresLongTermMemory`
