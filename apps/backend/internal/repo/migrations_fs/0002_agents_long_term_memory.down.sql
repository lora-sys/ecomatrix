-- 0002_agents_long_term_memory.down.sql
DROP INDEX IF EXISTS agents_ltm_gin_idx;
ALTER TABLE agents DROP COLUMN IF EXISTS long_term_memory;
