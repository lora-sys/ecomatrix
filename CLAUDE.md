# CLAUDE.md — EcoMatrix Source of Truth

> This file is the **operating system** for AI agents and humans working on EcoMatrix.
> It is mirrored to `AGENTS.md` for tooling compatibility. Edit either; sync the other.

## 1. What is EcoMatrix

A fully autonomous multi-agent sandbox in which AI agents with distinct jobs, balances, and survival goals publish needs, negotiate, and execute economic trades over a custom **A2A (Agent-to-Agent) protocol**. A "God's Eye" dashboard observes the economy as it evolves.

- Full PRD: `docs/product/prd.md`
- Roadmap: `docs/product/roadmap.md`
- Live status: `PROJECT_STATUS.md`

## 2. Tech Stack (locked)

| Layer    | Stack                                                                            |
| -------- | -------------------------------------------------------------------------------- |
| Frontend | Next.js (App Router) + React + Aceternity UI + Tailwind CSS + Framer Motion      |
| Backend  | Go (1.26+) + Fiber + Gorilla WebSocket + GORM                                   |
| Database | PostgreSQL 16 (row-level locks for trade atomicity)                              |
| AI       | Python 3.12 + LangGraph + an LLM provider (pluggable)                            |
| Infra    | Docker (dev), pnpm workspaces, GitHub Actions                                    |

Do not introduce a new framework, ORM, or DB without an ADR in `docs/decisions/`.

## 3. Repository Layout

```
.
├── apps/
│   ├── backend/         # Go service (HTTP + WS)
│   ├── frontend/        # Next.js dashboard
│   └── agent/           # Python LangGraph agent runner
├── docs/                # Source of truth (PRD, architecture, design, decisions, evidence)
├── memory/              # Curated, persistent lessons & facts
├── sessions/            # Per-run multi-agent logs
├── tasks/               # Coordinator task board mirror
├── templates/           # Issue / PR / Evidence / Phase templates
├── checklists/          # Evidence gate & reviewer checklists
├── skills/              # Project-local skills discovered during work
├── scripts/             # Bash helpers
└── .github/             # Issue & PR templates, CI workflows
```

Monorepo rule: every app has its own `package.json` / `go.mod` / `pyproject.toml`. No cross-app imports without an ADR.

## 4. Operating Principles (Harness Loop)

1. **Issues are the unit of work.** No change without an Issue in `.github/ISSUE_TEMPLATE/` form.
2. **Worktrees isolate agents.** Branch per Issue: `feature/ISSUE-<id>-<slug>`. Never edit `main` directly.
3. **PRs carry evidence.** No merge without green CI + at least 2 reviewer reports + Evidence pack.
4. **Adversarial review.** Bug Hunter, Behavior Reviewer, Architecture Reviewer (Security/UI on demand).
5. **L0–L3 context discipline.** Don't bulk-load `docs/`; load per Issue via the workflow.
6. **Memory is project state.** Stable conclusions in `docs/` + `memory/`; chat is ephemeral.

## 5. Forbidden Actions (L0 — never violate)

- ❌ Editing files outside your assigned scope (see Issue owner).
- ❌ Merging your own PR.
- ❌ Skipping the Evidence gate "because tests pass locally".
- ❌ Hand-rolling an auth, crypto, or financial primitive — use the shared one.
- ❌ Storing secrets in code; use `.env` + the documented env contract.
- ❌ Schema changes without a migration + rollback plan.
- ❌ Force-push to `main`/`master`.
- ❌ Loading the entire `docs/` tree into a single context window.

## 6. Agent Roster (use via the harness)

`coordinator`, `explore`, `plan`, `frontend`, `backend`, `database`, `qa`, `bug-hunter`, `behavior-reviewer`, `architecture-reviewer`, `security-reviewer`, `ui-reviewer`, `conflict-resolver`, `release`, `review-aggregator`, `context-assembly`, `memory-curator`.

Per-role scope, inputs, outputs, and acceptance criteria are pinned in `tasks/` per Issue.

## 7. A2A Protocol (v1.1)

```json
{
  "protocol_v": "1.1",
  "msg_id": "tx_req_9948",
  "sender": "agent_miner_01",
  "action": "EXECUTE_TRADE",
  "payload": { "...": "see docs/architecture/api.md" },
  "timestamp": 1713532588
}
```

Authoritative spec: `docs/architecture/api.md`. Backend must validate `protocol_v == "1.1"` and reject anything else with HTTP 400.

## 8. Evidence Gate (minimum)

| Type        | Evidence                                                            |
| ----------- | ------------------------------------------------------------------- |
| Backend     | `go test ./...` green + curl trace + race-detector pass + reviewer reports |
| Database    | Migration SQL, up/down dry run, data-safety diff                    |
| Frontend    | Playwright desktop+mobile screenshots, console clean, UI Reviewer  |
| Cross-cut   | `docs/evidence/<issue>/change-summary.md` + `verification.md`       |

## 9. Environment & Secrets

- Local dev DB: `postgres://repotwin:repotwin@localhost:5432/ecomatrix` (dev only).
- Each app ships `.env.example`; real secrets via `.env` (git-ignored).
- Required env vars are listed in `docs/architecture/backend.md` and `docs/architecture/frontend.md`.

## 10. First-time bootstrap

If `docs/INDEX.md` is missing or stale, run `bash scripts/bootstrap.sh` (or follow the harness `00-project-bootstrap` workflow). Do not improvise.

## 11. Quick Reference

- "What are we building?" → `docs/product/prd.md`
- "What's next?" → `PROJECT_STATUS.md`
- "How do I code?" → `ENGINEERING.md` + your app's `README.md`
- "How do I test?" → `TESTING.md`
- "How does it look?" → `DESIGN.md`
- "How do I propose a change?" → `CONTRIBUTING.md`
