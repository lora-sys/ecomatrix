# ADR-0003: Dashboard reads metrics from /v1/metrics, not /v1/agents

- **Status:** Accepted
- **Date:** 2026-07-11
- **Deciders:** Coordinator (Codex)

## Context
The dashboard needs three aggregates on every render: total GOLD, agent count, and QPS. With only `/v1/agents` available, the dashboard would have to scan the full agent list (200 rows) and recompute on every render. With `/v1/metrics`, the backend computes the aggregate once per request.

## Decision
Add `GET /v1/metrics` to the Go backend. The dashboard polls it every 3 seconds when the WebSocket is connected, every 1.5 seconds otherwise. WebSocket events (`trade.settled`, `trade.rejected`) update the feed and bump `recent_qps` via the server-side `NoteTrade()` call.

## Consequences
Positive:
- Single source of truth for the dashboard's KPI panel.
- Server-side `recent_qps` is more accurate than client-side deltas.
- `/v1/metrics` is the obvious extension point for Phase 4 (Prometheus exporter).

Negative:
- One more endpoint to maintain and document.
- Polling introduces a 3-second lag window even with the WS connected. Acceptable for an observatory.

Neutral:
- Future: when /v1/metrics grows (p99 latency, error rate, etc.), split into a Prometheus exposition format alongside the JSON view.

## Alternatives Considered
- **Server-Sent Events** instead of WebSocket + polling. Rejected — WS gives us a richer event model (rejected trades, replays) for the feed.
- **Recompute aggregates client-side.** Rejected — wastes CPU on every render and gives stale totals between WS events.

## References
- `apps/backend/internal/service/metrics.go`
- `apps/frontend/hooks/store.ts`, `components/live-provider.tsx`
