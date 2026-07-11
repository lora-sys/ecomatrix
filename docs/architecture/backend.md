# Backend Architecture (`apps/backend`)

## Stack
- Go 1.26, Fiber v2, GORM v2, `gofiber/contrib/websocket`, `slog`, `testify`.

## Layout
See `ENGINEERING.md §2`. Key files:
- `cmd/server/main.go` — wires config, db, routes, hub.
- `internal/config/config.go` — env loading + validation (fail-fast).
- `internal/domain/` — entities, errors, no infrastructure imports.
- `internal/service/trade.go` — trade orchestration; the only place that opens a tx.
- `internal/repo/` — GORM repositories; SQL escape hatch via `tx.Raw` is allowed.
- `internal/transport/http/router.go` — route registration.
- `internal/transport/ws/hub.go` — WebSocket fan-out.
- `pkg/a2a/codec.go` — protocol codec, shared with Python via JSON fixtures.

## Required Env Vars
| Var               | Default                                 | Notes                       |
| ----------------- | --------------------------------------- | --------------------------- |
| `ECOMATRIX_HTTP_ADDR` | `:8080`                              | HTTP listen address         |
| `ECOMATRIX_DB_DSN`    | `postgres://repotwin:repotwin@localhost:5432/ecomatrix?sslmode=disable` | Dev only |
| `ECOMATRIX_ADMIN_TOKEN`| `dev-admin-token`                    | Required header for admin   |
| `ECOMATRIX_LOG_LEVEL` | `info`                                | `debug`/`info`/`warn`/`error` |
| `ECOMATRIX_WS_HUB_BUFFER` | `64`                              | Per-conn buffer size        |

## Endpoints
See `docs/architecture/api.md §3`.

## Concurrency Invariants
1. `POST /v1/trades` opens a serializable transaction (Postgres default = `READ COMMITTED`; we explicitly set `SET TRANSACTION ISOLATION LEVEL READ COMMITTED` and take a `SELECT ... FOR UPDATE`).
2. The lock order is always: `MIN(from_id, to_id)` then `MAX(from_id, to_id)` — same in every endpoint that touches two agents — to avoid deadlock.
3. Idempotency: `transactions.msg_id` is `UNIQUE`. Insert conflict returns the existing row.

## Logging
`slog` JSON to stdout. One line per request:
```json
{"ts":"...","level":"INFO","msg":"http","method":"POST","path":"/v1/trades","status":200,"latency_ms":12,"request_id":"...","agent_id":"agent_miner_01"}
```

Never log: full A2A payloads, balances above 10000 GOLD, `credit_score` for `agent_hacker_*`.
