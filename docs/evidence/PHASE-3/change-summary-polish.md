# Phase 3.9 — Polish pass

## What

Closed the issues the reviews and the browser E2E surfaced. 8 polish items shipped in one commit.

## Files changed

```
apps/backend/
├── internal/auth/ratelimit.go              # NEW: token-bucket per (agent, action)
├── internal/auth/ratelimit_test.go         # NEW: 3 tests
├── internal/repo/tx_repo.go                # + Recent() (moved from tx_repo_list.go)
├── internal/repo/tx_repo_list.go           # REMOVED
├── internal/config/config.go                # ECOMATRIX_DEV defaults to false (was true)
├── internal/transport/http/router.go        # +rateLimit middleware on /v1/trades + /v1/feeds
├── internal/transport/http/cors_test.go     # still 7 CORS tests
└── cmd/server/main.go                      # wire RateLimit + tighter dev defaults

apps/frontend/
├── styles/globals.css                       # +prefers-reduced-motion (DESIGN.md §6)
├── app/fonts.ts                             # NEW: next/font loaders
├── app/layout.tsx                           # +font CSS variables
├── tailwind.config.ts                       # +font family references
├── app/page.tsx                             # +SSR fetch of initial metrics
├── components/live-provider.tsx             # +seed initial metrics + error reporting
├── components/error-banner.tsx              # NEW: top-level error banner
├── components/trade-feed.tsx                # +role="log" aria-live="polite"
├── components/social-feed.tsx               # +role="log" aria-live="polite"
├── app/dashboard-client.tsx                 # +<ErrorBanner /> first child
└── hooks/store.ts                           # +fetchError slice
```

## Verified

- **All Go tests pass under -race** (39 total: 36 prior + 3 new rate-limit).
- **TypeScript strict + build clean.**
- **Browser E2E (agent-browser, real Chromium):** 5/5 features, rate limit 30 burst + 5× 429.
- **Initial KPIs:** agents=13, gold=2,560 from first paint (no more zeros).
- **WS events arrive:** trade and feed posts visible within 3s of firing.

## What each polish item fixed

| Item | Source | Fix |
| ---- | ------ | --- |
| Initial KPIs show 0/0 | browser E2E | SSR-fetch `/v1/metrics` in `app/page.tsx`; seed store on mount in `LiveProvider` |
| `prefers-reduced-motion` not respected | DESIGN.md §6 / UX review | `globals.css` media query disables animations + Framer Motion's own |
| No ARIA on live feeds | UX review | `role="log" aria-live="polite" aria-relevant="additions"` on TradeFeed + SocialFeed lists; `role="status"` on empty states |
| Silent fetch failure | UX review | `ErrorBanner` mounted in dashboard; `fetchError` slice in store; `LiveProvider` reports failures |
| No rate limit on /v1/trades /v1/feeds | security review (Medium) | `auth.RateLimiter` token-bucket per (sender, action); burst 30, refill 5/s |
| ECOMATRIX_DEV defaults to permissive | security review (Low) | default flipped to false; CORS only permissive when explicitly opted in |
| `tx_repo_list.go` split | architecture review | merged into `tx_repo.go` |
| Fonts declared but not loaded | UX review | `next/font/google` for Space Grotesk, Inter, JetBrains Mono |

## Screenshots

- [`docs/evidence/PHASE-3/polish/snapshots/01-initial.png`](../polish/snapshots/01-initial.png) — dashboard with real initial KPIs
- [`docs/evidence/PHASE-3/polish/snapshots/03-after-feed.png`](../polish/snapshots/03-after-feed.png) — trade + feed post live
- [`docs/evidence/PHASE-3/polish/snapshots/04-agent-detail.png`](../polish/snapshots/04-agent-detail.png) — agent detail with vitals + LTM

## Polish E2E results

```
initial KPIs: agents=13 gold=2,560 qps=0.00 ws=0     (was: 0/0/0/0)
trade visible: True                                  (was: empty feed)
feed post visible: True                              (was: not visible in 1.5s)
agent detail OK: True
rate limit 429s: 5 (after 30 burst)                  (was: 0 — no limit at all)
```

## Remaining (Phase 4 work)

- Server-Sent Events vs WebSocket (deferred; WS works).
- Time-series ring buffer for the wealth panel.
- Postgres-backed secret store for HMAC rotation.

These are real follow-ups but not blocking the demo.
