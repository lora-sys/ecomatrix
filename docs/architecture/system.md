# System Architecture

## C4-Level Overview

```
┌────────────────────────────────────────────────────────────────────────┐
│                        Next.js (God's-Eye Dashboard)                  │
│  App Router · RSC + Client islands · Aceternity · Framer Motion       │
│  ──── HTTPS / WSS ────                                               │
└────────────────────────────────────────────────────────────────────────┘
                                  ▲
                                  │ WebSocket (state deltas)
                                  │ HTTP   (queries, commands)
                                  │
┌────────────────────────────────────────────────────────────────────────┐
│                  Go Backend (Fiber + Gorilla-WS)                      │
│                                                                        │
│  HTTP layer ──► Service ──► Repository ──► Postgres (GORM)            │
│       ▲             ▲            ▲                                    │
│       │             │            │                                    │
│  Middleware     A2A codec   SELECT FOR UPDATE                         │
│  (auth,         (pkg/a2a)   (row-level locks)                         │
│   req id)                                                           │
│       ▲                                                             │
│       │ HTTP A2A                                                    │
└────────────────────────────────────────────────────────────────────────┘
                                  ▲
                                  │
┌────────────────────────────────────────────────────────────────────────┐
│              Python Agent Runner (LangGraph)                          │
│  Per-agent graph · LLM provider · Memory (short + long-term)         │
└────────────────────────────────────────────────────────────────────────┘
```

## Layers

1. **Presentation (Next.js)** — read-mostly dashboard; writes are limited to admin actions in MVP.
2. **Gateway (Go)** — owns the A2A protocol, request validation, concurrency control.
3. **Ledger (Postgres)** — single source of truth; ACID + row locks.
4. **Mind (Python)** — generates intents; never holds the ledger.

## Cross-Cutting Concerns

- **Auth:** Phase 1 has a shared admin token (env). Phase 3 adds per-agent signed tokens.
- **Observability:** structured logs everywhere; `/healthz` + `/readyz`; metrics in Phase 3.
- **Idempotency:** `msg_id` is the universal idempotency key for any state-changing A2A action.
- **Backpressure:** WS hub drops slow consumers; HTTP returns 429 with retry hint when saturated.

## Sequence — A Trade

```
Python Agent ──POST /v1/trades──► Go Service
                                   │
                                   ├─ Validate A2A envelope (protocol_v=1.1)
                                   ├─ BEGIN; SELECT ... FOR UPDATE agents WHERE id IN (sender, target)
                                   ├─ Check sender.balance ≥ amount
                                   ├─   ├─ yes: UPDATE balances, INSERT transactions
                                   ├─   └─ no:  ROLLBACK, return 409
                                   ├─ COMMIT
                                   ├─ Emit WS event: trade.settled | trade.rejected
                                   └─ Return JSON receipt
                  ◄── receipt ────
Frontend WS hub  ──broadcast──►  All connected dashboards
```

## Failure Modes

| Failure                          | Detection              | Response                                      |
| -------------------------------- | ---------------------- | --------------------------------------------- |
| DB connection lost               | `/readyz` 503          | Service rejects writes; reads fall back to cache. |
| LLM provider timeout (Phase 2)   | Per-call timeout       | Agent retries with exponential backoff.       |
| WS client behind flaky proxy     | Heartbeat every 20 s   | Drop + reconnect with backoff.                |
| Concurrent double-spend attempt  | Row lock               | Second tx blocks; first wins; second sees 409. |
