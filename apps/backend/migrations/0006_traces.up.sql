-- 0006_traces.up.sql
-- Per-agent observability traces. Every decision, tool call, and error
-- is logged with latency, tokens (when available), and a stable reference
-- to the LLM call that produced it.

CREATE TABLE IF NOT EXISTS traces (
    id          BIGSERIAL PRIMARY KEY,
    agent_id    TEXT NOT NULL REFERENCES agents(string_id) ON DELETE CASCADE,
    kind        TEXT NOT NULL,        -- 'plan' | 'decision' | 'tool_call' | 'tool_result' | 'error' | 'observation' | 'reflection'
    content     TEXT NOT NULL,
    latency_ms  INT,
    tokens_in   INT,
    tokens_out  INT,
    tool_name   TEXT,
    tool_input  JSONB,
    tool_output JSONB,
    cost_usd    NUMERIC(12, 8),
    error_code  TEXT,
    parent_id   BIGINT REFERENCES traces(id) ON DELETE SET NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS traces_agent_created_idx ON traces(agent_id, created_at DESC);
CREATE INDEX IF NOT EXISTS traces_kind_idx ON traces(kind);
