# Phase 2 — Verification

## Environment
- Python 3.12 (uv-managed venv at `apps/agent/.venv`).
- LangGraph, langchain-core, httpx, pydantic.
- Go backend on `:8080` against Postgres `ecomatrix`.

## 1. Unit + Integration Tests

```
$ cd apps/agent && . .venv/bin/activate && pytest -v
tests/test_a2a_codec.py ........ (8 tests, mirror of pkg/a2a/codec_test.go)
tests/test_graphs.py    ...     (3 tests; miner, merchant, all four jobs)
tests/test_llm.py       .       (1 test; stub provider JSON contract)
tests/test_memory.py    ..      (2 tests; LTM roundtrip + STM cap)
============================== 14 passed in 0.28s ==============================
```

Full output: [`test-results/pytest.txt`](./test-results/pytest.txt).

## 2. Go Backend Regression

Re-ran the Phase 1 suite after adding `GET /v1/agents/by-string-id/:sid`:

```
$ cd apps/backend && go test -race -count=1 ./...
ok  github.com/ecomatrix/backend/internal/service  1.971s
ok  github.com/ecomatrix/backend/pkg/a2a         1.014s
```

Full output: [`../PHASE-1/test-results/go-test-race-with-by-string-id.txt`](../PHASE-1/test-results/go-test-race-with-by-string-id.txt).

## 3. Two-Agent End-to-End Scenario

Two scenarios captured: 5 ticks (10 trades) and 10 ticks (20 trades). Both runs:

- Started from a freshly seeded DB (`agent_miner_01` = 100 GOLD, `agent_merchant_01` = 200 GOLD).
- Ran each tick: `observe → think → act` for both agents.
- All trades settled; zero rejected; zero unexpected errors.
- **World-total GOLD conservation**: 2560 → 2560 after every run.

Run 1 summary:

```json
{
  "settled": 10,
  "rejected": 0,
  "errors": [],
  "initial":  {"miner": 100, "merchant": 200, "world": 2560},
  "final":    {"miner":  50, "merchant": 200, "world": 2560},
  "ledger_size": 64,
  "conservation": true
}
```

Run 2 summary (10 ticks, miner drains to 0):

```json
{
  "settled": 20,
  "rejected": 0,
  "errors": [],
  "initial":  {"miner": 100, "merchant": 200, "world": 2560},
  "final":    {"miner":   0, "merchant": 200, "world": 2560},
  "conservation": true
}
```

Full log: [`scenario/two-agent-run.txt`](./scenario/two-agent-run.txt).

## 4. New Endpoint

`GET /v1/agents/by-string-id/agent_miner_01` — used by the Python client. Trace: [`curl/by-string-id.txt`](./curl/by-string-id.txt).

404 case (`agent_nope`): [`curl/by-string-id-404.txt`](./curl/by-string-id-404.txt).

## 5. Ledger Verification

After the scenario, the `transactions` table contains 64+ rows (Phase 1 evidence + new scenario trades). Cross-section in [`curl/transactions-after-scenario.txt`](./curl/transactions-after-scenario.txt).

## Result

**PASS** — Phase 2 exit criteria met. The Python agent runtime drives the Go backend end-to-end with world-level conservation.
