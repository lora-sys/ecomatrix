# ISSUE-007 — Concurrency test (50 goroutines racing)

| Field         | Value                                                                          |
| ------------- | ------------------------------------------------------------------------------ |
| Status        | Todo                                                                           |
| Owner         | qa                                                                             |
| Branch        | `feature/ISSUE-007-concurrency-test`                                          |
| Phase         | 1                                                                              |
| Reviewers     | bug-hunter                                                                     |
| Evidence dir  | `docs/evidence/ISSUE-007/`                                                     |

## Goal
Prove the trade API is double-spend safe under load.

## Implementation Plan
1. Seed one agent with balance 1000 GOLD.
2. Spawn 50 goroutines, each submits a trade of 30 GOLD to the same target.
3. Wait for all, then assert: settled count ≤ 33 (1000/30), rejected count = remainder, balance on target equals settled×30.

## Acceptance Criteria
- [ ] Race-detector clean.
- [ ] Invariant holds across 5 reruns.
- [ ] Output captured in `test-results/`.

## Allowed Files
- `apps/backend/test/concurrency/**`
