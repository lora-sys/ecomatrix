# Phase 3.7 — Per-Agent HMAC Authentication

## What
State-mutating A2A endpoints (`POST /v1/trades`, `POST /v1/feeds`) now require a per-agent HMAC signature on every request. The shared `X-Admin-Token` is unchanged and still gates the `POST /v1/agents` admin route.

## Why
The previous auth model had a single shared admin token; any operator with that token could submit a trade or feed post claiming to be any agent (`agent_miner_01`, etc.). HMAC ties each A2A action to a per-agent shared secret that only the legitimate agent (or its operator) holds.

## Scheme

```
string_to_sign = METHOD + "\n" + PATH + "\n" + TIMESTAMP + "\n" + sha256_hex(BODY)
signature     = hex_hmac_sha256(secret, string_to_sign)
```

Headers required on signed requests:

| Header              | Value                                       |
| ------------------- | ------------------------------------------- |
| `X-Agent-Id`        | the agent's `string_id` (`agent_miner_01`)  |
| `X-Agent-Timestamp` | unix seconds; window ±5 min (replay defense) |
| `X-Agent-Signature` | hex digest                                  |

The middleware caches the body so downstream handlers can still read it after the middleware consumes it. Constant-time comparison via `hmac.Equal`.

## Configuration

```
ECOMATRIX_AGENT_SECRETS="agent_miner_01=miner-secret-a,agent_merchant_01=merchant-secret-a"
```

When unset, the middleware is a no-op — `make demo` continues to work without per-agent secrets.

## Files

```
apps/backend/
├── internal/auth/hmac.go                     # NEW: canonical form + Verify + sentinel errors
├── internal/auth/hmac_test.go                # NEW: 8 unit tests
├── internal/auth/agent_secrets.go            # NEW: env-backed AgentSecretStore
├── internal/auth/middleware.go               # NEW: RequireAgentSignature Fiber middleware
├── internal/transport/http/router.go         # wire middleware + CORSConfigFromConfig
└── cmd/server/main.go                        # construct secret store; pass into Server

apps/agent/
├── ecomatrix/a2a.py                          # _sign / _signing_headers; sign every POST
└── tests/test_a2a_codec.py                   # +4 parity tests (canonical form + env config)
```

## Verified

**Unit tests (Go):** 8 new auth tests; 36 Go tests total under `-race` (the 50-goroutine concurrency proof unchanged at 33/17).

**Unit tests (Python):** 4 new parity tests; 23 pytest tests total.

**Live end-to-end** (see [`curl/hmac-trace.txt`](./curl/hmac-trace.txt)):

| # | Scenario                                 | Expected | Got      |
| - | ----------------------------------------- | -------- | -------- |
| 1 | Properly signed trade                     | 200 + SETTLED | ✅ HTTP 200, tx_f3d99b... settled |
| 2 | Unsigned from configured agent            | 401      | ✅ HTTP 401 "agent not configured for HMAC signing" |
| 3 | Trade from unconfigured agent             | 401      | ✅ HTTP 401 "agent not configured for HMAC signing" |
| 4 | Tampered body, original signature         | 401 + HMAC_INVALID | ✅ HTTP 401 "auth: signature mismatch" |
