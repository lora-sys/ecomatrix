"""Agent Contract tests."""
import pytest
from ecomatrix.contracts import (
    AgentContract,
    CONTRACTS,
    MINER_CONTRACT,
    MERCHANT_CONTRACT,
    HACKER_CONTRACT,
    MEDIATOR_CONTRACT,
    get_contract,
)


def test_all_four_contracts_registered():
    assert set(CONTRACTS) == {"miner", "merchant", "hacker", "mediator"}


def test_each_contract_has_all_required_fields():
    for c in CONTRACTS.values():
        assert c.name
        assert c.job_type
        assert c.goal
        assert c.inputs
        assert c.outputs
        assert c.capabilities
        assert c.limitations
        assert c.estimated_cost_tokens > 0


def test_self_trade_is_a_limitation_for_every_agent():
    for c in CONTRACTS.values():
        joined = "\n".join(c.limitations).lower()
        assert "yourself" in joined, f"{c.name}: missing self-trade limitation"


def test_mediator_has_balance_threshold():
    joined = "\n".join(MEDIATOR_CONTRACT.limitations).lower()
    assert "350" in joined


def test_hacker_caps_probes():
    joined = "\n".join(HACKER_CONTRACT.limitations).lower()
    assert "5 gold" in joined or "5g" in joined


def test_get_contract_known():
    c = get_contract("miner")
    assert c is MINER_CONTRACT


def test_get_contract_unknown_raises():
    with pytest.raises(KeyError):
        get_contract("wizard")


def test_system_prompt_section_mentions_goal_and_limitations():
    text = MINER_CONTRACT.system_prompt_section()
    assert "Goal:" in text
    assert "MUST NOT" in text
    assert "miner" in text
    assert "Cost budget" in text


def test_to_json_round_trip():
    j = MINER_CONTRACT.to_json()
    assert "miner" in j
    assert "EXECUTE_TRADE" in j
    parsed = MINER_CONTRACT.to_dict()
    assert parsed["name"] == "miner"
