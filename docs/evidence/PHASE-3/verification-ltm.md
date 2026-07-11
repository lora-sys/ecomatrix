# Phase 3.2 — Long-Term Memory Verification

## 1. Migration 0002

```
$ docker exec repotwin-postgres-1 psql -U repotwin -d ecomatrix -c "\d agents"
 long_term_memory | jsonb | not null | '{}'::jsonb
 "agents_ltm_gin_idx" gin (long_term_memory)
```

The migration runner auto-applied 0002 on the next boot.

## 2. Go Backend Suite

```
$ cd apps/backend && go test -race -count=1 ./...
ok  github.com/ecomatrix/backend/internal/repo     1.106s   (new: 1 test)
ok  github.com/ecomatrix/backend/internal/service  2.344s
ok  github.com/ecomatrix/backend/pkg/a2a         1.028s
```

All previously passing tests still pass; the 50-goroutine concurrency proof is unaffected because the new column defaults to `'{}'` on INSERT.

## 3. Python Agent Suite

```
$ cd apps/agent && pytest
============================== 19 passed in 0.36s ==============================
```

Was 18; +1 PostgresLTM roundtrip test using `httpx.MockTransport` (no real backend required).

## 4. LTM API End-to-End

PUT seeds LTM:

```
PUT /v1/agents/by-string-id/agent_miner_01/long-term-memory
body: {"summary":"低体力，已与 merchant_03 多次交易","facts":[...]}
→ 200 {"long_term_memory":{"summary":"…","facts":[4 entries]}}
```

GET reads back the same payload:

```
GET /v1/agents/by-string-id/agent_miner_01/long-term-memory
→ 200 {"long_term_memory":{"summary":"…","facts":[4 entries]}}
```

Unknown agent:

```
GET /v1/agents/by-string-id/agent_nope/long-term-memory
→ 404 {"error":"agent not found"}
```

Files: [`curl/ltm-put.json`](./curl/ltm-put.json), [`curl/ltm-get.json`](./curl/ltm-get.json), [`curl/multi-with-ltm.txt`](./curl/multi-with-ltm.txt), [`curl/ltm-after-multi.json`](./curl/ltm-after-multi.json).

## 5. Multi-Scenario with Postgres LTM

```
$ python -m ecomatrix.runner --scenario multi --ticks 3 --tick-seconds 0.3
{
  "agents": 13,
  "ticks": 3,
  "settled": 39,
  "rejected": 3,
  "feeds_posted": 50,
  "errors": [],
  "world_initial": 2540,
  "world_final": 2540,
  "conservation": true
}
```

LTM after the run grew by 3 new `settled tx_…` facts for `agent_miner_01` (proving the Python agent's `ltm.update(append_fact=…)` round-trips through the Go PUT endpoint).

## 6. Dashboard Screenshot

The agent detail page now has a third panel — `长期记忆 · LTM` — alongside vitals and recent trades:

[`screenshots/agent-desktop.png`](./screenshots/agent-desktop.png) shows:
- Vitals: BAL 70, VIT 80, CR 60 (after the multi-scenario reduced balance from 100).
- LTM: summary "低体力，已与 merchant_03 多次交易" + 7 facts (4 seed + 3 fresh from the run).
- Recent trades: 4 SETTLED `agent_miner_01 → agent_merchant_01` entries via the tracing beam.

## Result

**PASS.** Migration 0002 ships, the LTM endpoints are protected with a 500/50 cap on summary/facts, the Python agent writes through them on every settle, and the dashboard renders them. ISS-015 is closed.
