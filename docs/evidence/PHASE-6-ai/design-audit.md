# Phase 6 — Design Audit (Pre-Implementation)

Date: 2026-07-12
Spec source: user-provided design document for "Production Agent = LLM + Tools + State + Memory + Workflow + Evaluation + Observability"

## Coverage matrix

| Spec section | Required | Current state | Gap |
| ----------- | -------- | ------------- | --- |
| Agent ≠ Chatbot (goal-driven) | Goal in state | `goal` not in AgentState | ❌ Add goal to state |
| Single-goal principle | One job per agent | 4 agent types, each with own job ✓ | OK |
| Agent Contract (Input/Output/Capability/Limitation/Example/Cost) | Formal contract | None — implicit in strategy_prompt | ❌ Add AgentContract dataclass |
| Architecture: Controller + Planner + Memory + Executor | All present | Have observe→think→act but no Planner separate from Executor | ⚠ Single graph |
| Reasoning structured (Plan/Action/Observation/NextStep) | Explicit | Think produces `decision` but no `plan` field | ❌ Add plan/current_step/next_step |
| Tool design (small, clear, strict schema) | Quality over quantity | 3 tools, well-defined | ✓ |
| MCP server | Standard protocol | None (direct A2A HTTP) | Deferred (Phase 7) |
| Short-term memory (current task state) | Bounded | `ShortTermMemory` with cap of 50 | ✓ |
| Long-term memory (user preference, skills, knowledge) | Persistent | `PostgresLongTermMemory` + JSONB | ✓ |
| Memory compression | Avoid pollution | No compression | ❌ Add summarize_short_term |
| Workflow types (Simple/ReAct/Planner/Multi-Agent) | Choose by task | Only one workflow (think→act) | ❌ Add ReAct + Planner |
| Prompt engineering (Identity/Goal/Constraints/Tools/Output) | Standard | Strategy prompt covers Goal/Tools only | ⚠ Formalize |
| Security (least privilege, human-in-the-loop) | Required for high-risk | Admin token only | ❌ Add approval for high-value trades |
| Error handling (retry → alternative tool → ask user) | Required | Retry only | ❌ Add alternative tool + ask user |
| Observability (Trace: User → Decision → Tool → Result) | Required | Conversation log has decisions + tools but not cost | ❌ Add traces + cost tracking |
| Cost control | Required | None | ❌ Add token budget |
| Evaluation framework | Required | None | ❌ Add eval.py with golden cases |
| Failure knowledge base | Required | Errors logged but not learned from | ❌ Add failure cache |
| Common failures (over-complex / no tools / too many tools / memory pollution) | Awareness | Single-purpose agents, 3 tools, bounded memory | ✓ |
| Multi-agent supervision | Avoid free chat | Each agent is a Supervisor in multi scenario | ✓ |

## Priority for Phase 6

1. **Agent Contract** (foundational — everything else references it)
2. **Structured State** (plan/action/observation/next_step) — required for ReAct
3. **ReAct workflow** (the spec's main recommendation for exploration)
4. **Memory compression** (avoid memory pollution)
5. **Cost control** (token budget per agent per tick)
6. **Human-in-the-loop** for high-risk trades
7. **Traces + Observability** (decision / tool / error / cost)
8. **Evaluation framework** with golden test cases

## Deferred to Phase 7+

- MCP server (would need an MCP SDK integration)
- Vector-based semantic memory
- Multi-agent supervisor pattern (we have a flat multi-agent, not a hierarchical one)
- LLM-as-judge eval
- Failure knowledge base (auto-learning from errors)
