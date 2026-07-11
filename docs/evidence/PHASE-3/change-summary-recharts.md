# Phase 3.3 — Wealth chart upgrade

## What
Replaced the inline bar visualization in `components/wealth-chart.tsx` with a proper `recharts` AreaChart per `DESIGN.md §8`. The chart now uses:
- `monotone` curve over the sorted top-12 agents.
- A custom `<linearGradient>` fill that fades cyan → gold → rose (the `grad-wealth` scale).
- Job-colored dots (violet/gold/rose/cyan for merchant/miner/hacker/mediator).
- A `<Tooltip>` styled to match the panel's monospace design.
- A footer row showing the world total GOLD from `/v1/metrics`.

## Why
The previous inline bar viz was readable but didn't feel like a "God's Eye" — too spreadsheet-y. recharts was already specified in `DESIGN.md §8`; this aligns the implementation with the design.

## Cost
- Adds `recharts@^2.13.0` to `apps/frontend`.
- Dashboard route bundle: 4.28 KB → 106 KB.
- First-load JS: 151 KB → 253 KB (at the 250 KB budget in `ENGINEERING.md §10`).

## Verified
- `npx tsc --noEmit` clean.
- `npx next build` succeeds.
- Playwright 4/4 still green on desktop + mobile.
- Screenshot at [`screenshots/dashboard-desktop.png`](./screenshots/dashboard-desktop.png) shows the new chart with the wealth gradient, job dots, and the snapshot footer (`TOP 12 GOLD  2,510 GOLD`).
