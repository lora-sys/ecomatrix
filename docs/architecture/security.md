# Security Architecture

> MVP is a sandbox. Real-world threats are limited but the design accommodates them.

## 1. Threat Model (Phase 1)

| Asset                    | Threat                              | Mitigation                                |
| ------------------------ | ----------------------------------- | ----------------------------------------- |
| Admin token              | Leak via logs                       | Never logged; only stored in env.         |
| Trade ledger             | Double-spend via concurrency        | Row-level locks + idempotency.            |
| A2A endpoint             | Protocol downgrade / fuzz           | Strict envelope validation, 400 on bad.   |
| DB credentials           | Leak via error pages                 | Errors never echo DSN.                    |

## 2. Auth Phases
- Phase 1: shared `ECOMATRIX_ADMIN_TOKEN`; agent calls use a fixed dev token in code (test only).
- Phase 3: per-agent HMAC token with timestamp window.

## 3. Logging Hygiene
- Mask balances > 10 000 GOLD (potential PII leak in prod).
- Never log A2A `payload.reasoning` (free-text; could carry user content).
- `credit_score` for `agent_hacker_*` is a sensitive signal; never log at INFO.

## 4. Supply Chain
- Go: `go mod tidy` + `govulncheck` in CI.
- npm: `pnpm audit --prod` + Renovate.
- Python: `pip-audit` + `uv` lockfile.

## 5. What We Are NOT Doing Yet
- Full OAuth / OIDC.
- Per-IP rate limiting beyond a coarse token bucket.
- KMS-managed secrets.
