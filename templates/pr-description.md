## Summary
- Bullet 1
- Bullet 2
- Bullet 3

Closes #<issue-id>

## Scope
What changed and what didn't.

## Evidence
- `docs/evidence/<id>/change-summary.md`
- `docs/evidence/<id>/verification.md`
- Test command + result:
  ```
  $ go test -race ./...
  ok    apps/backend/internal/service   0.412s
  ```
- Screenshots (UI only):
  - desktop: ![](./docs/evidence/<id>/screenshots/desktop.png)
  - mobile:  ![](./docs/evidence/<id>/screenshots/mobile.png)

## Reviewer Reports
- bug-hunter: docs/evidence/<id>/review-bug-hunter.md
- architecture-reviewer: docs/evidence/<id>/review-architecture.md

## Risk & Rollback
- Risk: ...
- Rollback: revert PR; migration `NNNN.down.sql` is idempotent.

## Checklist
- [ ] Linked Issue
- [ ] Rebased on main
- [ ] Evidence pack complete
- [ ] No secrets in diff
- [ ] CI green
