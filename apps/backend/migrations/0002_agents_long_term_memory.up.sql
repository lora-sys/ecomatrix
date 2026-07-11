-- 0002_agents_long_term_memory.up.sql
-- Adds the long-term-memory JSONB column to agents.

ALTER TABLE agents
    ADD COLUMN IF NOT EXISTS long_term_memory JSONB NOT NULL DEFAULT '{}'::jsonb;

-- GIN index so we can search inside the JSONB without full table scans later.
CREATE INDEX IF NOT EXISTS agents_ltm_gin_idx ON agents USING GIN (long_term_memory);
