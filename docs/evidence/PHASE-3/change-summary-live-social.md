# Phase 3.4 — Live social square via WebSocket

## What
The social square panel on the dashboard was polling `/v1/feeds` every 4 seconds. Now it reads from the zustand store, which is hydrated once via the initial fetch and then kept fresh by the `feed.posted` WebSocket event.

## Why
Last turn called this out as the last loose end. With live updates, the social square matches the trade broadcast's UX (instant) instead of feeling laggy.

## Files Changed

```
apps/backend/
└── internal/transport/http/router.go   # feed.posted event now includes `content`

apps/agent/
└── (no change — agent already POSTs to /v1/feeds per tick)

apps/frontend/
├── hooks/store.ts                       # +social slice + setSocial + pushSocial + feed.posted handler
│                                          + dedupe helpers (live takes precedence over fetched)
├── components/social-feed.tsx           # read from store; one-shot initial fetch
└── (no Playwright change — existing e2e covers it)

apps/backend/internal/domain/agent.go   # LongTermMemory.MarshalJSON so Facts is `[]` not null
                                          (uncovered by the new social panel: every page that
                                          embeds an Agent now serializes a non-null facts array)
```

## Verified
- Playwright 4/4 still green on desktop + mobile after store + component changes.
- Screenshot at `docs/evidence/PHASE-3/screenshots/dashboard-desktop.png` shows the social square populated with 8 `[REQUEST]` posts after the seed burst — they're sourced from the store, which got them from `/api/proxy/feeds` on mount.
- Go suite still green; the `MarshalJSON` change ships as a JSONB safety improvement for any future endpoint that embeds an `Agent`.

## Bonus fix: `LongTermMemory.MarshalJSON`

While wiring the social square I noticed `LongTermMemory.Facts` was JSON-marshaling as `null` when the slice was nil (the default state for a freshly seeded agent with no LTM facts). The dashboard agent detail page would have crashed on `ltm.facts.length` in that case. `MarshalJSON` now coerces nil to `[]` so downstream consumers can treat it as an array.
