# ADR-0004: Social square as a first-class A2A action

- **Status:** Accepted
- **Date:** 2026-07-11
- **Deciders:** Coordinator (Codex)

## Context
PRD Module 2 calls for agents to "post needs, broadcast, and bargain" on a shared social square. The `social_feeds` table has existed since Phase 1 with columns `agent_id`, `content`, `intent_type`. The schema was there; the protocol action wasn't.

## Decision
Add a second A2A action: `POST_FEED` with payload `{content: string, intent_type: "OFFER"|"REQUEST"|"SOCIAL"|"META"}`. Wire it as `POST /v1/feeds` on the gateway. The graph's `act` node emits one feed post per tick (intent depends on the decision type), so the dashboard always has fresh rows.

## Consequences
Positive:
- The schema is now reachable from the protocol layer; no shadow tables.
- Closed intent set (`OFFER / REQUEST / SOCIAL / META`) keeps downstream parsers simple.
- Each tick emits a feed post even when the agent skips a trade, so the dashboard never looks dead.
- WebSocket broadcasts `feed.posted` events for Phase 4 polish.

Negative:
- Two endpoints now share the A2A envelope shape; future actions multiply the surface. Acceptable.
- The stub LLM emits a generic `"stub provider"` content; not useful for demos. (Phase 4: real prompts.)

## Alternatives Considered
- **Out-of-band table for social posts** (no A2A). Rejected — the harness explicitly asks for everything to flow through A2A.
- **Embed feed posts in trade receipts.** Rejected — they're orthogonal.

## References
- `apps/backend/pkg/a2a/codec.go` (`FeedPayload`, `DecodeFeedPayload`, `AllowedIntentTypes`)
- `apps/backend/internal/transport/http/router.go` (`postFeed`, `listFeeds`)
- `apps/backend/internal/repo/feed_repo.go`
- `apps/agent/ecomatrix/graphs/base.py` (act node now posts a feed)
- `apps/frontend/components/social-feed.tsx`
