# Session — ISS-030 supervisor detail and agent participation

- Started: 2026-07-14T09:00:00+08:00
- Coordinator: Codex
- Branch: `feature/ISSUE-030-supervisor-detail`
- Phase: 10 / Phase 8 supervisor follow-up

## Plan

1. Land the new `ByID` / `ByAgent` repo helpers and the
   `GET /v1/supervisor/runs/:id` plus
   `GET /v1/agents/by-string-id/:sid/supervisor-runs` routes.
2. Extend `MetricsService.Snapshot` with supervisor counters.
3. Add the frontend `/supervisor/[id]` page and the per-agent history card.
4. Wire `<SupervisorLog>` to navigate to the detail page.
5. Green up all local gates, capture Playwright screenshots, and record
   evidence.

## Outcomes

- Backend: `SupervisorRunsRepo.ByID` returns `ErrSupervisorRunNotFound` for
  unknown ids. `ByAgent` uses `JSONB @>` over `worker_results` with an
  inlined literal. `MetricsService.Collect` adds a single
  `SELECT COUNT(*), MAX(finished_at)` against `supervisor_runs`.
- HTTP: two new endpoints registered with the correct order
  (`/supervisor/runs/:id` ahead of any prefix collision); both return
  404 for unknown resources, 400 on bad ids, 503 if the supervisor repo
  is not wired.
- Python: `A2AClient.fetch_supervisor_run` /
  `A2AClient.list_agent_supervisor_runs` mirror the new routes, with
  `A2AError` carrying the upstream HTTP status.
- Frontend: `/supervisor/[id]/page.tsx` SSR-fetches and renders the run;
  the `<SupervisorLog>` panel exposes `data-supervisor-link` anchors
  both in the "latest" block and the historical list, and the
  `<AgentSupervisorHistory>` list is now mounted inside
  `/agents/[id]/page.tsx` after a parallel fetch.
- Tests:
  - Backend: 6 new test cases (3 in `repo`, 2 in `service`, 1 router
    integration test per handler via the new
    `router_supervisor_test.go`).
  - Agent: 4 new test cases in `test_supervisor_detail.py` covering 200,
    404, query param wiring, empty key fallback.
  - Frontend: 2 new Playwright tests in `supervisor.spec.ts` for both
    viewports. Full suite: **18 passed, 2 skipped** locally.

## Handoff

- Branch has all uncommitted changes staged in the working tree; commit
  and PR push to be the next step.
- Evidence: `docs/evidence/ISSUE-030/{change-summary,verification,review-report}.md`
  plus 6 PNGs covering desktop + mobile renderings.
