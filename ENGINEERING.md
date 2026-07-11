# ENGINEERING.md — Engineering Rules

> Cross-cutting rules for code, schema, API, and Git. Per-app detail lives in `docs/architecture/*.md`.

## 1. Languages & Tooling

| Domain  | Tooling                                              |
| ------- | ---------------------------------------------------- |
| Go      | Go 1.26+, gofmt, goimports, golangci-lint, testify   |
| TS/Next | TypeScript 5.x strict, ESLint (next/core-web-vitals), Prettier |
| Python  | Python 3.12, ruff, mypy, pytest                      |
| SQL     | Plain SQL migrations in `apps/backend/migrations/`   |

All three languages are pinned in `apps/*/.tool-versions` (mise/asdf) or equivalent.

## 2. Backend (Go) — `apps/backend`

- **Framework:** Fiber v2. WebSocket via `gofiber/contrib/websocket`.
- **Layout:**
  ```
  apps/backend/
  ├── cmd/server/main.go
  ├── internal/
  │   ├── config/        # env loading, validation
  │   ├── domain/        # entities, value objects, errors
  │   ├── service/       # business logic, no Fiber/HTTP imports
  │   ├── transport/http/ # handlers, middleware, route registration
  │   ├── transport/ws/   # websocket hub
  │   ├── repo/          # GORM repositories, SQL is fine where needed
  │   └── migrations/    # embed.FS for SQL files
  ├── pkg/a2a/           # protocol codec (shared with Python agent)
  └── test/              # integration tests
  ```
- **Concurrency:** every state-mutating endpoint must take a row-level lock (`SELECT ... FOR UPDATE`) inside a transaction. The PRD forbids double-spend; this is the line of defense.
- **Errors:** typed sentinel errors in `internal/domain/errors.go`. HTTP layer maps to status codes; never leak raw SQL errors to clients.
- **Logging:** `slog` JSON handler. One log line per request with `request_id`, `agent_id` (if any), `latency_ms`.
- **WebSocket hub:** per-connection buffered channel (size 64), slow-consumer must drop, never block publisher.
- **Test:** `testify/assert` + `testify/require`. Run race detector in CI: `go test -race ./...`.

## 3. Frontend (Next.js) — `apps/frontend`

- **App Router only.** No pages router.
- **RSC by default.** Only mark `"use client"` when the component needs state, effects, or browser APIs. WebSocket subscriptions live in client components.
- **State:** server components fetch via fetch/RSC; client state via `zustand` (single store per feature). No Redux.
- **Styling:** Tailwind only. Component primitives from Aceternity UI; never modify upstream files — wrap locally.
- **Data fetching:** typed via `zod` schemas shared with the backend response shapes (codegen optional, manual OK in MVP).
- **Testing:** Vitest for unit, Playwright for E2E. Required for any new screen.

## 4. AI Layer (Python) — `apps/agent`

- **Framework:** LangGraph. Each agent is a graph; state is a typed `TypedDict`.
- **LLM provider:** abstract behind `apps/agent/ecomatrix/llm.py`. Default OpenAI-compatible; pluggable.
- **Memory:** short-term in graph state; long-term summaries in Postgres (`agents.long_term_memory JSONB`).
- **Transport:** HTTP only in Phase 2. JSON shape must match `docs/architecture/api.md`.

## 5. Database — Postgres

- One schema: `public`. App uses dedicated DB (`ecomatrix`) and user (`repotwin` in dev only).
- **Naming:** `snake_case`, plural tables (`agents`, `transactions`, `social_feeds`).
- **IDs:** `bigserial` primary keys; plus a `string_id` (e.g. `agent_miner_01`) for the A2A layer.
- **Money:** store as `BIGINT` (satoshis-style). Never `FLOAT`. Currency type is an enum column.
- **Time:** `timestamptz` everywhere. UTC at rest.
- **Migrations:** forward-only with paired down file. CI runs `up` against an ephemeral DB.
- **Forbidden:** raw ORM auto-migrate in production code paths. Migrations are reviewed.

## 6. API — A2A Protocol v1.1

Authoritative spec: `docs/architecture/api.md`. Highlights:

- All A2A messages are JSON, UTF-8, snake_case.
- Server enforces `protocol_v == "1.1"`. Unknown actions → HTTP 400 with structured error.
- Trade endpoints must be **idempotent** by `msg_id` (unique constraint on `transactions.msg_id`).

## 7. Git Workflow

- Branch naming: `feature/ISSUE-<id>-<slug>` (e.g. `feature/ISSUE-007-trade-row-lock`).
- Commit format (Conventional Commits): `type(scope): description (#issue)`.
- Squash-merge into `main`. PR title becomes commit subject.
- Rebase your branch on `main` before requesting review.
- Tag releases: `v0.1.0` (Phase 1), `v0.2.0` (Phase 2), `v0.3.0` (Phase 3).

## 8. Code Review

Minimum 2 reviewers per PR (the harness's adversarial review). Reviewer scopes:

- `backend` → bug-hunter + architecture-reviewer
- `frontend` → behavior-reviewer + ui-reviewer
- `database` → architecture-reviewer + security-reviewer (if PII/secrets)
- `cross-cut` → all three

Reviewer reports land in `docs/evidence/<issue>/review-report.md`. Aggregator file Fix Tasks back to owner.

## 9. Observability

- `/healthz` (liveness) and `/readyz` (readiness, includes DB ping) on the backend.
- Structured logs to stdout; never log full A2A payloads (PII risk). Mask `agents.credit_score` and balances above a threshold in non-prod logs.
- Metrics: QPS, in-flight trades, ws connections. Phase 3 adds Prometheus.

## 10. Performance Budgets

- API p95 latency < 80 ms for trade endpoints under 100 RPS.
- WebSocket end-to-end tick → render < 250 ms.
- Frontend initial JS bundle < 250 KB gzipped (per route).

## 11. Forbidden Engineering Practices

- ❌ `panic` for control flow in Go services.
- ❌ `interface{}` / `any` in domain code (use generics or typed structs).
- ❌ `as` casts in TypeScript outside test files.
- ❌ Auto-migrate schema in production.
- ❌ Skipping row-level locks "for performance" — fix the lock instead.
- ❌ Long-lived feature branches (> 5 days) — split the Issue.
