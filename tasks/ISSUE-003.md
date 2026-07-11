# ISSUE-003 — Postgres connection + migrations runner

| Field         | Value                                                                          |
| ------------- | ------------------------------------------------------------------------------ |
| Status        | Todo                                                                           |
| Owner         | database                                                                       |
| Branch        | `feature/ISSUE-003-db-migrations`                                             |
| Phase         | 1                                                                              |
| Reviewers     | architecture-reviewer, security-reviewer                                       |
| Evidence dir  | `docs/evidence/ISSUE-003/`                                                     |

## Goal
Wire GORM to the dev Postgres at `localhost:5432/ecomatrix` with a forward-only SQL migrations runner and a `make migrate-up` / `make migrate-down` target.

## Non-Goals
- Schema content (ISSUE-004).

## Implementation Plan
1. Add `internal/config/config.go`: env loading with defaults.
2. Add `internal/repo/db.go`: GORM open with pool tuning (`SetMaxOpenConns(50)`, `SetMaxIdleConns(10)`, `SetConnMaxLifetime(30m)`).
3. Add `internal/migrations/runner.go`: reads `migrations/NNNN_*.{up,down}.sql`, tracks applied versions in `schema_migrations(version INT PK, applied_at TIMESTAMPTZ)`.
4. Add Make targets.

## Acceptance Criteria
- [ ] `make migrate-up` against a fresh `ecomatrix_test` DB succeeds.
- [ ] `make migrate-down` reverses cleanly.
- [ ] `make migrate-up` is idempotent (second run is a no-op).

## Evidence Requirements
- `verification.md` with both command outputs.
- Migration runner unit test.

## Allowed Files
- `apps/backend/internal/config/**`
- `apps/backend/internal/repo/db.go`
- `apps/backend/internal/migrations/**`
- `apps/backend/Makefile`
