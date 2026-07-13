CREATE TABLE IF NOT EXISTS supervisor_runs (
    id            BIGSERIAL PRIMARY KEY,
    goal          TEXT NOT NULL,
    status        TEXT NOT NULL,        -- 'started' | 'finished'
    error         TEXT NOT NULL DEFAULT '',
    warnings      JSONB NOT NULL DEFAULT '[]'::jsonb,
    subtasks      JSONB NOT NULL DEFAULT '[]'::jsonb,
    worker_results JSONB NOT NULL DEFAULT '[]'::jsonb,
    final_summary TEXT NOT NULL DEFAULT '',
    tokens_used   INT NOT NULL DEFAULT 0,
    tokens_budget INT NOT NULL DEFAULT 0,
    started_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    finished_at   TIMESTAMPTZ,
    duration_ms   INT NOT NULL DEFAULT 0
);
CREATE INDEX IF NOT EXISTS supervisor_runs_started_at_idx
    ON supervisor_runs (started_at DESC);
