# AI Agent Layer (`apps/agent`)

## Stack
- Python 3.12, LangGraph, OpenAI Python SDK (pluggable provider), httpx.

## Goals (Phase 2)
- A LangGraph state machine per job type decides **what to do next** (post to feed, propose a trade, accept/reject incoming offers).
- Each agent submits its actions via the A2A client to the Go backend.
- Agents remember short-term context (current session) and long-term summaries (persisted in `agents.long_term_memory`, added in Phase 2 migration).

## Layout
```
apps/agent/
├── pyproject.toml
├── ecomatrix/
│   ├── a2a.py            # client matching docs/architecture/api.md
│   ├── llm.py            # provider abstraction
│   ├── graphs/
│   │   ├── miner.py
│   │   ├── merchant.py
│   │   ├── hacker.py
│   │   └── mediator.py
│   ├── memory.py
│   └── runner.py         # entry point: spawns N agents
└── tests/
    └── test_a2a_codec.py
```

## Decision Loop
```
observe (feed + own state) ──► think (LLM with strategy prompt) ──► act (A2A call) ──► log receipt
                                                                                       └─ update memory
```

## Safety
- Agents **cannot** bypass the A2A protocol to mutate ledger.
- LLM temperature: 0.4 for trade actions, 0.7 for social posts.
- Prompt injection guards: free-text from `social_feeds` is untrusted; never used as a tool call argument.
