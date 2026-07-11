# Security Audit (live + cold-start)

Date: 2026-07-11
Method: cross-language HMAC parity check + 7 live adversarial tests against the running backend.

## 1. HMAC canonical form: Go ↔ Python byte-for-byte

Both languages produce the same signature for the same inputs:

```
secret  = miner-secret-a
method  = POST
path    = /v1/trades
ts      = 1713532588
body    = {"protocol_v":"1.1"}

canonical = POST\n/v1/trades\n1713532588\n219940e4ac84c481c91a08d2d3d765398788e95241a12e02fd5a1a5e77d21926
sig      = 57c4bbb688c947d40ce707c965ff66572a6ab2f9ebe9319efdf7fc65856be921
```

- Go: `apps/backend/internal/auth/hmac.go::ComputeSignature` (raw `hmac.New(sha256.New, …)` over `METHOD\nPATH\nTS\nsha256_hex(BODY)`).
- Python: `apps/agent/ecomatrix/a2a.py::_sign` (same canonical form).
- Constant-time compare via `hmac.Equal` on the Go side.

✓ Parity confirmed. A replay-defended HMAC scheme is in place; tampering with the body changes the digest, expired timestamps are rejected within the 5-min window, and unknown agents get 401 before the request reaches the service.

## 2. Live adversarial tests

All against the running backend (`/tmp/ecomatrix-server` with `ECOMATRIX_AGENT_SECRETS=agent_miner_01=miner-secret-a`).

| # | Attack                                            | Expected | Got      | Verdict |
| - | ------------------------------------------------- | -------- | -------- | ------- |
| 1 | Unsigned trade (configured agent)                 | 401      | 401 "agent not configured for HMAC signing" | ✓ |
| 2 | Properly signed trade                              | 200 + SETTLED | 200 SETTLED | ✓ |
| 3 | Stale timestamp (5 min outside window)             | 401      | 401       | ✓ |
| 4 | SQL injection in body (`'; DROP TABLE…`)            | 4xx      | 400 (rejected by codec before SQL) | ✓ |
| 5 | SQL injection in URL path (`by-string-id/<inj>`)    | 401      | 401 (rejected by HMAC; never reaches repo) | ✓ |
| 6 | 1 MB body                                          | 4xx      | 400 (Fiber body limit) | ✓ |
| 7 | Content-Length header spoofing                     | 401 or signature mismatch | 401 (HMAC on actual body, not header) | ✓ |
| 8 | Unauthenticated GET /v1/agents (read)              | 200      | 200 (intentional — read endpoints are open) | ✓ (by design) |
| 9 | Unknown agent via by-string-id                     | 404      | 404       | ✓ |
| 10 | DB schema intact after all attacks                 | 11 rows  | 11 rows   | ✓ |

## 3. Findings (by severity)

### Critical
- None.

### High
- **BFF proxy in dev is broad** (`/api/proxy/{feeds,metrics}` passthroughs). Acceptable for the dev loop but in prod the dashboard should call the backend directly with a tightened CORS allowlist (see Architecture Review #3).

### Medium
- **No rate limiting** on `/v1/trades` or `/v1/feeds`. An attacker with a valid HMAC could DOS the ledger or pollute the social feed. Recommended: token-bucket per agent_id (Phase 4).
- **WS endpoint is open to all origins within the CORS allowlist** (anyone can subscribe to `/v1/stream`). Used by the dashboard, but a malicious in-allowlist client could scrape live state. Recommended: HMAC for the upgrade handshake too, or at least `Sec-WebSocket-Protocol` pinning (Phase 4).
- **Replay window is the 5-min skew**, not a nonce. An attacker who captures a valid request can replay it within 5 minutes. The DB-level idempotency by `msg_id` partially mitigates this (the same trade won't double-execute), but a different `msg_id` per replay still works. Recommended: nonce + seen-set (Phase 4).

### Low
- **`ECOMATRIX_DEV` defaults to `true`**, which makes CORS permissive by default. A misconfigured prod deploy gets `Access-Control-Allow-Origin: *`. Recommended: flip the default to `false`, or remove the env var entirely and require explicit opt-in.
- **Agent secrets in env vars** are visible in `/proc/<pid>/environ` if the process is introspectable. Not a unique issue but a general one for any env-based secret store. Recommended: move to Postgres-backed store (Phase 4).
- **No font files** loaded in the frontend despite DESIGN.md §3 specifying Space Grotesk / Inter / JetBrains Mono. Demo uses system-ui. Low priority — purely cosmetic.

## 4. Verdict

The security baseline is real. The two big holes — shared admin token for everything, and permissive CORS — are closed by Phase 3.7 (per-agent HMAC) and Phase 3.6 (CORS allowlist). The remaining items are forward-looking: rate limiting, replay nonces, secret rotation. None are blocking the demo or the harness operating system.

The 7 live adversarial tests all behaved as expected. SQL injection attempts are rejected at the A2A codec or the HMAC middleware, never reaching SQL. Body limits are enforced by Fiber. HMAC parity between Go and Python is byte-for-byte.

## 5. Reference

- [security-review.md](./security-review.md) — full grep-based audit.
- [architecture-review.md](./architecture-review.md) — coupling and tech-debt notes.
- [ux-review.md](./ux-review.md) — a11y and motion-preference notes.
- [security-review.md HMAC trace](../../PHASE-3/curl/hmac-trace.txt) — original 4-scenario HMAC test.
- [security-review.md live trace](../../PHASE-3/curl/cors-allowlist-trace.txt) — CORS allowlist trace.
