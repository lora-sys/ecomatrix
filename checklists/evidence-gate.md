# Evidence Gate Checklist

The Coordinator must verify **every** item before promoting an Issue from Review → Done.

## Universal
- [ ] Linked PR exists and is squash-merged.
- [ ] CI green on the merge commit.
- [ ] `docs/evidence/<id>/change-summary.md` present.
- [ ] `docs/evidence/<id>/verification.md` present with command output.
- [ ] `docs/evidence/<id>/review-report.md` aggregates ≥ 2 reviewer reports.
- [ ] No Critical/High findings open. (Low/Nit deferred with linked Issues.)

## Backend
- [ ] `go test -race ./...` output captured in `test-results/`.
- [ ] Race-detector clean.
- [ ] At least one concurrency test for any state-mutating endpoint.

## Database
- [ ] Migration `up` succeeded on a fresh DB.
- [ ] Migration `down` succeeded.
- [ ] `seed` script re-runnable.

## Frontend
- [ ] Playwright suite green.
- [ ] Desktop (1440×900) and mobile (390×844) screenshots.
- [ ] Browser console clean in screenshots.
- [ ] UI reviewer report present.

## Cross-Cutting
- [ ] No secrets in diff.
- [ ] No `TODO` without a linked Issue.
- [ ] `docs/INDEX.md` still accurate (or updated).
