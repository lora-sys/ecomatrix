# ISSUE-009 — Health/Readiness + structured logging

| Field         | Value                                                                          |
| ------------- | ------------------------------------------------------------------------------ |
| Status        | Todo                                                                           |
| Owner         | backend                                                                        |
| Branch        | `feature/ISSUE-009-observability`                                              |
| Phase         | 1                                                                              |
| Reviewers     | architecture-reviewer                                                          |
| Evidence dir  | `docs/evidence/ISSUE-009/`                                                     |

## Goal
`/healthz` (liveness, no DB) and `/readyz` (readiness, DB ping); one structured log line per HTTP request.

## Acceptance Criteria
- [ ] `curl /healthz` returns 200 always.
- [ ] `curl /readyz` returns 200 with DB; 503 with DB down (test by closing pool).
- [ ] Each request emits exactly one JSON log line.

## Allowed Files
- `apps/backend/internal/transport/http/health.go`
- `apps/backend/internal/observability/**`
