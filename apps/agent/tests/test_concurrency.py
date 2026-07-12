"""Concurrency stress test for the LLM + tool loop.

50 concurrent LLM calls (30% failure rate) and 100 concurrent tool calls
with sender isolation. No real backend required.
"""
import os
import time
import threading
import concurrent.futures

import pytest

from ecomatrix.llm import MockLLMWithFailures, LLMError, LLMRateLimitError
from ecomatrix.tools import execute_tool, ToolResult


def test_50_concurrent_llm_calls_with_partial_failure():
    """50 concurrent LLM calls; 30% fail. None should crash."""
    llm = MockLLMWithFailures(failure_rate=0.3, failure_kind=LLMRateLimitError)
    n = 50
    results = {"ok": 0, "error": 0}
    lock = threading.Lock()

    def call(i: int):
        try:
            r = llm.complete([{"role": "user", "content": f"agent_id=agent_miner_{i:02d} target=agent_merchant_01 amount=10"}])
            with lock:
                results["ok"] += 1
            return r
        except LLMError:
            with lock:
                results["error"] += 1
            return None

    t0 = time.time()
    with concurrent.futures.ThreadPoolExecutor(max_workers=20) as ex:
        futures = [ex.submit(call, i) for i in range(n)]
        concurrent.futures.wait(futures, timeout=15)
    elapsed = time.time() - t0

    # Note: it's possible some futures didn't complete in 15s, but MockLLM is sync
    # so this should be instant.
    done = sum(1 for f in futures if f.done())
    assert done == n, f"only {done}/{n} futures completed"

    assert results["ok"] + results["error"] == n
    assert 0.20 <= results["ok"] / n <= 0.80, f"unusual ratio: {results}"
    print(f"  {n} concurrent LLM calls: ok={results['ok']} error={results['error']} in {elapsed:.3f}s")


def test_concurrent_tool_execution_no_cross_contamination():
    """100 tool calls across 50 senders; each result attributed to its sender."""
    class CapturingClient:
        def __init__(self):
            self.calls = []
            self.lock = threading.Lock()

        def get_agent(self, agent_id):
            with self.lock:
                self.calls.append(("get", agent_id))
            return {"StringID": agent_id, "Balance": 100}

        def execute_trade(self, *, sender, target_agent, amount, reasoning):
            from ecomatrix.a2a import Receipt
            with self.lock:
                self.calls.append(("trade", sender, target_agent, amount))
            return Receipt(
                tx_id=f"tx_{sender}", from_=sender, to=target_agent,
                amount=amount, currency_type="GOLD",
                balance_after_from=0, balance_after_to=0,
            )

    c = CapturingClient()
    n = 50

    def tick(i: int):
        sender = f"agent_x_{i:02d}"
        r1 = execute_tool("get_agent_state", {}, client=c, sender=sender)
        r2 = execute_tool("execute_trade", {"target_agent": "agent_y", "amount": 5}, client=c, sender=sender)
        assert r1.ok and r2.ok
        assert r1.result["string_id"] == sender
        return sender

    t0 = time.time()
    with concurrent.futures.ThreadPoolExecutor(max_workers=20) as ex:
        futures = [ex.submit(tick, i) for i in range(n)]
        concurrent.futures.wait(futures, timeout=15)
    elapsed = time.time() - t0

    assert len(c.calls) == 2 * n
    trade_senders = {c[1] for c in c.calls if c[0] == "trade"}
    assert trade_senders == {f"agent_x_{i:02d}" for i in range(n)}
    print(f"  {2*n} tool calls in {elapsed:.3f}s, no cross-contamination")
