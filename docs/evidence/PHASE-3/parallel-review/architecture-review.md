# Architecture Review (cold-start)

Date: 2026-07-11T08:17:36Z
Reviewer: architecture-reviewer (focused lens)

## Scope
- Monorepo layout
- Three apps: Go (backend), Next.js (frontend), Python (agent)
- Inter-app communication

## Findings

### 1. App boundaries

#### Backend
apps/backend/cmd:
seed
server

apps/backend/internal:
auth
config
domain
migrations
observability
repo
service
transport

apps/backend/migrations:
0001_init.down.sql
0001_init.up.sql
0002_agents_long_term_memory.down.sql
0002_agents_long_term_memory.up.sql

apps/backend/pkg:
a2a

Concerns:
- **tx_repo_list.go**: split across two files with no functional reason. Consolidate.
- **pkg/a2a**: shared between backend and the python world, but the python side reads JSON. Good API boundary, but verify no Go-specific types leak.

### 2. Service / Repo / Transport layering

Backend layers:
apps/backend/internal/repo:
agent_repo.go
db.go
feed_repo.go
ltm_test.go
migrations_fs
tx_repo.go
tx_repo_list.go

apps/backend/internal/service:
metrics.go
metrics_test.go
trade.go
trade_test.go

apps/backend/internal/transport:
http
ws

Concerns:
- **TradeService owns the tx boundary** (correct per ENGINEERING.md). Verify repo methods never open tx themselves.
- Service tests construct the *real* DB; this is the right call for lock tests.
- HTTP layer is a thin shell — good.

### 3. Frontend ↔ Backend boundary

BFF routes:
feeds
metrics

Concerns:
- **/api/proxy/{feeds,metrics}** are passthroughs. They exist only to avoid CORS preflights in dev.
- In production, the dashboard could call the backend directly with a tightened CORS allowlist; the BFF would become dead code.

### 4. Agent ↔ Backend boundary

Python A2A client + LangGraph graphs:
apps/agent/ecomatrix:
__init__.py
__pycache__
a2a.py
graphs
llm.py
memory.py
runner.py

apps/agent/ecomatrix/graphs:
__init__.py
__pycache__
base.py
hacker.py
mediator.py
merchant.py
miner.py

Concerns:
- **HTTP-only transport**. Reasonable for MVP but no gRPC/queue option.
- Stub LLM provider makes tests deterministic. Real LLM path is open but not exercised in tests.

### 5. Coupling to the harness operating system

Harness docs referenced from code:

Concerns: code references docs in comments but no automated link-check; stale doc references would rot.

### 6. Tech debt items already noted

From previous review reports (PHASE-3/review-report.md etc.):
docs/evidence/PHASE-1/review-report.md:- **Tech debt:** The `tx_repo_list.go` file is split from `tx_repo.go` for no real reason; consider merging in a follow-up. *(Tracked as Low.)*
docs/evidence/PHASE-1/review-report.md:- **Action:** None outstanding (Low noted above).

## Verdict
Architecture is clean. Three concerns to track:
1. tx_repo_list.go consolidation (trivial).
2. BFF proxy is dev-only; mark as such.
3. Doc references in code are nice-to-have but not enforced.
