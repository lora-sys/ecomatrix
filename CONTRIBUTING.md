# CONTRIBUTING.md — How to Propose a Change

> AI agents and humans follow the same flow. **Issues are the unit of work.**

## 1. The Flow

```
Idea → Issue → branch + worktree → Implementation Plan → code → tests → PR → review → evidence → merge
```

No PR is opened without an Issue. No Issue is implemented without an Implementation Plan linked from the Issue body.

## 2. Pick an Issue

1. Read `PROJECT_STATUS.md` for the next `Todo` item.
2. Claim it: assign yourself (or your agent) and move it to `Implementing` in `PROJECT_STATUS.md`.
3. Create branch: `feature/ISSUE-<id>-<slug>`.
4. Create worktree if running concurrent work.

## 3. Write the Code

- Stay inside the file allow-list pinned in the Issue owner section.
- Follow `ENGINEERING.md` and `DESIGN.md`.
- Add or update tests for any non-trivial change.
- Don't drive-by refactor unrelated files.

## 4. Self-Review Before Requesting Review

- [ ] I rebased on `main`.
- [ ] `go test -race ./...` (or the relevant suite) is green locally.
- [ ] I produced `docs/evidence/<id>/{change-summary,verification}.md`.
- [ ] I added screenshots / curl traces for UI or API changes.
- [ ] No secrets in the diff.
- [ ] No unrelated file edits.

## 5. Open the PR

Use `.github/PULL_REQUEST_TEMPLATE.md`. The PR must include:

- Linked Issue (`Closes #N`).
- Scope summary (≤ 5 bullets).
- Evidence pack link.
- Risk and rollback notes.

## 6. Review

Two reviewers minimum. Reviewer roles per change type are in `ENGINEERING.md §8`. Address every Critical/High before merge; Low/Nit can be deferred with an Issue.

## 7. Merge

- Squash-merge.
- PR body becomes the merge commit message body.
- Close the linked Issue. Update `PROJECT_STATUS.md` to `Done` for that row.

## 8. After Merge

- Memory: append durable lessons to `memory/lessons.md`.
- Architecture: if you made a non-obvious decision, write an ADR in `docs/decisions/`.
- Index: run `bash scripts/refresh-index.sh` if you changed `docs/`.

## 9. For AI Agents Specifically

- You may not skip the Issue step.
- You may not self-approve.
- You may not bulk-load `docs/` — load the L1 manifest your coordinator gave you.
- You may not write to `main`/`master` directly. Period.

## 10. Issue Templates

- `templates/issue-feature.md`
- `templates/issue-bug.md`
- `templates/issue-refactor.md`
- `templates/issue-spike.md`

(Also available as `.github/ISSUE_TEMPLATE/*.md` for the GitHub UI.)
