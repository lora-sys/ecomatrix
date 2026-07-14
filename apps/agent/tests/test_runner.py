"""Runner dispatch tests for the supervisor scenario."""

import logging

from ecomatrix import runner
from ecomatrix.supervisor import SupervisorResult
from tests.test_react import FakeA2A


class SeededFakeA2A(FakeA2A):
    def list_agents(self, limit=200):
        return [
            {"StringID": "agent_miner_01", "JobType": "miner"},
            {"StringID": "agent_merchant_01", "JobType": "merchant"},
            {"StringID": "ignored", "JobType": "unknown"},
        ]


def test_run_supervisor_scenario_uses_seeded_supported_agents(monkeypatch, capsys):
    monkeypatch.setenv("ECOMATRIX_AGENT_TRACES", "0")

    result = runner.run_supervisor_scenario(
        SeededFakeA2A(),
        goal="trade 10 GOLD",
        max_subtasks=1,
        log=logging.getLogger("test"),
    )

    assert result.error is None
    assert result.worker_results[0]["agent_id"] in {
        "agent_miner_01",
        "agent_merchant_01",
    }
    assert '"goal": "trade 10 GOLD"' in capsys.readouterr().out


def test_build_worker_registry_includes_contract_and_is_stable():
    registry = runner.build_worker_registry(SeededFakeA2A())

    assert [worker.agent_id for worker in registry] == [
        "agent_merchant_01",
        "agent_miner_01",
    ]
    assert all(worker.capabilities for worker in registry)
    assert all(worker.limitations for worker in registry)


def test_run_supervisor_scenario_reports_no_supported_workers(monkeypatch):
    client = SeededFakeA2A()
    client.list_agents = lambda limit=200: [{"StringID": "x", "JobType": "unknown"}]
    monkeypatch.setenv("ECOMATRIX_AGENT_TRACES", "0")

    result = runner.run_supervisor_scenario(
        client,
        goal="trade",
        max_subtasks=1,
        log=logging.getLogger("test"),
    )

    assert result.error == "at least one worker is required"


def test_main_dispatches_supervisor_and_returns_success(monkeypatch):
    class ClientContext:
        def __init__(self, backend):
            self.backend = backend

        def __enter__(self):
            return self

        def __exit__(self, *exc):
            return None

    captured = {}

    def fake_run(client, *, goal, max_subtasks, log):
        captured.update(goal=goal, max_subtasks=max_subtasks)
        return SupervisorResult(goal, [], [], "done", 1)

    monkeypatch.setattr(runner, "A2AClient", ClientContext)
    monkeypatch.setattr(runner, "run_supervisor_scenario", fake_run)

    exit_code = runner.main(
        [
            "--scenario",
            "supervisor",
            "--goal",
            "stabilize economy",
            "--max-subtasks",
            "2",
        ]
    )

    assert exit_code == 0
    assert captured == {"goal": "stabilize economy", "max_subtasks": 2}


def test_main_supervisor_partial_failure_returns_error(monkeypatch):
    class ClientContext:
        def __init__(self, backend):
            pass

        def __enter__(self):
            return self

        def __exit__(self, *exc):
            return None

    worker_results = [
        {"error": None},
        {"error": "worker failed"},
    ]
    monkeypatch.setattr(runner, "A2AClient", ClientContext)
    monkeypatch.setattr(
        runner,
        "run_supervisor_scenario",
        lambda *args, **kwargs: SupervisorResult(
            "goal", [], worker_results, "partial", 1
        ),
    )

    assert runner.main(["--scenario", "supervisor", "--goal", "goal"]) == 1


def test_main_keeps_two_agent_dispatch(monkeypatch):
    class ClientContext:
        def __init__(self, backend):
            self.backend = backend

        def __enter__(self):
            return self

        def __exit__(self, *exc):
            return None

    monkeypatch.setattr(runner, "A2AClient", ClientContext)
    monkeypatch.setattr(
        runner,
        "run_two_agent",
        lambda client, **kwargs: {"conservation": True, "errors": []},
    )

    assert runner.main(["--scenario", "two_agent", "--ticks", "1"]) == 0


def test_main_keeps_multi_dispatch(monkeypatch):
    class ClientContext:
        def __init__(self, backend):
            pass

        def __enter__(self):
            return self

        def __exit__(self, *exc):
            return None

    monkeypatch.setattr(runner, "A2AClient", ClientContext)
    monkeypatch.setattr(
        runner,
        "run_multi_agent",
        lambda client, **kwargs: {"conservation": True, "errors": []},
    )

    assert runner.main(["--scenario", "multi", "--ticks", "1"]) == 0


def test_parse_error_handles_string_and_object_forms():
    """GraphResult.error can be a serialized string or an A2AError-shaped
    object. _parse_error must return (code, message) in both cases so the
    runner can recognize INSUFFICIENT_FUNDS specifically."""
    from ecomatrix.runner import _parse_error

    none_code, none_msg = _parse_error(None)
    assert none_code is None and none_msg is None

    code, msg = _parse_error("trade: INSUFFICIENT_FUNDS: not enough gold")
    assert code == "INSUFFICIENT_FUNDS"
    assert "INSUFFICIENT_FUNDS" in msg

    code, msg = _parse_error("parse: bad json")
    assert code == "parse"
    assert "bad json" in msg

    class Obj:
        code = type("C", (), {"value": "FORBIDDEN"})()
        message = "denied"

    code, msg = _parse_error(Obj())
    assert code == "FORBIDDEN"
    assert msg == "denied"
