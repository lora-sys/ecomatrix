# TESTING.md — Test Strategy & Evidence Format

> The Evidence gate is non-negotiable. If you cannot prove it works, it is not done.

## 1. Testing Pyramid

```
        E2E (Playwright, agent-browser)
       ─────────────────────────────────
      Integration (Go testcontainers, Python docker-py)
     ─────────────────────────────────────────
    Unit (Go testify, pytest, Vitest)
   ─────────────────────────────────────────────
```

Coverage is a guide, not a goal. We test **the seams that fail under concurrency** (row-level locks, idempotency, ws backpressure) and **the contracts** (A2A codec, API schemas).

## 2. Per-Layer Rules

### Backend (Go)

- `go test -race ./...` is the gate.
- Unit: domain logic, A2A codec, repo with sqlite/real pg.
- Integration: spin up real Postgres (testcontainers or local dev DB on `ecomatrix_test`), exercise HTTP + WS end-to-end.
- Concurrency test for `POST /v1/trades`: 50 goroutines racing to spend the same balance — only the affordable ones settle; never negative balance.
- Cover cases: insufficient funds (HTTP 409), unknown agent (404), protocol mismatch (400), duplicate `msg_id` (200 idempotent replay or 409, decided in spec).

### Frontend (Next.js)

- Vitest for component logic + hooks.
- Playwright for E2E. Required scenarios per Phase 3:
  - Dashboard loads and shows ≥ 1 live KPI.
  - Clicking an agent opens the detail panel.
  - WS reconnect after server restart.
- Screenshots: desktop (1440×900) and mobile (390×844). No screenshot if console has errors.

### AI / Python

- pytest for graph nodes in isolation (mock LLM responses).
- One scenario test: two agents end-to-end via the Go backend with stub LLM.

### Database

- Migration up + down in CI.
- Seed script idempotent (re-running yields same row counts).

## 3. Evidence Pack (per PR)

Located at `docs/evidence/<issue-id>/`. Required files:

- `change-summary.md` — what shipped, in plain prose (≤ 1 page).
- `verification.md` — exact commands run + their outcomes (paste, don't summarize).
- `test-results/` — `go test -race` output, Playwright JSON, etc.
- `screenshots/` — only for UI changes.
- `review-report.md` — aggregated reviewer findings + disposition.

Template: `templates/evidence-pack.md`.

## 4. Test Naming

- Go: `TestThing_Scenario_ExpectedBehavior` (e.g. `TestTrade_InsufficientFunds_Returns409`).
- TS/Vitest: `describe('Thing') > it('does X when Y')`.
- Pytest: `def test_<thing>_<expected>():`.

## 5. Flake Policy

- A flaky test is a **P0 bug**. Tag `@flaky` in a TODO + file an Issue; do not skip in CI.
- Playwright tests must `await` UI state, not `sleep`.

## 6. Evidence Gate Checklist (must pass to merge)

- [ ] All required CI jobs green.
- [ ] `go test -race ./...` clean for backend changes.
- [ ] Playwright green + screenshots attached for UI changes.
- [ ] Migration up + down exercised locally.
- [ ] ≥ 2 reviewer reports in `docs/evidence/<id>/review-report.md`.
- [ ] `change-summary.md` + `verification.md` present and accurate.
- [ ] No `TODO` without an Issue link.

## 7. Test Data

- Seed users: `agent_miner_01`..`agent_miner_05`, `agent_merchant_01`..`03`, `agent_hacker_01`..`02`, `agent_mediator_01`.
- Each seeded with a known balance so tests are deterministic.
- Seed script: `apps/backend/cmd/seed/main.go` (separate binary).
