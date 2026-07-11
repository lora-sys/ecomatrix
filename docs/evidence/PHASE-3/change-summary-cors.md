# Phase 3.6 — CORS Allowlist

## What
Replaced the permissive CORS middleware (reflected any `Origin`) with an explicit allowlist driven by `ECOMATRIX_CORS_ALLOWED_ORIGINS` (comma-separated).

## Why
The previous middleware reflected the request's `Origin` unconditionally, which is the textbook insecure CORS configuration — any origin could call `/v1/*` with the browser's credentials. The new allowlist:
- `"*"` → wildcard (intentional, dev only).
- `"http://localhost:3100,https://dashboard.example.com"` → exact match.
- Unset + `ECOMATRIX_DEV=true` (the `make demo` default) → defaults to `"*"`.
- Unset + `ECOMATRIX_DEV=false` → **no CORS headers** (the browser blocks the response client-side). Production is locked by default; the operator must explicitly set the allowlist.

Preflight requests (`OPTIONS`) from disallowed origins return **403**.

## Files

```
apps/backend/
├── internal/config/config.go                 # +CORSAllowedOrigins +DevMode +mustBools helper
├── internal/transport/http/router.go         # rewrite corsMiddleware; +corsConfig; loadCORSOrigins()
└── internal/transport/http/cors_test.go      # NEW: 7 tests covering wildcard/exact/empty/preflight

docs/architecture/security.md                 # +§3.5 CORS Allowlist
apps/backend/.env.example                     # +ECOMATRIX_CORS_ALLOWED_ORIGINS +ECOMATRIX_DEV
```

## Verified
7 new tests pass; existing suite stays green (28 Go tests total under `-race`):

```
TestCORS_Wildcard_AllowsAnyOrigin             PASS
TestCORS_ExactMatch_AllowsListedOrigin        PASS
TestCORS_ExactMatch_BlocksUnknownOrigin       PASS  (no CORS headers; browser blocks)
TestCORS_EmptyAllowlist_NoHeaders             PASS  (prod default: locked down)
TestCORS_Preflight_AllowedOrigin_Returns204   PASS
TestCORS_Preflight_DisallowedOrigin_Returns403 PASS
TestCORS_NoOrigin_NoHeaders                   PASS  (server-to-server unaffected)
```

`make demo` continues to work because `ECOMATRIX_DEV=true` defaults the allowlist to `"*"` (matching the prior behavior).
