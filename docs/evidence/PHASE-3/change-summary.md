# Phase 3 — Change Summary

## What
The Next.js dashboard is live. The "God's Eye" view at `/` shows four KPI tiles (live agent count, total GOLD, recent QPS, WS connections), a wealth distribution chart with the `grad-wealth` color scale, a live trade broadcast feed driven by the WebSocket hub, an agent list, and four job-type 3D cards. The agent detail page at `/agents/{string_id}` shows vitals + recent trades with the tracing-beam component.

## Why
PRD §7 calls for the **God's Eye** in Week 3: a dashboard that observes the live economy without participating in it. Without Phase 3, the agent economy runs but has no observer.

## Notable Decisions
- **Port 3100 for the dev server** — port 3000 is already in use on this host by an unrelated project. The Playwright config reads `PLAYWRIGHT_PORT` so the e2e tests point at the right port without code changes.
- **Aceternity components authored locally** — Aceternity UI is not packaged on npm; the components (GlowingCard, TracingBeam, ThreeDCard) follow the visual language described in `DESIGN.md` and live in `apps/frontend/components/`. No upstream drift to track.
- **WebSocket reconnect with exponential backoff** — `use-economatrix-stream` reconnects with `min(30s, 1s × 2^n)` + ±20% jitter; never spams the server, never gives up.
- **Value damping for KPI tiles** — counters animate over ~220 ms (τ) toward the latest value rather than flash-cutting, per `DESIGN.md §6`.
- **RSC for first paint, client islands for live updates** — the page is rendered server-side via `app/page.tsx` (calls `fetchAgents()` at request time), then `LiveProvider` mounts the WS hook and the polling loop. This keeps the initial bundle small (~151 KB First Load JS) and the time-to-interactive low.
- **CORS middleware on the Go backend** — permissive in dev; tighten before production.
- **New `GET /v1/metrics` endpoint** — gives the dashboard a single source of truth (agent_count, total_gold, jobs_breakdown, recent_qps, ws_connections, last_trade_at) instead of forcing it to re-derive aggregates from `/v1/agents` on every render.

## Out of Scope
- `recharts` was specified in `DESIGN.md` for the wealth chart; MVP uses a denser inline bar visualization (sort + horizontal bars + `grad-wealth` colors) because it scans faster on the dashboard.
- Authentication on the dashboard — Phase 1 ships a shared admin token for admin endpoints; the dashboard is read-only and unauthenticated for now.
- Per-agent HMAC tokens (ISS-015 follow-up).
- Postgres-backed long-term memory for agents (ISS-015 follow-up).
- WebSocket subscription filtering by job type (Phase 4 polish).

## Files Shipped (Phase 3)

```
apps/backend/
├── internal/service/metrics.go            # NEW: MetricsService + NoteTrade + Collect
├── internal/service/metrics_test.go       # NEW: 3 tests
├── internal/service/trade.go              # wire NoteTrade() after settled
├── internal/transport/http/router.go      # +getMetrics +corsMiddleware
└── cmd/server/main.go                     # wire MetricsService

apps/frontend/                              # NEW
├── package.json, tsconfig.json, next.config.mjs, tailwind.config.ts,
│   postcss.config.mjs, playwright.config.ts, .env.example
├── app/
│   ├── layout.tsx, page.tsx, dashboard-client.tsx
│   └── agents/[id]/page.tsx
├── components/
│   ├── glowing-card.tsx, tracing-beam.tsx, three-d-card.tsx
│   ├── kpi-tile.tsx, wealth-chart.tsx, trade-feed.tsx
│   └── live-provider.tsx
├── hooks/store.ts, hooks/use-economatrix-stream.ts
├── lib/{api,types,damping}.ts
├── messages/zh-CN.json
├── styles/globals.css
├── e2e/dashboard.spec.ts
└── test-results/                           # Playwright artifacts
```
