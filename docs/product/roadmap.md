# EcoMatrix Roadmap

> Source of truth for what we ship, in what order, by when. Mirrored to `PROJECT_STATUS.md`.

## Phases

### Phase 1 — Physical Engine (Week 1) → tag `v0.1.0`

**Goal:** Go backend + Postgres operational; trade atomicity proven under concurrency.

- Monorepo scaffold (this bootstrap).
- Backend skeleton: Fiber + GORM + Postgres + WebSocket hub + slog.
- Migrations: `agents`, `transactions`, `social_feeds`.
- A2A protocol codec (`pkg/a2a`).
- Agent CRUD: `GET /v1/agents`, `GET /v1/agents/{id}`, `POST /v1/agents` (admin).
- Trade API: `POST /v1/trades` with `SELECT ... FOR UPDATE` row lock + idempotency by `msg_id`.
- Concurrency test: 50 goroutines racing to overspend — invariant holds.
- `/healthz`, `/readyz`.
- Dev runner (Makefile + `docker compose up` for Postgres).
- Evidence: backend `change-summary.md` + race-test output + curl traces.

Exit criteria: `go test -race ./...` green; curl trace shows concurrent overspend rejected.

### Phase 2 — Brain Onboarded (Week 2) → tag `v0.2.0`

**Goal:** Python LangGraph agent generates strategies and calls the Go trade API end-to-end.

- Python agent skeleton (`apps/agent`).
- LangGraph state machine per agent job type (miner / merchant / hacker / mediator).
- LLM provider abstraction + OpenAI-compatible default.
- A2A client in Python (`apps/agent/ecomatrix/a2a.py`), shape-matching the Go codec.
- Memory: short-term in graph state; long-term summaries in `agents.long_term_memory` (added by migration).
- Two-agent scenario: `agent_miner_01` ↔ `agent_merchant_01` execute one trade; balance deltas match the ledger.
- Seeded agents drive themselves for ≥ 60 s without crash.
- Evidence: scenario test output + sample agent CoT trace.

Exit criteria: two-agent trade loop runs headlessly; ledger is consistent with strategy log.

### Phase 3 — God's Eye (Week 3) → tag `v0.3.0`

**Goal:** Next.js dashboard renders live economy from WebSocket.

- Next.js (App Router) + Tailwind + Aceternity UI + Framer Motion.
- Routes: `/` (dashboard), `/agents/[id]` (detail).
- WS client with reconnect + value damping.
- KPI tiles (live agent count, total wealth, peak QPS) backed by `GET /v1/metrics`.
- Social feed timeline (Aceternity `Tracing Beam`).
- Agent detail panel with CoT viewer.
- Playwright screenshots desktop + mobile.
- i18n: `zh-CN` default.
- Evidence: screenshots + Playwright JSON + UI reviewer report.

Exit criteria: dashboard renders ≥ 3 KPIs ticking live; opening an agent shows CoT trace.

### Post-MVP Enhancements — Phase 4–7

- Phase 4: time-series metrics, Postgres-backed HMAC secrets, and final UX evidence.
- Phase 5: OpenAI-compatible LLM provider, tool calls, conversations, cache, and failure handling.
- Phase 6: agent contracts, bounded ReAct workflow, traces, cost controls, memory compression, and golden evals.
- Phase 7: live dashboard and agent-detail demo captures.

These enhancements are shipped in commits `56680b1` through `450262b`; their
evidence is under `docs/evidence/PHASE-{4-final,5-ai,6-ai,7-ai}/`.

### Phase 8 — Hierarchical Supervisor

**Goal:** replace flat-only orchestration with an optional supervisor that can
decompose a high-level goal and delegate bounded work to specialized agents.

- Validate a non-empty goal and available worker registry.
- Ask the configured LLM for at most four well-shaped subtasks.
- Route each subtask by agent ID or job type and execute it via `run_react`.
- Aggregate worker receipts and errors into a final summary.
- Expose the workflow through `--scenario supervisor --goal ...`.
- Keep existing `two_agent` and `multi` scenarios unchanged.
- Evidence: full Python tests, touched-file Ruff, and two reviewer reports.

Exit criteria: the supervisor scenario is CLI-reachable, bounded, tolerant of
malformed/failing LLM responses, and the complete Python test suite is green.

## Cross-cutting (every phase)

- Adversarial review (≥ 2 reviewers per PR).
- Evidence pack per Issue.
- `docs/INDEX.md` and `memory/` updated at end of phase.
- ADR per non-obvious decision.
