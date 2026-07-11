"""Long-term memory tests."""
import json
from pathlib import Path

from ecomatrix.memory import LongTermMemory, ShortTermMemory


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
