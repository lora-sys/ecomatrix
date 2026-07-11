# Phase 1 — Verification

## Environment
- Go 1.26.4, Node 24, pnpm 11, Docker.
- Postgres 16 (`repotwin-postgres-1` container, db `ecomatrix`).
- Linux 6.x.

## 1. Build

```
$ cd apps/backend && go build ./...
(no output)
```

## 2. Tests (race detector)

```
$ go test -race -count=1 ./...
ok  github.com/ecomatrix/backend/internal/service  1.963s
ok  github.com/ecomatrix/backend/pkg/a2a         1.016s
```

Full per-test output: [`test-results/go-test-race.txt`](./test-results/go-test-race.txt).

The 50-goroutine concurrency test settled 33 / rejected 17 from a sender with 1000 GOLD, each attempting 30 GOLD:

```
trade_test.go:243: settled=33 rejected=17 sender=10 target=990
--- PASS: TestTradeService_Settle_50ConcurrentRacesNoDoubleSpend (0.51s)
```

Invariant verified: balance never went negative, sender ended at exactly `1000 − 33×30 = 10`, target at `33×30 = 990`.

## 3. End-to-end Curl Traces

| Scenario                              | Status | File                                      |
| ------------------------------------- | ------ | ----------------------------------------- |
| `GET /healthz`                        | 200    | implicit (returned `{"status":"ok"}`)     |
| `GET /readyz`                         | 200    | implicit (DB ping succeeded)              |
| `GET /v1/agents?limit=3`              | 200    | [curl/list-agents.txt](./curl/list-agents.txt) |
| `GET /v1/agents/1`                    | 200    | [curl/get-agent.txt](./curl/get-agent.txt) |
| `GET /v1/agents/9999` (not found)     | 404    | [curl/get-agent-404.txt](./curl/get-agent-404.txt) |
| `POST /v1/trades` happy path          | 200    | [curl/trade-happy.txt](./curl/trade-happy.txt) |
| `POST /v1/trades` insufficient funds  | 422    | [curl/trade-insufficient.txt](./curl/trade-insufficient.txt) |
| `POST /v1/trades` protocol 0.9        | 400    | [curl/trade-bad-protocol.txt](./curl/trade-bad-protocol.txt) |
| `POST /v1/trades` self-trade          | 422    | [curl/trade-self.txt](./curl/trade-self.txt) |
| `POST /v1/trades` idempotent replay   | 200    | [curl/trade-replay.txt](./curl/trade-replay.txt) (replay=true, same tx_id) |
| `POST /v1/agents` no admin token      | 401    | [curl/create-agent-noauth.txt](./curl/create-agent-noauth.txt) |
| `POST /v1/agents` with admin token    | 201    | [curl/create-agent-ok.txt](./curl/create-agent-ok.txt) |
| `GET /v1/transactions`                | 200    | [curl/list-transactions.txt](./curl/list-transactions.txt) |
| WebSocket `trade.settled` fan-out     | 1 evt  | [curl/ws-smoke.txt](./curl/ws-smoke.txt)  |

## 4. Result

**PASS** — Phase 1 exit criteria met. Backend is ready for Phase 2 (Python agent) integration.
