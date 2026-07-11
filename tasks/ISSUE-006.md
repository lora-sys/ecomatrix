# ISSUE-006 — Trade API with row-level lock + idempotency

| Field         | Value                                                                          |
| ------------- | ------------------------------------------------------------------------------ |
| Status        | Todo (crown jewel of Phase 1)                                                   |
| Owner         | backend                                                                        |
| Branch        | `feature/ISSUE-006-trade-api`                                                  |
| Phase         | 1                                                                              |
| Reviewers     | bug-hunter, architecture-reviewer, security-reviewer                           |
| Evidence dir  | `docs/evidence/ISSUE-006/`                                                     |

## Goal
`POST /v1/trades` settles an A2A `EXECUTE_TRADE` atomically.

## Implementation Plan
1. `internal/domain/trade.go`: `TradeRequest`, `Receipt`, `Error` types.
2. `internal/service/trade.go`: opens tx, locks two agent rows in ascending id order, validates, updates balances, inserts transaction row.
3. `internal/transport/http/trades.go`: parses A2A envelope via `pkg/a2a`, calls service, maps errors.
4. Tests: happy path, insufficient funds, unknown agent, duplicate `msg_id` (idempotent replay), self-trade rejection.

## Acceptance Criteria
- [ ] All five test cases pass under `-race`.
- [ ] Duplicate `msg_id` returns original receipt with HTTP 200.
- [ ] Balance can never go negative (test asserts CHECK + service guard).
- [ ] Lock order documented in code comment.

## Evidence Requirements
- Test output (`go test -race -v ./internal/service/...`).
- curl trace showing a successful and a rejected trade.

## Allowed Files
- `apps/backend/internal/**`
