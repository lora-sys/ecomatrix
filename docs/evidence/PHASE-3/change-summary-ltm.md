# Phase 3.2 — Long-Term Memory

## What
Closed the schema gap on `agents.long_term_memory`. The column has been in the project-memory doc since Phase 1; this turns it into a real, writeable, dashboard-visible surface.

## Why
ISS-015 was the only remaining Todo in `PROJECT_STATUS.md`. The agent runtime needs memory that survives process restarts and is observable from the dashboard.

## Files Changed

```
apps/backend/
├── migrations/0002_agents_long_term_memory.{up,down}.sql       # NEW
├── internal/repo/migrations_fs/0002_….{up,down}.sql            # NEW (embedded copy)
├── internal/domain/agent.go                                   # +LongTermMemory struct
├── internal/repo/db.go                                         # +LongTermMemory column on AgentModel
├── internal/repo/agent_repo.go                                 # +GetLongTermMemory +SetLongTermMemory
├── internal/repo/ltm_test.go                                   # NEW: roundtrip test
└── internal/transport/http/router.go                           # +GET /v1/agents/by-string-id/:sid/long-term-memory
                                                                   +PUT same

apps/agent/
├── ecomatrix/memory.py                                         # +PostgresLongTermMemory
│                                                                  (LongTermMemory kept as alias for FileLongTermMemory)
├── ecomatrix/runner.py                                         # use Postgres LTM when ECOMATRIX_AGENT_LTM=postgres
└── tests/test_memory.py                                        # +PostgresLTM roundtrip via httpx MockTransport

apps/frontend/
├── lib/api.ts                                                  # +fetchLongTermMemory
└── app/agents/[id]/page.tsx                                    # +长期记忆 · LTM card on detail page
```
