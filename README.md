# EcoMatrix

A fully autonomous multi-agent sandbox. AI agents with distinct jobs, balances, and survival goals publish needs, negotiate, and execute economic trades over a custom **A2A (Agent-to-Agent) protocol**, observed through a "God's-Eye" dashboard.

## Status

| Phase | Scope                                | State |
| ----- | ------------------------------------ | ----- |
| 1     | Physical Engine — Go backend + Postgres | ✅ Shipped (`docs/evidence/PHASE-1/`) |
| 2     | Brain — Python LangGraph agent         | ✅ Shipped (`docs/evidence/PHASE-2/`) |
| 3     | God's Eye — Next.js dashboard         | ✅ Shipped (`docs/evidence/PHASE-3/`) |

## Dashboard Demo (ISS-FRONTEND)

Latest creative pass on the dashboard. Hit with the `$frontend-creative` skill:
extended type, ambient drift + grain + scanline, hairline gradient frames,
live ticker ribbon, transmissal hero on every detail page.

|  | Before | After |
| --- | --- | --- |
| Dashboard (desktop) | ![Dashboard before](docs/design/ISSUE-FRONTEND/before/dashboard-desktop.png) | ![Dashboard after](docs/design/ISSUE-FRONTEND/after/dashboard-desktop.png) |
| Dashboard (mobile)  | ![Dashboard before](docs/design/ISSUE-FRONTEND/before/dashboard-mobile.png)  | ![Dashboard after](docs/design/ISSUE-FRONTEND/after/dashboard-mobile.png)  |
| Agent dossier (desktop) | ![Agent before](docs/design/ISSUE-FRONTEND/before/agents-desktop.png) | ![Agent after](docs/design/ISSUE-FRONTEND/after/agents-desktop.png) |
| Agent dossier (mobile)  | ![Agent before](docs/design/ISSUE-FRONTEND/before/agents-mobile.png)  | ![Agent after](docs/design/ISSUE-FRONTEND/after/agents-mobile.png)  |
| Supervisor run (desktop) | ![Supervisor before](docs/design/ISSUE-FRONTEND/before/supervisor-desktop.png) | ![Supervisor after](docs/design/ISSUE-FRONTEND/after/supervisor-desktop.png) |
| Supervisor run (mobile)  | ![Supervisor before](docs/design/ISSUE-FRONTEND/before/supervisor-mobile.png)  | ![Supervisor after](docs/design/ISSUE-FRONTEND/after/supervisor-mobile.png)  |

What changed (cf. `docs/design/ISSUE-FRONTEND/brief.md`):

- **Ambient layer** (`components/ambient-bg.tsx`). Three drifting radial blobs, scanline rasterizer, SVG film grain — placed once in the root layout, fixed to the viewport.
- **Masthead** (`components/masthead.tsx`). Layered kicker + gradient display headline (`gradient-text-cyan-violet` / `gradient-text-rainbow`) + secondary subhead + glyph. Drives every page hero.
- **Live ticker ribbon** (`components/ticker-ribbon.tsx`). Marquee of live agents, GOLD reserve, QPS, WS, jobs breakdown, last-tx, supervisor runs. Streams animations pause on hover.
- **BigMetric** (`components/big-metric.tsx`). KPI tile replacement with hairline gradient frames, animated bottom strip, big tabular-numeral value, dual-tone delta hint.
- **Dossier layout** on `/agents/[id]`. Three BigMetric tiles up top (BALANCE / VITALITY / CREDIT) so the dossier feels like a personnel file, then long-form panels below.
- **Transmissal layout** on `/supervisor/[id]`. Run header with run-id gradient + lightning glyph, four telemetry tiles, then the full run trace underneath.

Stack stays within Next.js + Tailwind + framer-motion (no GSAP, no R3F). All gates green locally: `npm run typecheck`, `npm run lint`, `npm run build`, full Playwright e2e (18 passed / 2 skipped across desktop + mobile).

## Project Completion

This project is feature-complete against the PRD. See [PROJECT_COMPLETION.md](PROJECT_COMPLETION.md) for the closing record.

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

## Quickstart — one command

```bash
# Bring up Postgres, build, seed, start backend (8080), frontend (3100),
# and the multi-agent scenario, all in one shot. Ctrl-C stops everything.
make demo
```

Then open http://localhost:3100.

## Quickstart — manual

```bash
# 1. Postgres
docker compose up -d db

# 2. Backend (terminal 1)
cd apps/backend && make build
ECOMATRIX_DB_DSN='postgres://repotwin:repotwin@localhost:5432/ecomatrix?sslmode=disable' ./bin/seed
ECOMATRIX_DB_DSN='postgres://repotwin:repotwin@localhost:5432/ecomatrix?sslmode=disable' ./bin/server

# 3. Dashboard (terminal 2)
cd apps/frontend && npm install && PORT=3100 npx next dev -p 3100

# 4. Agents (terminal 3)
cd apps/agent && uv venv --python 3.12 .venv && . .venv/bin/activate && uv pip install -e '.[dev]'
. .venv/bin/activate && python -m ecomatrix.runner --scenario multi --ticks 5 --tick-seconds 0.3
```

## Tests

```bash
make test    # backend -race + agent pytest + frontend tsc + playwright
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
