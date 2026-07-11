# ISSUE-002 — A2A protocol codec + error envelope

| Field         | Value                                                                          |
| ------------- | ------------------------------------------------------------------------------ |
| Status        | Todo                                                                           |
| Owner         | backend                                                                        |
| Branch        | `feature/ISSUE-002-a2a-codec`                                                  |
| Phase         | 1                                                                              |
| Reviewers     | bug-hunter, architecture-reviewer                                              |
| Evidence dir  | `docs/evidence/ISSUE-002/`                                                     |

## Context
The A2A protocol (v1.1) is the contract between Python agents and the Go gateway, and between the gateway and the WebSocket fan-out. A single canonical codec prevents drift.

## Goal
Implement `pkg/a2a` (Go) with: envelope validate, action dispatcher stub, typed errors, JSON fixtures for tests.

## Non-Goals
- HTTP transport (ISSUE-005/006).
- WebSocket frames (ISSUE-008).

## Implementation Plan
1. Add `pkg/a2a/envelope.go`: types + JSON tags matching `docs/architecture/api.md` exactly.
2. Add `pkg/a2a/codec.go`: `Validate(env Envelope) error` checks `protocol_v`, `msg_id` regex, action allow-list.
3. Add `pkg/a2a/errors.go`: sentinel errors mapped to HTTP codes by `transport/http`.
4. Add `pkg/a2a/fixtures/testdata/*.json` for valid + invalid envelopes (shared with Python later).
5. Unit tests: `codec_test.go` covering happy path + every error code.

## Acceptance Criteria
- [ ] `go test ./pkg/a2a/...` covers: ok, protocol mismatch, missing msg_id, unknown action, self-trade (separate concern, just stub), bad timestamp.
- [ ] Codec rejects `protocol_v != "1.1"` with sentinel `ErrProtocolMismatch`.
- [ ] No external deps beyond `encoding/json` + stdlib regex.

## Evidence Requirements
- `go test -race -v ./pkg/a2a/...` output in `docs/evidence/ISSUE-002/test-results/`.
- `change-summary.md` + `verification.md`.

## Allowed Files
- `apps/backend/pkg/a2a/**`
- `apps/backend/go.mod`, `apps/backend/go.sum`

## Related Docs
- `docs/architecture/api.md` §1, §2
