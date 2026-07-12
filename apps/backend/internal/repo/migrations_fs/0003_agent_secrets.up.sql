-- 0003_agent_secrets.up.sql
-- Per-agent HMAC secrets stored in Postgres so they can be rotated without
-- a service restart. env-var secrets (ECOMATRIX_AGENT_SECRETS) still take
-- precedence; the table is the long-term storage.

CREATE TABLE IF NOT EXISTS agent_secrets (
    agent_id     TEXT PRIMARY KEY REFERENCES agents(string_id) ON DELETE CASCADE,
    secret       TEXT NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    rotated_at   TIMESTAMPTZ
);
