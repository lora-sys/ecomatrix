-- 0004_conversations.up.sql
-- Per-agent LLM conversation log. Every LLM call and tool result is appended
-- so the dashboard's "AI Thought Trace" can show what the agent reasoned
-- about, what tools it called, and what the model returned.

CREATE TABLE IF NOT EXISTS conversations (
    id          BIGSERIAL PRIMARY KEY,
    agent_id    TEXT NOT NULL REFERENCES agents(string_id) ON DELETE CASCADE,
    role        TEXT NOT NULL,        -- 'user' | 'assistant' | 'tool' | 'system' | 'error'
    content     TEXT NOT NULL,
    tool_name   TEXT,                  -- set when role = 'tool'
    tool_input  JSONB,
    tool_output JSONB,
    error_code  TEXT,                  -- set when role = 'error'
    latency_ms  INT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS conversations_agent_created_idx ON conversations(agent_id, created_at DESC);
