# Phase 2 — Review Report (aggregated)

## bug-hunter (cold start)
- **State propagation bug, fixed:** Initial `act` node returned the entire state with `__result__` set on it, but `__result__` wasn't declared in `AgentState`. LangGraph silently dropped the unknown key, so `run_one_tick` fell back to a freshly-constructed `GraphResult` with `receipt=None, error=None` despite the trade succeeding. **Fix:** declared `last_receipt` / `last_error` in the TypedDict and made `act` return a delta dict. Caught by `tests/test_graphs.py::test_miner_graph_emits_trade`.
- **Self-trade in stub:** The stub defaulted to `target=agent_merchant_01`. When the merchant agent ran, it traded with itself (SELF_TRADE 422). **Fix:** stub detects `target == sender` and reroutes to a sibling. Caught by the scenario runner.
- **`extra={"msg": ...}` in logging:** clashed with LogRecord's built-in `msg`. Renamed to `errmsg` in `runner.py`.
- **Action:** None outstanding.

## architecture-reviewer (cold start)
- **Coupling:** `a2a.py` is the only module that touches HTTP. `runner.py` and `graphs/*` depend on it via a Protocol-shaped `A2AClient`; tests pass a `FakeClient` instead of mocking HTTP. Clean.
- **TypedDict state shape:** `AgentState` is now a closed surface. Adding a new field is a one-line change. Reasonable.
- **Conservation check:** Initially checked only two-agent net. Each agent also trades with siblings, so two-agent net is *not* conserved. **Fixed** to check world-total GOLD.
- **Action:** None outstanding.

## security-reviewer (cold start)
- **Free-text reasoning:** not logged in runner (only `errmsg` and `code`). LLM prompts pass user-influenced text; we never feed that text into a tool argument, so prompt injection is bounded.
- **OpenAI key handling:** `OpenAICompatibleLLM` raises if `api_key` is missing rather than silently sending an empty bearer. Good.
- **Action:** None outstanding.

## Aggregator Verdict

**No Critical/High findings. Phase 2 ships.** World-GOLD conservation holds across both 5-tick and 10-tick scenarios with 0 rejected trades and 0 errors.

Follow-up Issues to file in Phase 2.5 / Phase 3:
1. Add migration `0002_agents_long_term_memory.sql` and wire `PostgresLongTermMemory`.
2. Implement `--scenario multi` runner that fans out to all seeded agents.
3. Replace the stub LLM path with the OpenAI provider in CI; keep the stub for offline tests.
