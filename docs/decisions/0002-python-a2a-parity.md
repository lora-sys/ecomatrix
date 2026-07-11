# ADR-0002: Python A2A parity via shared tests, not shared code

- **Status:** Accepted
- **Date:** 2026-07-11
- **Deciders:** Coordinator (Codex)

## Context
Phase 2 needs a Python agent that speaks A2A v1.1 to the Go backend. Two options for keeping the codec in sync:
1. Generate both implementations from a single schema (e.g. JSON Schema + codegen).
2. Hand-write both and lock them with a parity test suite that mirrors the Go codec tests.

## Decision
Hand-write both. The Python codec (`apps/agent/ecomatrix/a2a.py`) mirrors the Go codec (`apps/backend/pkg/a2a`) shape-for-shape; `tests/test_a2a_codec.py` mirrors `pkg/a2a/codec_test.go` one-for-one.

## Consequences
Positive:
- No codegen step in CI; both languages stay readable.
- Drift is caught the moment a parity test fails.
- Tests are the contract; refactoring either side is safe.

Negative:
- Two implementations to maintain.
- A new field requires touching both sides and adding both tests.

Neutral:
- If Phase 2.5 or 3 adds a third language (TypeScript client), we should revisit and adopt codegen.

## Alternatives Considered
- **JSON Schema codegen** — adds a build step and a codegen tool; overkill for MVP.
- **WASM** of the Go codec loaded into Python — fragile, adds a runtime dep.
- **HTTP-only contract, no shared types** — leaves drift undetected until production.

## References
- `apps/agent/ecomatrix/a2a.py`
- `apps/backend/pkg/a2a/*.go`
- `docs/architecture/api.md`
