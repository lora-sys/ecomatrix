# EcoMatrix Documentation Index

> Agents: **read this first** before opening any other doc. Pull the L1 manifest pinned to your Issue.

## Layer Map (L0 → L3)

### L0 — Always-on (in `AGENTS.md`)
- Project identity, tech stack, forbidden actions, A2A envelope, evidence gate.

### L1 — Per Issue
Pull from your Issue's "Related Docs" section. Always includes:
- `docs/product/prd.md` (PRD)
- `docs/product/roadmap.md` (current phase scope)
- the relevant `docs/architecture/*.md`
- the relevant `DESIGN.md` section if UI

### L2 — On Demand
- `docs/architecture/system.md` for cross-layer questions
- `docs/decisions/` ADRs (filter by tag relevant to your Issue)
- `memory/lessons.md` for prior pain

### L3 — Deep / Rarely
- `docs/evidence/<id>/` (only when investigating a specific PR)
- `sessions/` (only when debugging a multi-agent run)

## Document Catalog

### Product
| Doc | Purpose |
| --- | ------- |
| [PRD](product/prd.md) | What EcoMatrix is, who it's for, milestones. |
| [Roadmap](product/roadmap.md) | Phase 1–3 deliverables and exit criteria. |

### Architecture
| Doc | Purpose |
| --- | ------- |
| [System](architecture/system.md) | C4-style overview + trade sequence. |
| [Backend](architecture/backend.md) | Go service rules, env vars, invariants. |
| [Frontend](architecture/frontend.md) | Next.js routes, data flow, i18n. |
| [Database](architecture/db.md) | Schema, locking, migrations, seed. |
| [Agent](architecture/agent.md) | LangGraph agent design. |
| [API / A2A](architecture/api.md) | Protocol v1.1 authoritative spec. |
| [Security](architecture/security.md) | Threat model, logging hygiene. |

### Design
| Doc | Purpose |
| --- | ------- |
| [DESIGN.md](../DESIGN.md) | Brand, color tokens, components, motion. |

### Engineering
| Doc | Purpose |
| --- | ------- |
| [ENGINEERING.md](../ENGINEERING.md) | Cross-cutting rules. |
| [TESTING.md](../TESTING.md) | Test strategy + evidence gate. |
| [CONTRIBUTING.md](../CONTRIBUTING.md) | Issue-first workflow. |

### Decisions
| Doc | Purpose |
| --- | ------- |
| [decisions/](decisions/) | One ADR per non-obvious decision. |

### Evidence
| Doc | Purpose |
| --- | ------- |
| [evidence/](evidence/) | One folder per merged Issue; PR evidence pack. |

### Sessions
| Doc | Purpose |
| --- | ------- |
| [sessions/](../sessions/) | Multi-agent run logs. |
