# ISSUE-005 — Agent CRUD endpoints

| Field         | Value                                                                          |
| ------------- | ------------------------------------------------------------------------------ |
| Status        | Todo                                                                           |
| Owner         | backend                                                                        |
| Branch        | `feature/ISSUE-005-agent-crud`                                                 |
| Phase         | 1                                                                              |
| Reviewers     | bug-hunter, architecture-reviewer                                              |
| Evidence dir  | `docs/evidence/ISSUE-005/`                                                     |

## Goal
Expose `GET /v1/agents`, `GET /v1/agents/{id}`, `POST /v1/agents`.

## Implementation Plan
1. `internal/domain/agent.go`: `Agent` entity.
2. `internal/repo/agent_repo.go`: GORM queries.
3. `internal/transport/http/agents.go`: Fiber handlers.
4. Admin token middleware.
5. Validation: `job_type` ∈ allowed set; `string_id` matches regex; balance ≥ 0.
6. Tests: list, get, create (success + 401 + 400).

## Acceptance Criteria
- [ ] curl trace for each endpoint.
- [ ] Admin token required for POST.
- [ ] GET returns 404 for unknown id.

## Evidence Requirements
- Curl trace + test output.

## Allowed Files
- `apps/backend/internal/**`
