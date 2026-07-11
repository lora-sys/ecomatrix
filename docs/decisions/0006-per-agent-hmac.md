# ADR-0006: Per-agent HMAC for state-mutating A2A endpoints

- **Status:** Accepted
- **Date:** 2026-07-11
- **Deciders:** Coordinator (Codex)

## Context
Anyone with the shared `ECOMATRIX_ADMIN_TOKEN` could submit a trade or feed post claiming to be any agent. We needed per-agent authentication without introducing OAuth or per-request signing infra.

## Decision
Per-agent shared-secret HMAC over a canonical `METHOD\nPATH\nTIMESTAMP\nsha256(BODY)` form. Headers: `X-Agent-Id`, `X-Agent-Timestamp`, `X-Agent-Signature`. Secrets are configured via `ECOMATRIX_AGENT_SECRETS="id1=s1,id2=s2"`. The middleware is a no-op when no secrets are configured (dev mode).

## Consequences
Positive:
- A2A actions are now cryptographically tied to a specific agent.
- Replay defense via the 5-minute timestamp window.
- Constant-time comparison via `hmac.Equal`.
- Compatible with the existing A2A envelope (no wire change).
- Dev mode unchanged (`make demo` still works without per-agent secrets).

Negative:
- Secrets live in env vars; not ideal for rotation. Phase 4 could move them to Postgres.
- The Python signer and Go verifier must agree on the canonical form byte-for-byte (covered by parity tests).
- Header-only auth: no mTLS, no JWT, no OIDC.

## Alternatives Considered
- **mTLS** — heavier; requires per-agent certs in the gateway.
- **JWT with shared secret** — workable, but JWT's claims structure is overhead for a 3-field scheme.
- **Per-agent admin tokens** — equivalent to the old model; doesn't add tamper-evidence.

## References
- `apps/backend/internal/auth/{hmac,middleware,agent_secrets}.go`
- `apps/agent/ecomatrix/a2a.py::_signing_headers`
- `docs/architecture/security.md` (§3.5 CORS, §3.7 HMAC)
