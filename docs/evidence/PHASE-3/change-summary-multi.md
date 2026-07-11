# Phase 3 — Multi-Scenario Update

## What
Wired the social square end-to-end (`POST /v1/feeds`, A2A `POST_FEED`, dashboard panel) and added `--scenario multi` to the Python runner. The dashboard now renders a 5-panel layout with a live "社交广场" column that shows agents posting OFFER/REQUEST/SOCIAL/META intents.

## Why
PRD Module 2 calls for an "A2A Feed" where agents "post needs, broadcast, and bargain." The `social_feeds` table has been in the schema since Phase 1; this finishes the round-trip. The `--scenario multi` runner lights up every seeded agent simultaneously so the dashboard actually shows continuous activity, not a static demo.

## Files Changed

```
apps/backend/
├── pkg/a2a/envelope.go           # +ActionPostFeed
├── pkg/a2a/codec.go              # +FeedPayload +DecodeFeedPayload +AllowedIntentTypes
├── pkg/a2a/codec_test.go         # +3 tests
├── internal/domain/transaction.go # +FeedPost
├── internal/repo/feed_repo.go    # NEW: FeedRepo
├── internal/transport/http/router.go  # +getFeeds +postFeed +Feed field
└── cmd/server/main.go            # wire Feed

apps/agent/
├── ecomatrix/a2a.py              # +Action.POST_FEED +FeedPayload +decode_feed_payload +post_feed +list_feeds
├── ecomatrix/graphs/base.py      # act node now also posts a feed per tick
├── ecomatrix/runner.py           # +run_multi_agent +--scenario multi (ThreadPoolExecutor)
└── tests/
    ├── test_a2a_codec.py         # +4 parity tests
    └── test_graphs.py            # FakeClient.post_feed + feeds list

apps/frontend/
├── components/social-feed.tsx    # NEW: client component polling /api/proxy/feeds
├── app/api/proxy/feeds/route.ts  # NEW: BFF
├── app/api/proxy/metrics/route.ts # NEW: BFF
└── app/dashboard-client.tsx      # swap feed column for trade+social two-up
```
