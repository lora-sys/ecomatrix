---
name: Feature
about: New user-facing capability
title: "[FEATURE] ISSUE-<id> <short slug>"
labels: feature,phase-1
assignees: ''
---

## Context
Why this matters; link the PRD section or roadmap line.

## Goal
One sentence outcome.

## Non-Goals
What is **not** in scope.

## Implementation Plan
High-level steps (≤ 8).

## Acceptance Criteria
- [ ] Testable criterion 1
- [ ] Testable criterion 2
- [ ] ...

## Evidence Requirements
- Backend: `go test -race ./...` + curl trace
- DB: migration up + down
- Frontend: Playwright screenshots desktop + mobile

## Reviewer Requirements
- [ ] bug-hunter
- [ ] architecture-reviewer
- [ ] (security-reviewer if applicable)

## Related Docs
- docs/architecture/api.md §
- docs/architecture/backend.md §
- docs/architecture/db.md §

## Allowed Files (allow-list)
- apps/backend/internal/...
- apps/backend/migrations/...

## Linked Roadmap Item
- Phase X, milestone "..."
