# ISSUE-004 — DB schema (agents, transactions, social_feeds)

| Field         | Value                                                                          |
| ------------- | ------------------------------------------------------------------------------ |
| Status        | Todo                                                                           |
| Owner         | database                                                                       |
| Branch        | `feature/ISSUE-004-db-schema`                                                  |
| Phase         | 1                                                                              |
| Reviewers     | architecture-reviewer, security-reviewer                                       |
| Evidence dir  | `docs/evidence/ISSUE-004/`                                                     |

## Goal
Apply the schema from `docs/architecture/db.md` as migration `0001_init`.

## Implementation Plan
1. Write `apps/backend/migrations/0001_init.up.sql` and `.down.sql` matching the spec verbatim.
2. Add `apps/backend/cmd/seed/main.go` with seeded agents.
3. Add a `seed` Make target.

## Acceptance Criteria
- [ ] Tables created with all CHECK constraints.
- [ ] Seed inserts 5 miners, 3 merchants, 2 hackers, 1 mediator.
- [ ] Seed is re-runnable.

## Evidence Requirements
- psql `\d agents`, `\d transactions`, `\d social_feeds` output captured.
- Seed run output.

## Allowed Files
- `apps/backend/migrations/**`
- `apps/backend/cmd/seed/**`
- `apps/backend/Makefile`
