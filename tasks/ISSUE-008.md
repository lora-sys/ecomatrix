# ISSUE-008 — WebSocket hub + /v1/stream

| Field         | Value                                                                          |
| ------------- | ------------------------------------------------------------------------------ |
| Status        | Todo                                                                           |
| Owner         | backend                                                                        |
| Branch        | `feature/ISSUE-008-ws-hub`                                                     |
| Phase         | 1                                                                              |
| Reviewers     | architecture-reviewer, bug-hunter                                              |
| Evidence dir  | `docs/evidence/ISSUE-008/`                                                     |

## Goal
Broadcast `trade.settled` / `trade.rejected` / `agent.heartbeat` events to all connected dashboards.

## Implementation Plan
1. `internal/transport/ws/hub.go`: per-conn buffered channel, fan-out via central broker.
2. Hook the trade service to publish events after commit.
3. Heartbeat every 20 s.
4. Backpressure: drop slow consumers, never block publisher.
5. Test: 10 connections receive 1 event each from a single trade.

## Acceptance Requirements
- [ ] WebSocket smoke test green.
- [ ] Memory doesn't grow unbounded over 1000 events.

## Allowed Files
- `apps/backend/internal/transport/ws/**`
- `apps/backend/internal/service/trade.go` (publish hook)
