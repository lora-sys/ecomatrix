"""ISS-030 additions: A2AClient supervisor-detail / per-agent helpers.

These helpers mirror the new backend routes:
  GET /v1/supervisor/runs/{id}
  GET /v1/agents/by-string-id/{sid}/supervisor-runs?limit=N
"""

from __future__ import annotations

import json
from typing import Any

import pytest

import httpx

from ecomatrix.a2a import A2AClient, A2AError


class _StubResponse:
    def __init__(self, status_code: int, payload: Any = None, text: str = "") -> None:
        self.status_code = status_code
        self._payload = payload
        self.text = text

    def json(self) -> Any:
        if self._payload is not None:
            return self._payload
        return json.loads(self.text or "{}")


class _StubClient:
    def __init__(self, responses: list[_StubResponse]) -> None:
        self._responses = list(responses)
        self.calls: list[tuple[str, dict[str, Any] | None]] = []

    def get(self, path: str, params: dict[str, Any] | None = None) -> _StubResponse:
        self.calls.append((path, params))
        if not self._responses:
            raise AssertionError(f"unexpected GET {path}")
        return self._responses.pop(0)

    def post(self, path: str, json: Any = None) -> _StubResponse:
        self.calls.append((path, json))
        if not self._responses:
            raise AssertionError(f"unexpected POST {path}")
        return self._responses.pop(0)

    def close(self) -> None:  # pragma: no cover - parity with httpx.Client
        pass


@pytest.fixture
def client(monkeypatch: pytest.MonkeyPatch) -> tuple[A2AClient, _StubClient]:
    stub = _StubClient([])
    monkeypatch.setattr("ecomatrix.a2a.httpx.Client", lambda *a, **kw: stub)
    c = A2AClient("http://localhost:8080")
    return c, stub


def test_fetch_supervisor_run_returns_payload(client: tuple[A2AClient, _StubClient]) -> None:
    a2a, stub = client
    stub._responses.append(_StubResponse(200, payload={
        "id": 17,
        "goal": "trade",
        "status": "finished",
        "worker_results": [{"agent_id": "agent_miner_01"}],
    }))
    out = a2a.fetch_supervisor_run(17)
    assert out["id"] == 17
    assert stub.calls[0][0] == "/v1/supervisor/runs/17"


def test_fetch_supervisor_run_raises_on_404(client: tuple[A2AClient, _StubClient]) -> None:
    a2a, stub = client
    stub._responses.append(_StubResponse(404, text="not found"))
    with pytest.raises(A2AError) as exc:
        a2a.fetch_supervisor_run(99)
    assert exc.value.http_status == 404


def test_list_agent_supervisor_runs_returns_list(client: tuple[A2AClient, _StubClient]) -> None:
    a2a, stub = client
    stub._responses.append(_StubResponse(200, payload={
        "runs": [
            {"id": 1, "goal": "x", "worker_results": [{"agent_id": "agent_miner_01"}]},
            {"id": 2, "goal": "y", "worker_results": [{"agent_id": "agent_miner_01"}]},
        ]
    }))
    out = a2a.list_agent_supervisor_runs("agent_miner_01", limit=5)
    assert len(out) == 2
    assert stub.calls[0][0] == "/v1/agents/by-string-id/agent_miner_01/supervisor-runs"
    assert stub.calls[0][1] == {"limit": 5}


def test_list_agent_supervisor_runs_returns_empty_on_missing_runs_key(
    client: tuple[A2AClient, _StubClient],
) -> None:
    a2a, stub = client
    stub._responses.append(_StubResponse(200, payload={}))
    out = a2a.list_agent_supervisor_runs("agent_miner_01")
    assert out == []
