# Phase 3 — Verification

## Environment
- Node.js 24, Next.js 15.1.0 (App Router), React 19.
- Playwright Chromium 1229 for E2E.
- Backend on `:8080` against Postgres `ecomatrix`.

## 1. Backend Metrics

`GET /v1/metrics` returns a single Snapshot:

```json
{
  "agent_count": 13,
  "total_gold": 2560,
  "jobs_breakdown": {"hacker": 2, "mediator": 1, "merchant": 4, "miner": 6},
  "recent_qps": 0,
  "ws_connections": 0,
  "generated_at": "2026-07-11T06:15:03.399255459Z"
}
```

After 5 trades fired in a tight burst:

```json
{
  "agent_count": 13,
  "total_gold": 2560,
  "jobs_breakdown": {...},
  "recent_qps": 5,
  "ws_connections": 0,
  "last_trade_at": "2026-07-11T06:15:03.591823452Z",
  "generated_at": "2026-07-11T06:15:03.60076276Z"
}
```

Files: [`curl/metrics-initial.{json,txt}`](./curl/metrics-initial.txt), [`curl/metrics-after-burst.{json,txt}`](./curl/metrics-after-burst.txt).

## 2. Go Backend Suite (regression)

After wiring MetricsService + CORS middleware:

```
$ cd apps/backend && go test -race -count=1 ./...
ok  github.com/ecomatrix/backend/internal/service  1.97s
ok  github.com/ecomatrix/backend/pkg/a2a         1.03s
```

Full output: [`test-results/go-test-race.txt`](./test-results/go-test-race.txt). The 50-goroutine concurrency proof still passes (33 settled, 17 rejected, no negative balances).

## 3. Frontend Build

```
$ cd apps/frontend && npx next build
✓ Compiled successfully
Route (app)                  Size     First Load JS
┌ ƒ /                        4.28 kB  151 kB
├ ○ /_not-found              982 B    107 kB
└ ƒ /agents/[id]             1.89 kB  149 kB
+ First Load JS shared       106 kB
```

First-load JS is **151 KB** (within the 250 KB budget in `ENGINEERING.md §10`).

Full build output: [`build/next-build.txt`](./build/next-build.txt). Typecheck log: [`build/typecheck.txt`](./build/typecheck.txt) (zero errors under `strict: true`).

## 4. Playwright E2E

```
$ npx playwright test
✓ desktop › renders KPI tiles + chart + feed on desktop (1.7s)
✓ desktop › agent detail page renders vitals + recent trades (2.8s)
✓ mobile  › renders KPI tiles + chart + feed on desktop (1.9s)
✓ mobile  › agent detail page renders vitals + recent trades (902ms)
4 passed (9.0s)
```

Each test asserts the dashboard is visible, the KPI labels render, and there are zero console errors (which caught the missing CORS middleware before CORS was added — see Review Report).

Full log: [`test-results/playwright.log`](./test-results/playwright.log).

## 5. Visual Evidence

- Desktop dashboard (1440×900): [`screenshots/dashboard-desktop.png`](./screenshots/dashboard-desktop.png)
- Mobile dashboard (390×844): [`screenshots/dashboard-mobile.png`](./screenshots/dashboard-mobile.png)
- Desktop agent detail: [`screenshots/agent-desktop.png`](./screenshots/agent-desktop.png)
- Mobile agent detail: [`screenshots/agent-mobile.png`](./screenshots/agent-mobile.png)

The desktop screenshot shows ≥ 3 live signals above the fold: 4 KPI tiles, wealth distribution chart with the `grad-wealth` color scale, trade broadcast placeholder, and agent list. The agent detail page shows vitals (Balance 100, Vitality 80, Credit 60) and 5 recent SETTLED trades via the tracing-beam component.

## 6. WebSocket End-to-End

The dashboard subscribes to `ws://127.0.0.1:8080/v1/stream`. The trade.settled fan-out was verified during Phase 1 ([Phase 1 curl/ws-smoke.txt](../PHASE-1/curl/ws-smoke.txt)) and is reused here as [`curl/ws-smoke-cross-ref.txt`](./curl/ws-smoke-cross-ref.txt). The new `LiveProvider` client uses the same endpoint with reconnect-with-backoff.

## Result

**PASS** — Phase 3 exit criteria met. The dashboard renders against a live backend, the WebSocket integrates, screenshots are captured at desktop + mobile viewports, and the build stays inside the JS budget.
