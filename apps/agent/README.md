# EcoMatrix Agent Runner

Python service that drives LangGraph agents, which call the Go backend via the A2A v1.1 protocol.

## Quickstart (dev)

```bash
# Install
uv sync --extra dev

# Or:
pip install -e '.[dev]'

# Run the two-agent scenario
ECOMATRIX_AGENT_BACKEND_URL=http://localhost:8080 \
  python -m ecomatrix.runner --scenario two_agent --ticks 5
```

## Layout

```
ecomatrix/
  a2a.py       # client matching pkg/a2a (Go)
  llm.py       # provider abstraction; stub + openai-compatible
  memory.py    # short-term (graph state) + long-term (postgres JSONB)
  graphs/
    miner.py   # LangGraph state machine per job type
    merchant.py
    hacker.py
    mediator.py
  runner.py    # entry point
tests/
  test_a2a_codec.py
  test_graphs.py
  test_scenario.py
```

## A2A parity

The Python codec is the mirror of `apps/backend/pkg/a2a`. `tests/test_a2a_codec.py`
verifies parity by sharing the JSON fixtures from the Go side (when present).

## Scenario

The Phase 2 exit scenario spawns `agent_miner_01` and `agent_merchant_01`,
runs N ticks, and asserts the ledger is consistent with the strategy log.
