-- 0005_llm_cache.up.sql
-- LLM response cache. Keyed by sha256 of (model, system, user messages, temperature).
-- Allows the dashboard to replay prior decisions without re-calling the LLM,
-- and lets a fleet of agent processes share cached responses.

CREATE TABLE IF NOT EXISTS llm_cache (
    key          TEXT PRIMARY KEY,         -- sha256 hex
    model        TEXT NOT NULL,
    response     TEXT NOT NULL,
    prompt_hash  TEXT NOT NULL,             -- for debugging
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at   TIMESTAMPTZ NOT NULL,
    hit_count    INT NOT NULL DEFAULT 0
);
CREATE INDEX IF NOT EXISTS llm_cache_expires_idx ON llm_cache(expires_at);
