# Phase 2 — Change Summary

## What
The Python LangGraph agent runtime is operational. The `apps/agent` service runs LangGraph state machines for each job type (miner, merchant, hacker, mediator), drives them via a deterministic stub LLM by default (or any OpenAI-compatible provider), and submits A2A v1.1 trades to the Go backend. The two-agent scenario (`agent_miner_01` ↔ `agent_merchant_01`) runs to completion with world-GOLD conservation intact.

## Why
PRD §7 calls for the **Brain** in Week 2: AI agents that autonomously decide when to trade, observe outcomes, and remember context. Without Phase 2, the Phase 3 dashboard has nothing living to visualize.

## Notable Decisions
- **A2A parity in Python** — `apps/agent/ecomatrix/a2a.py` mirrors `apps/backend/pkg/a2a` shape-for-shape. Parity is enforced by `tests/test_a2a_codec.py`, which mirrors the Go codec tests one-for-one.
- **Stub LLM by default** — `StubLLM` produces deterministic JSON so the scenario is reproducible offline. `OpenAICompatibleLLM` is the production path; flipped via `ECOMATRIX_AGENT_LLM_PROVIDER=openai`.
- **LangGraph state propagation** — `AgentState` is a TypedDict; the `act` node returns **deltas only** (e.g. `{"last_receipt": {...}}`), not the whole state, so LangGraph's reducer merges them correctly. (Caught during testing: returning the whole state caused `__result__` to be silently dropped.)
- **Conservation check at world level** — Two-agent net balance isn't conserved because each agent also trades with other agents. The runner sums all agent balances; world total GOLD is unchanged after any number of settled trades.
- **`GET /v1/agents/by-string-id/{sid}`** added to the Go backend in Phase 2 because the Python agent speaks `string_id` (`agent_miner_01`), not the internal numeric `id`.

## Out of Scope
- Long-term memory persistence in Postgres (the file-backed LTM store is wired; Postgres-backed `agents.long_term_memory` column migration is a small follow-up).
- A scenario with all 11 seeded agents running concurrently (the runner exposes `--scenario multi` as a stub for Phase 2 follow-up).
- Per-agent HMAC tokens.

## Files Shipped (Phase 2)

```
apps/agent/
├── pyproject.toml
├── README.md
├── .env.example
├── ecomatrix/
│   ├── __init__.py
│   ├── a2a.py            # Python mirror of pkg/a2a
│   ├── llm.py            # Stub + OpenAI-compatible providers
│   ├── memory.py         # Short-term + file-backed long-term memory
│   ├── runner.py         # Entry point: scenarios
│   └── graphs/
│       ├── __init__.py
│       ├── base.py       # Common LangGraph factory (observe→think→act)
│       ├── miner.py
│       ├── merchant.py
│       ├── hacker.py
│       └── mediator.py
└── tests/
    ├── __init__.py
    ├── conftest.py
    ├── test_a2a_codec.py
    ├── test_llm.py
    ├── test_memory.py
    └── test_graphs.py
```

Plus `apps/backend/internal/transport/http/router.go` got a new `GET /v1/agents/by-string-id/:sid` endpoint.
