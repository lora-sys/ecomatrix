"""Evaluation framework tests."""
from ecomatrix.eval import (
    run_eval, EvalCase, EvalResult, EvalReport,
    DEFAULT_CASES,
    case_stubllm_executes_trade, case_handles_llm_failure,
    case_respects_max_iterations, case_contract_loadable,
    case_human_approval_threshold,
)
import time


def test_eval_report_pass_rate():
    r = EvalReport(total=10, passed=7)
    assert r.pass_rate() == 0.7


def test_eval_report_avg_duration():
    r = EvalReport(total=10, total_duration_ms=200)
    assert r.avg_duration_ms() == 20.0


def test_run_eval_passes_and_fails():
    cases = [
        EvalCase("p1", "passing", run=lambda: EvalResult(case="p1", passed=True, duration_ms=10)),
        EvalCase("p2", "failing", run=lambda: EvalResult(case="p2", passed=False, duration_ms=20)),
    ]
    rep = run_eval(cases)
    assert rep.total == 2
    assert rep.passed == 1
    assert rep.total_duration_ms == 30


def test_run_eval_catches_exception_in_case():
    def boom():
        raise RuntimeError("kaboom")
    cases = [EvalCase("boom", "always fails", run=boom)]
    rep = run_eval(cases)
    assert rep.total == 1
    assert rep.passed == 0
    assert "kaboom" in rep.cases[0].notes


def test_to_dict_shape():
    r = EvalReport(total=2, passed=1, total_duration_ms=100, total_cost_tokens=50)
    d = r.to_dict()
    assert d["total"] == 2
    assert d["passed"] == 1
    assert d["pass_rate"] == 0.5
    assert d["avg_duration_ms"] == 50.0
    assert d["total_cost_tokens"] == 50
    assert isinstance(d["cases"], list)


def test_default_cases_present():
    assert len(DEFAULT_CASES) == 5
    names = {c.name for c in DEFAULT_CASES}
    assert "case_stubllm_executes_trade" in names
    assert "case_human_approval_threshold" in names


def test_golden_eval_passes():
    """Run the full default eval suite. Should have at least 80% pass rate."""
    rep = run_eval(DEFAULT_CASES)
    assert rep.pass_rate() >= 0.8, f"golden eval pass rate {rep.pass_rate():.0%} below 80%: {[(c.case, c.passed, c.notes) for c in rep.cases]}"
