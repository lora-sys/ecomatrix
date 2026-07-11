"""Agent memory.

Two layers:

- **Short-term**: lives inside the LangGraph state dict per tick.
- **Long-term**: optional Postgres JSONB column ``agents.long_term_memory``.
  We expose a no-op file-backed store by default so the agent can run offline;
  the Postgres-backed implementation is wired in Phase 2 by migration 0002.
"""

from __future__ import annotations

import json
import os
from dataclasses import dataclass, field
from pathlib import Path
from typing import Any


@dataclass
class ShortTermMemory:
    observations: list[str] = field(default_factory=list)
    last_receipts: list[dict[str, Any]] = field(default_factory=list)

    def observe(self, text: str) -> None:
        self.observations.append(text)
        if len(self.observations) > 50:
            self.observations = self.observations[-50:]

    def record_receipt(self, receipt: dict[str, Any]) -> None:
        self.last_receipts.append(receipt)
        if len(self.last_receipts) > 20:
            self.last_receipts = self.last_receipts[-20:]

    def to_dict(self) -> dict[str, Any]:
        return {"observations": list(self.observations), "last_receipts": list(self.last_receipts)}


class LongTermMemory:
    """File-backed long-term memory keyed by agent string_id."""

    def __init__(self, path: str | os.PathLike[str] | None = None) -> None:
        if path is None:
            path = os.environ.get(
                "ECOMATRIX_AGENT_LTM_PATH",
                str(Path.home() / ".ecomatrix" / "ltm.json"),
            )
        self._path = Path(path)
        self._path.parent.mkdir(parents=True, exist_ok=True)
        if not self._path.exists():
            self._path.write_text("{}")

    def _load(self) -> dict[str, Any]:
        return json.loads(self._path.read_text() or "{}")

    def _save(self, data: dict[str, Any]) -> None:
        tmp = self._path.with_suffix(".tmp")
        tmp.write_text(json.dumps(data, indent=2))
        tmp.replace(self._path)

    def get(self, agent_id: str) -> dict[str, Any]:
        return self._load().get(agent_id, {"summary": "", "facts": []})

    def update(self, agent_id: str, *, summary: str | None = None,
               append_fact: str | None = None) -> dict[str, Any]:
        data = self._load()
        entry = data.setdefault(agent_id, {"summary": "", "facts": []})
        if summary is not None:
            entry["summary"] = summary
        if append_fact is not None:
            entry.setdefault("facts", []).append(append_fact)
            entry["facts"] = entry["facts"][-50:]
        data[agent_id] = entry
        self._save(data)
        return entry
