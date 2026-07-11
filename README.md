# EcoMatrix

A fully autonomous multi-agent sandbox. AI agents with distinct jobs, balances, and survival goals publish needs, negotiate, and execute economic trades over a custom **A2A (Agent-to-Agent) protocol**, observed through a "God's-Eye" dashboard.

## Status

| Phase | Scope                                | State |
| ----- | ------------------------------------ | ----- |
| 1     | Physical Engine — Go backend + Postgres | ✅ Shipped (`docs/evidence/PHASE-1/`) |
| 2     | Brain — Python LangGraph agent         | ✅ Shipped (`docs/evidence/PHASE-2/`) |
| 3     | God's Eye — Next.js dashboard         | ✅ Shipped (`docs/evidence/PHASE-3/`) |

## Repository Layout

```
apps/
  backend/   Go service (HTTP + WebSocket)        — Phase 1
  frontend/  Next.js dashboard (Aceternity UI)    — Phase 3
  agent/     Python LangGraph agent runner        — Phase 2
docs/        Source of truth (PRD, architecture, design, decisions, evidence)
memory/      Curated, persistent lessons & facts
sessions/    Per-run multi-agent logs
tasks/       Phase 1–3 issue drafts
templates/   Issue / PR / Evidence / ADR templates
checklists/  Evidence gate, PR-merge
scripts/     bootstrap / refresh-index / new-session / evidence-pack
```

## Quickstart

```bash
# Postgres (re-uses repotwin-postgres-1 if present, else this compose)
docker compose up -d db

# Build + seed + run
cd apps/backend
make build
ECOMATRIX_DB_DSN='postgres://repotwin:repotwin@localhost:5432/ecomatrix?sslmode=disable' ./bin/seed
ECOMATRIX_DB_DSN='postgres://repotwin:repotwin@localhost:5432/ecomatrix?sslmode=disable' ./bin/server

# Smoke
curl http://localhost:8080/healthz
curl http://localhost:8080/v1/agents
make test-race
```

## Operating System for AI Agents

This repo is bootstrapped with the `ai-engineering-harness` workflow. Read these in order:

1. [CLAUDE.md](CLAUDE.md) / [AGENTS.md](AGENTS.md) — global rules, forbidden actions, file allow-list.
2. [PROJECT_STATUS.md](PROJECT_STATUS.md) — live board of what's next.
3. [docs/INDEX.md](docs/INDEX.md) — doc catalog + L0–L3 context rules.
4. [CONTRIBUTING.md](CONTRIBUTING.md) — Issue-first workflow.

## Documentation

- PRD: [docs/product/prd.md](docs/product/prd.md)
- Roadmap: [docs/product/roadmap.md](docs/product/roadmap.md)
- System architecture: [docs/architecture/system.md](docs/architecture/system.md)
- A2A protocol v1.1: [docs/architecture/api.md](docs/architecture/api.md)
- Design system: [DESIGN.md](DESIGN.md)
- Engineering rules: [ENGINEERING.md](ENGINEERING.md)
- Testing & evidence: [TESTING.md](TESTING.md)

## License

TBD.
