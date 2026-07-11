"""Long-term memory tests."""
import json
from pathlib import Path

from ecomatrix.memory import LongTermMemory, PostgresLongTermMemory, ShortTermMemory


def test_short_term_caps_observations():
    m = ShortTermMemory()
    for i in range(80):
        m.observe(f"obs {i}")
    assert len(m.observations) == 50
    assert m.observations[0] == "obs 30"
    assert m.observations[-1] == "obs 79"


def test_long_term_roundtrip(tmp_path: Path):
    p = tmp_path / "ltm.json"
    ltm = LongTermMemory(path=p)
    ltm.update("agent_miner_01", summary="low on vitality", append_fact="bought food")
    ltm.update("agent_miner_01", append_fact="bought tools")

    data = json.loads(p.read_text())
    assert data["agent_miner_01"]["summary"] == "low on vitality"
    assert data["agent_miner_01"]["facts"] == ["bought food", "bought tools"]


def test_postgres_ltm_roundtrip_with_mock_transport():
    """Verify PostgresLongTermMemory uses PUT/GET correctly via httpx MockTransport."""
    import json
    from ecomatrix.a2a import A2AClient
    import httpx

    store = {"summary": "", "facts": []}

    def handler(request: httpx.Request) -> httpx.Response:
        body = json.loads(request.content) if request.content else {}
        if request.method == "GET":
            return httpx.Response(200, json={"long_term_memory": dict(store)})
        if request.method == "PUT":
            store.clear()
            store.update(body)
            return httpx.Response(200, json={"long_term_memory": dict(store)})
        return httpx.Response(405)

    transport = httpx.MockTransport(handler)
    fake = httpx.Client(base_url="http://example", transport=transport)
    # Build a minimal stand-in for A2AClient without going through __init__.
    class FakeA2A:
        _client = fake
    pg = PostgresLongTermMemory(FakeA2A())  # type: ignore[arg-type]
    pg.update("agent_test_01", summary="hello", append_fact="fact-1")
    pg.update("agent_test_01", append_fact="fact-2")
    out = pg.get("agent_test_01")
    assert out["summary"] == "hello"
    assert out["facts"] == ["fact-1", "fact-2"]
