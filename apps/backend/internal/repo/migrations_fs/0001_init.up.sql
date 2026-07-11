-- 0001_init.up.sql
-- Schema per docs/architecture/db.md.

CREATE TABLE IF NOT EXISTS agents (
    id              BIGSERIAL PRIMARY KEY,
    string_id       TEXT NOT NULL UNIQUE,
    job_type        TEXT NOT NULL,
    balance         BIGINT NOT NULL DEFAULT 0 CHECK (balance >= 0),
    vitality        INT    NOT NULL DEFAULT 100 CHECK (vitality BETWEEN 0 AND 100),
    credit_score    INT    NOT NULL DEFAULT 50  CHECK (credit_score BETWEEN 0 AND 100),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS agents_job_type_idx ON agents(job_type);

CREATE TABLE IF NOT EXISTS transactions (
    id              BIGSERIAL PRIMARY KEY,
    tx_id           TEXT NOT NULL UNIQUE,
    msg_id          TEXT NOT NULL UNIQUE,
    from_id         BIGINT NOT NULL REFERENCES agents(id),
    to_id           BIGINT NOT NULL REFERENCES agents(id),
    amount          BIGINT NOT NULL CHECK (amount > 0),
    currency_type   TEXT  NOT NULL DEFAULT 'GOLD',
    status          TEXT  NOT NULL,
    error_code      TEXT,
    reasoning       TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (from_id <> to_id)
);
CREATE INDEX IF NOT EXISTS transactions_from_created_idx ON transactions(from_id, created_at DESC);
CREATE INDEX IF NOT EXISTS transactions_to_created_idx   ON transactions(to_id,   created_at DESC);

CREATE TABLE IF NOT EXISTS social_feeds (
    id              BIGSERIAL PRIMARY KEY,
    agent_id        BIGINT NOT NULL REFERENCES agents(id),
    content         TEXT NOT NULL,
    intent_type     TEXT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS social_feeds_created_idx ON social_feeds(created_at DESC);
