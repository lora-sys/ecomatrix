# Phase 3 Multi-Scenario — Verification

## 1. Go Backend Suite

```
$ cd apps/backend && go test -race -count=1 ./...
ok  github.com/ecomatrix/backend/internal/service  2.24s   (was 1.97s)
ok  github.com/ecomatrix/backend/pkg/a2a         1.02s
```

20 a2a tests now pass (was 17), including 3 new feed codec tests.

## 2. Python Agent Suite

```
$ cd apps/agent && pytest
============================== 18 passed in 0.36s ==============================
```

Was 14 tests; +4 feed codec parity tests.

## 3. Multi-Agent Scenario Live Run

3 ticks across all 13 seeded agents (5 miners, 3 merchants, 2 hackers, 1 mediator, 2 leftover race_* agents):

```json
{
  "agents": 13,
  "ticks": 3,
  "settled": 37,
  "rejected": 2,
  "feeds_posted": 39,
  "errors": [],
  "world_initial": 2560,
  "world_final": 2560,
  "conservation": true
}
```

Full output: [`scenario-multi/multi-3-ticks.txt`](./scenario-multi/multi-3-ticks.txt). Conservation holds across 37 settled trades.

## 4. /v1/feeds Trace

`GET /v1/feeds?limit=10` returns the most recent posts including OFFER/REQUEST/SOCIAL intents: [`curl/feeds-sample.txt`](./curl/feeds-sample.txt).

## 5. /v1/metrics After Multi-Scenario

```json
{
  "agent_count": 13,
  "total_gold": 2560,
  "recent_qps": 16,
  "ws_connections": 0,
  "last_trade_at": "2026-07-11T06:36:25.426782854Z"
}
```

QPS rose from 0 to 16 during the burst. World GOLD unchanged. Full: [`curl/metrics-after-multi.txt`](./curl/metrics-after-multi.txt).

## 6. Dashboard Screenshot

Updated dashboard shows the new 5-panel layout:

- **KPI row**: 13 agents, 2,520 GOLD, 6.00 QPS, 1 WS connection (live).
- **Wealth chart**: agent_race_target at 960 leads.
- **Trade broadcast**: shows the empty state when no live trade has hit the WS yet.
- **Social square** (NEW): [REQUEST] "需要 50 GOLD 补给体力" from agent_miner_01, [SOCIAL] "广播 · 在线" from mediator/hacker/merchant, [OFFER] "stub provider" from previous multi-scenario tick.
- **Citizen list**: 8 agents with balances.
- **Job-type 3D cards**: miner/merchant/hacker/mediator.

Desktop: [`screenshots/dashboard-desktop.png`](./screenshots/dashboard-desktop.png) — captured at 1440×900.

## 7. Playwright E2E (regression)

4/4 still green on desktop + mobile. The social-feed panel doesn't break the existing assertions (it adds a new panel; KPI labels and "上帝视角" heading still visible).

## Result

**PASS.** Multi-agent scenario runs end-to-end, world conservation holds, social feed renders live, and the dashboard composes into a 5-panel layout. The harness operating system now demonstrates all three PRD modules (God's Eye, A2A Feed, Agent Detail) with live data.
