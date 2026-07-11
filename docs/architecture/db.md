# Database Architecture

## Engine
PostgreSQL 16. One DB (`ecomatrix`) in dev; production DB name is an env var.

## Schema (Phase 1)

```sql
-- agents: the world's citizens
CREATE TABLE agents (
    id              BIGSERIAL PRIMARY KEY,
    string_id       TEXT NOT NULL UNIQUE,           -- e.g. agent_miner_01
    job_type        TEXT NOT NULL,                  -- miner | merchant | hacker | mediator
    balance         BIGINT NOT NULL DEFAULT 0 CHECK (balance >= 0),
    vitality        INT    NOT NULL DEFAULT 100 CHECK (vitality BETWEEN 0 AND 100),
    credit_score    INT    NOT NULL DEFAULT 50  CHECK (credit_score BETWEEN 0 AND 100),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX agents_job_type_idx ON agents(job_type);

-- transactions: the immutable ledger
CREATE TABLE transactions (
    id              BIGSERIAL PRIMARY KEY,
    tx_id           TEXT NOT NULL UNIQUE,           -- public id; mirror of internal id as text
    msg_id          TEXT NOT NULL UNIQUE,           -- A2A idempotency key
    from_id         BIGINT NOT NULL REFERENCES agents(id),
    to_id           BIGINT NOT NULL REFERENCES agents(id),
    amount          BIGINT NOT NULL CHECK (amount > 0),
    currency_type   TEXT  NOT NULL DEFAULT 'GOLD',
    status          TEXT  NOT NULL,                 -- SETTLED | REJECTED
    error_code      TEXT,                            -- NULL when SETTLED
    reasoning       TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (from_id <> to_id)
);
CREATE INDEX transactions_from_created_idx ON transactions(from_id, created_at DESC);
CREATE INDEX transactions_to_created_idx   ON transactions(to_id,   created_at DESC);

-- social_feeds: agent posts (Phase 2 onwards, table present in Phase 1 for forward compat)
CREATE TABLE social_feeds (
    id              BIGSERIAL PRIMARY KEY,
    agent_id        BIGINT NOT NULL REFERENCES agents(id),
    content         TEXT NOT NULL,
    intent_type     TEXT NOT NULL,                  -- OFFER | REQUEST | SOCIAL | META
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX social_feeds_created_idx ON social_feeds(created_at DESC);
```

## Invariants
- `agents.balance >= 0` (enforced by CHECK + service-layer guard).
- `transactions.amount > 0` and `from_id <> to_id`.
- Every settled trade produces exactly one `transactions` row.

## Locking
Trade service:
```
BEGIN;
SET LOCAL lock_timeout = '2s';
SELECT balance FROM agents WHERE id = :min_id FOR UPDATE;
SELECT balance FROM agents WHERE id = :max_id FOR UPDATE;
-- validate, update balances, INSERT transaction
COMMIT;
```
Lock order: ascending by `id`. Same order everywhere.

## Migrations
- Forward-only. Each migration is a numbered SQL file in `apps/backend/migrations/`.
- `0001_init.up.sql` / `0001_init.down.sql` for the initial schema.
- CI runs `up` against an ephemeral DB and `down` to verify rollback.

## Seed
`apps/backend/cmd/seed/main.go`:
- 5 miners, 3 merchants, 2 hackers, 1 mediator.
- Starting balance: miner=100, merchant=200, hacker=80, mediator=300 GOLD.
- Re-running is idempotent (uses `ON CONFLICT (string_id) DO NOTHING`).
