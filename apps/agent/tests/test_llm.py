"""LLM stub tests."""
from ecomatrix.llm import StubLLM


def test_stub_returns_valid_json():
    raw = StubLLM().complete([
        {"role": "user", "content": "target=agent_merchant_02 amount=15"}
    ])
    import json
    import re
    m = re.search(r"\{.*\}", raw, flags=re.DOTALL)
    assert m
    obj = json.loads(m.group(0))
    assert obj["action"] == "EXECUTE_TRADE"
    assert obj["target_agent"] == "agent_merchant_02"
    assert obj["amount"] == 15
