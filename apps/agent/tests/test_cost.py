"""Cost control + human-in-the-loop tests."""
from ecomatrix.cost import CostLedger, needs_human_approval, HUMAN_APPROVAL_THRESHOLD


def test_ledger_starts_empty():
    l = CostLedger()
    r = l.report()
    assert r["tick_used"] == 0
    assert r["cumulative_used"] == 0


def test_ledger_can_spend_within_budget():
    l = CostLedger(budget_per_tick=100, budget_cumulative=1000)
    assert l.can_spend(50)
    assert l.spend(50)
    assert l.tick_used == 50
    assert l.cumulative_used == 50


def test_ledger_rejects_over_budget():
    l = CostLedger(budget_per_tick=10)
    assert l.spend(5)
    assert not l.can_spend(6)  # 5+6=11 > 10
    assert not l.spend(6)


def test_ledger_rejects_over_cumulative():
    l = CostLedger(budget_per_tick=100, budget_cumulative=10)
    assert l.spend(5)
    assert not l.can_spend(6)  # 5+6=11 > 10 cumulative


def test_ledger_reset_tick():
    l = CostLedger(budget_per_tick=100)
    l.spend(50)
    l.reset_tick()
    assert l.tick_used == 0
    assert l.cumulative_used == 50  # cumulative persists


def test_needs_human_approval_below_threshold():
    assert not needs_human_approval(50)


def test_needs_human_approval_at_threshold():
    assert needs_human_approval(HUMAN_APPROVAL_THRESHOLD)


def test_needs_human_approval_above_threshold():
    assert needs_human_approval(150)
