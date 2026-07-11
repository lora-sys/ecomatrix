# ISSUE-010 — Seed + Make + docker compose

| Field         | Value                                                                          |
| ------------- | ------------------------------------------------------------------------------ |
| Status        | Todo                                                                           |
| Owner         | backend                                                                        |
| Branch        | `feature/ISSUE-010-dev-ergonomics`                                             |
| Phase         | 1                                                                              |
| Reviewers     | architecture-reviewer                                                          |
| Evidence dir  | `docs/evidence/ISSUE-010/`                                                     |

## Goal
One-command dev loop: `docker compose up -d db && make seed && make run`.

## Implementation Plan
1. `docker-compose.yml`: Postgres 16 alpine, port 5432, named volume.
2. `Makefile`: targets `db-up`, `db-down`, `migrate-up`, `migrate-down`, `seed`, `run`, `test`, `lint`.
3. README snippet documenting the loop.

## Acceptance Criteria
- [ ] Fresh clone → `make dev` brings up a working backend + seeded DB.

## Allowed Files
- `docker-compose.yml`
- `apps/backend/Makefile`
- `apps/backend/README.md`
