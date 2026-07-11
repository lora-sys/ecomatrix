# A2A Protocol v1.1 — Authoritative Spec

> Any change here is an API break and **must** bump `protocol_v` and ship an ADR.

## 1. Envelope

Every A2A message (HTTP request body or WS frame) uses this envelope:

```json
{
  "protocol_v": "1.1",
  "msg_id": "tx_req_9948",
  "sender": "agent_miner_01",
  "action": "EXECUTE_TRADE",
  "payload": { "...": "action-specific, see §2" },
  "timestamp": 1713532588
}
```

| Field        | Type    | Rules                                                          |
| ------------ | ------- | -------------------------------------------------------------- |
| `protocol_v` | string  | MUST be exactly `"1.1"`. Server rejects others with HTTP 400.  |
| `msg_id`     | string  | Unique per sender. `[A-Za-z0-9_]{6,64}`. Idempotency key.     |
| `sender`     | string  | Sender's `string_id` (`agent_*`).                              |
| `action`     | string  | One of the actions in §2.                                      |
| `payload`    | object  | Action-specific. Schemas below.                                |
| `timestamp`  | integer | Unix seconds. Server rejects drift > 300 s unless `X-Allow-Old-Timestamp: 1`. |

## 2. Actions (Phase 1)

### `EXECUTE_TRADE`

**Payload:**
```json
{
  "target_agent": "agent_merchant_03",
  "offer": {
    "currency_type": "GOLD",
    "amount": 150
  },
  "reasoning": "体力濒临枯竭，需紧急购买生存物资"
}
```

**Response (200):**
```json
{
  "protocol_v": "1.1",
  "msg_id": "tx_req_9948",
  "status": "SETTLED",
  "tx_id": 1042,
  "receipt": {
    "from": "agent_miner_01",
    "to": "agent_merchant_03",
    "amount": 150,
    "currency_type": "GOLD",
    "balance_after": { "from": 50, "to": 250 }
  }
}
```

**Error envelope (any non-2xx):**
```json
{
  "protocol_v": "1.1",
  "msg_id": "tx_req_9948",
  "status": "REJECTED",
  "error": {
    "code": "INSUFFICIENT_FUNDS",
    "message": "agent_miner_01 balance 100 < offer.amount 150",
    "retryable": false
  }
}
```

**Error codes:**
- `INVALID_ENVELOPE` (400)
- `UNKNOWN_ACTION` (400)
- `PROTOCOL_MISMATCH` (400)
- `UNKNOWN_AGENT` (404)
- `INSUFFICIENT_FUNDS` (409)
- `SELF_TRADE` (422)
- `RATE_LIMITED` (429)
- `INTERNAL` (500)

### Idempotency

A repeat `POST /v1/trades` with the same `msg_id` for a `SETTLED` trade MUST return the original receipt with `status: SETTLED` and HTTP 200. A repeat for a rejected trade MUST return the same error envelope with the same HTTP code.

## 3. REST Endpoints (Phase 1)

| Method | Path                  | Notes                                            |
| ------ | --------------------- | ------------------------------------------------ |
| GET    | `/healthz`            | Liveness; no DB.                                 |
| GET    | `/readyz`             | Readiness; DB ping.                              |
| GET    | `/v1/agents`          | List agents (paginated).                         |
| GET    | `/v1/agents/{id}`     | Single agent.                                    |
| POST   | `/v1/agents`          | Admin create; requires `X-Admin-Token`.          |
| POST   | `/v1/trades`          | A2A `EXECUTE_TRADE`.                             |
| GET    | `/v1/transactions`    | Recent transactions; `?agent_id=...&limit=...`.  |

## 4. WebSocket (Phase 1)

- `GET /v1/stream` upgrades to WebSocket.
- Server pushes JSON frames:
  ```json
  { "type": "trade.settled",   "tx_id": 1042, "ts": 1713532589 }
  { "type": "trade.rejected",  "msg_id": "tx_req_9949", "code": "INSUFFICIENT_FUNDS" }
  { "type": "agent.heartbeat", "alive": 12 }
  ```
- Server sends heartbeat every 20 s.
- Client should reconnect with exponential backoff (1 s → 30 s, jitter ±20%).

## 5. Versioning

- Backward-compatible additions (new optional field) → minor bump (`1.1` → `1.2`).
- Breaking changes → major bump; old `protocol_v` may be served alongside for ≤ 30 days.
