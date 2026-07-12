-- 0005_llm_cache.down.sql
DROP INDEX IF EXISTS llm_cache_expires_idx;
DROP TABLE IF EXISTS llm_cache;
