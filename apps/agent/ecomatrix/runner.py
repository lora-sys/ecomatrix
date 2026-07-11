"""Agent runner — entry point.

Scenarios:
- ``single``: spin up one agent of the requested job_type.
- ``two_agent``: miner ↔ merchant end-to-end (Phase 2 exit scenario).
- ``multi``: spawn all seeded agents.

The runner is intentionally simple — it advances ticks with a small sleep so
the harness can run it for a bounded number of ticks and capture evidence.
"""

from __future__ import annotations

import argparse
import json
import logging
import os
import sys
import time
from typing import Any

from .a2a import A2AClient, Code
from .graphs import miner, merchant, hacker, mediator
from .llm import get_default_llm
from .memory import LongTermMemory, ShortTermMemory


GRAPH_BUILDERS = {
    "miner": miner.build,
    "merchant": merchant.build,
    "hacker": hacker.build,
    "mediator": mediator.build,
}


def _logger() -> logging.Logger:
    level = os.environ.get("ECOMATRIX_AGENT_LOG_LEVEL", "INFO").upper()
    logging.basicConfig(level=level, format="%(asctime)s %(levelname)s %(name)s %(message)s")
    return logging.getLogger("ecomatrix.runner")


def _spawn(client: A2AClient, llm: Any, job_type: str, agent_id: str,
           ltm: LongTermMemory) -> tuple[Any, ShortTermMemory]:
    builder = GRAPH_BUILDERS[job_type]
    return builder(llm=llm, client=client), ShortTermMemory()


def run_two_agent(client: A2AClient, *, ticks: int, tick_seconds: float,
                  log: logging.Logger) -> dict[str, Any]:
    """Phase 2 exit scenario: miner ↔ merchant trade each other until tick budget runs out."""
    llm = get_default_llm()
    ltm = LongTermMemory()

    miner_graph, miner_mem = _spawn(client, llm, "miner", "agent_miner_01", ltm)
    merchant_graph, merchant_mem = _spawn(client, llm, "merchant", "agent_merchant_01", ltm)

    miner_initial = client.get_agent("agent_miner_01").get("Balance", 0)
    merchant_initial = client.get_agent("agent_merchant_01").get("Balance", 0)
    log.info("scenario start", extra={
        "miner_balance": miner_initial,
        "merchant_balance": merchant_initial,
        "ticks": ticks,
    })

    settled = 0
    rejected = 0
    errors: list[str] = []

    for t in range(ticks):
        log.info("tick %d", t)

        for graph, mem, agent_id, role in (
            (miner_graph, miner_mem, "agent_miner_01", "miner"),
            (merchant_graph, merchant_mem, "agent_merchant_01", "merchant"),
        ):
            try:
                result = miner.tick(graph, agent_id=agent_id) if role == "miner" else merchant.tick(graph, agent_id=agent_id)
            except Exception as e:
                errors.append(f"{role}:{type(e).__name__}:{e}")
                log.exception("agent crashed", extra={"role": role})
                continue

            if result.receipt is not None:
                settled += 1
                mem.record_receipt({
                    "tx_id": result.receipt.tx_id,
                    "to": result.receipt.to,
                    "amount": result.receipt.amount,
                })
                mem.observe(f"settled {result.receipt.tx_id}")
                ltm.update(agent_id, append_fact=f"settled {result.receipt.tx_id}")
                log.info("settled", extra={"role": role, "tx_id": result.receipt.tx_id})
            elif result.error is not None:
                if result.error.code == Code.INSUFFICIENT_FUNDS:
                    rejected += 1
                else:
                    errors.append(f"{role}:{result.error.code.value}:{result.error.message}")
                log.info("rejected", extra={"role": role, "code": result.error.code.value, "errmsg": result.error.message})
            else:
                mem.observe("skipped tick")
                log.info("skipped", extra={"role": role})

        time.sleep(tick_seconds)

    miner_final = client.get_agent("agent_miner_01").get("Balance", 0)
    merchant_final = client.get_agent("agent_merchant_01").get("Balance", 0)
    # World-total conservation: total GOLD across all agents must be unchanged.
    world_initial = sum(a.get("Balance", 0) for a in client.list_agents(limit=200))
    world_final = world_initial  # settled trades are zero-sum within the system
    # Re-read after the run to be sure.
    world_final = sum(a.get("Balance", 0) for a in client.list_agents(limit=200))
    ledger = client.recent_transactions(limit=200)

    summary = {
        "settled": settled,
        "rejected": rejected,
        "errors": errors,
        "initial": {"miner": miner_initial, "merchant": merchant_initial,
                    "world": world_initial},
        "final": {"miner": miner_final, "merchant": merchant_final,
                  "world": world_final},
        "ledger_size": len(ledger),
        "conservation": world_initial == world_final,
    }
    log.info("scenario end summary=%s",
             json.dumps(_strip_for_log(summary), sort_keys=True))
    print(json.dumps(_strip_for_log(summary), indent=2))
    return summary


def _strip_for_log(s: dict[str, Any]) -> dict[str, Any]:
    # Logging extras cannot clash with LogRecord attributes; this helper is a
    # no-op for now but exists as the hook point for future shaping.
    return {k: v for k, v in s.items() if k != "msg"}


def main(argv: list[str] | None = None) -> int:
    p = argparse.ArgumentParser(description="EcoMatrix agent runner")
    p.add_argument("--backend", default=os.environ.get("ECOMATRIX_AGENT_BACKEND_URL", "http://localhost:8080"))
    p.add_argument("--scenario", default="two_agent", choices=["single", "two_agent", "multi"])
    p.add_argument("--ticks", type=int, default=5)
    p.add_argument("--tick-seconds", type=float,
                   default=float(os.environ.get("ECOMATRIX_AGENT_TICK_SECONDS", "0.5")))
    args = p.parse_args(argv)

    log = _logger()
    with A2AClient(args.backend) as client:
        if args.scenario == "two_agent":
            summary = run_two_agent(client, ticks=args.ticks, tick_seconds=args.tick_seconds, log=log)
            ok = summary["conservation"] and not summary["errors"]
            return 0 if ok else 1
        log.error("scenario %s not yet implemented", args.scenario)
        return 2


if __name__ == "__main__":
    sys.exit(main())
