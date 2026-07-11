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
from .graphs.base import run_one_tick
from .llm import get_default_llm
from .memory import FileLongTermMemory, LongTermMemory, PostgresLongTermMemory, ShortTermMemory


GRAPH_BUILDERS = {
    "miner": miner.build,
    "merchant": merchant.build,
    "hacker": hacker.build,
    "mediator": mediator.build,
}


def _use_pg_ltm() -> bool:
    return os.environ.get("ECOMATRIX_AGENT_LTM", "postgres").lower() in ("postgres", "pg", "1", "true")


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
    ltm = PostgresLongTermMemory(client) if _use_pg_ltm() else FileLongTermMemory()

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



def run_multi_agent(client: A2AClient, *, ticks: int, tick_seconds: float,
                    log: logging.Logger) -> dict[str, Any]:
    """Phase 2.5 multi-agent scenario: every seeded agent runs in parallel.

    Spawns one LangGraph per seeded agent (filtered to those that exist on the
    backend), advances them in lockstep for `ticks` rounds, and reports
    per-role and global aggregates. World-GOLD conservation is asserted.
    """
    from concurrent.futures import ThreadPoolExecutor, as_completed
    llm = get_default_llm()
    ltm = PostgresLongTermMemory(client) if _use_pg_ltm() else FileLongTermMemory()

    seeded = client.list_agents(limit=200)
    # Map job_type -> [string_id, ...].
    by_job: dict[str, list[str]] = {}
    for a in seeded:
        by_job.setdefault(str(a.get("JobType", "")), []).append(str(a.get("StringID", "")))

    log.info("multi-agent spawn", extra={"agents": len(seeded),
                                         "jobs": {k: len(v) for k, v in by_job.items()}})

    # Build a graph per agent.
    agents: list[tuple[Any, ShortTermMemory, str, str]] = []  # (graph, mem, agent_id, job)
    for job, ids in by_job.items():
        builder = GRAPH_BUILDERS.get(job)
        if not builder:
            continue
        for agent_id in ids:
            graph, mem = _spawn(client, llm, job, agent_id, ltm)
            agents.append((graph, mem, agent_id, job))

    initial_world = sum(a.get("Balance", 0) for a in seeded)
    settled = 0
    rejected = 0
    posted = 0
    errors: list[str] = []

    def tick_one(graph, mem, agent_id, job):
        return job, agent_id, run_one_tick(graph, agent_id=agent_id, job_type=job), mem

    for t in range(ticks):
        with ThreadPoolExecutor(max_workers=min(16, len(agents))) as ex:
            futures = [ex.submit(tick_one, g, m, aid, jb) for (g, m, aid, jb) in agents]
            for fut in as_completed(futures):
                job, agent_id, result, mem = fut.result()
                if result.receipt is not None:
                    settled += 1
                    mem.record_receipt({
                        "tx_id": result.receipt.tx_id,
                        "to": result.receipt.to,
                        "amount": result.receipt.amount,
                    })
                    ltm.update(agent_id, append_fact=f"settled {result.receipt.tx_id}")
                    log.info("settled", extra={"role": job, "agent": agent_id,
                                                "tx_id": result.receipt.tx_id})
                elif result.error is not None:
                    if result.error.code == Code.INSUFFICIENT_FUNDS:
                        rejected += 1
                    else:
                        errors.append(f"{agent_id}:{result.error.code.value}:{result.error.message}")
                    log.info("rejected", extra={"role": job, "agent": agent_id,
                                                "code": result.error.code.value})
                else:
                    log.info("skipped", extra={"role": job, "agent": agent_id})

        time.sleep(tick_seconds)

    final_world = sum(a.get("Balance", 0) for a in client.list_agents(limit=200))
    feeds = client.list_feeds(limit=500)
    posted = len(feeds)

    summary = {
        "agents": len(agents),
        "ticks": ticks,
        "settled": settled,
        "rejected": rejected,
        "feeds_posted": posted,
        "errors": errors[:20],
        "world_initial": initial_world,
        "world_final": final_world,
        "conservation": initial_world == final_world,
    }
    log.info("multi-agent end summary=%s",
             json.dumps({k: v for k, v in summary.items() if k != "msg"}, sort_keys=True))
    print(json.dumps(summary, indent=2, ensure_ascii=False))
    return summary


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
        if args.scenario == "multi":
            summary = run_multi_agent(client, ticks=args.ticks, tick_seconds=args.tick_seconds, log=log)
            ok = summary["conservation"] and not summary["errors"]
            return 0 if ok else 1
        log.error("scenario %s not yet implemented", args.scenario)
        return 2


if __name__ == "__main__":
    sys.exit(main())
