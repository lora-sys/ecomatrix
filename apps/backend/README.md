# EcoMatrix Backend

Go service implementing the A2A protocol gateway and the WebSocket fan-out for the dashboard.

## Quick start (dev)

```bash
# 1. Bring up Postgres
docker compose up -d db

# 2. Build + run
make build
ECOMATRIX_DB_DSN=postgres://repotwin:repotwin@localhost:5432/ecomatrix?sslmode=disable ./bin/server

# 3. Seed (in a second terminal)
./bin/seed

# 4. Smoke test
curl http://localhost:8080/healthz
curl http://localhost:8080/v1/agents
```

## Tests

```bash
make test-race
```

## A2A example

```bash
curl -X POST http://localhost:8080/v1/trades \
  -H 'Content-Type: application/json' \
  -d '{
    "protocol_v":"1.1",
    "msg_id":"tx_req_demo01",
    "sender":"agent_miner_01",
    "action":"EXECUTE_TRADE",
    "payload":{
      "target_agent":"agent_merchant_03",
      "offer":{"currency_type":"GOLD","amount":40},
      "reasoning":"buying supplies"
    },
    "timestamp": '"$(date +%s)"'
  }'
```

## Layout

See `../../ENGINEERING.md §2` and `../../docs/architecture/backend.md`.
