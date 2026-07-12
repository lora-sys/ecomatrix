"""Memory compression tests."""
from ecomatrix.memory import ShortTermMemory, summarize_short_term


def test_empty_short_term_returns_empty():
    m = ShortTermMemory()
    assert summarize_short_term(m) == ""


def test_short_term_under_3_returns_joined():
    m = ShortTermMemory()
    m.observe("a")
    m.observe("b")
    assert summarize_short_term(m) == "a | b"


def test_short_term_over_3_uses_heuristic_without_llm():
    m = ShortTermMemory()
    for i in range(10):
        m.observe(f"obs{i}")
    summary = summarize_short_term(m)  # no llm
    assert "obs0" in summary  # head kept
    assert "obs9" in summary  # tail kept
    assert "obs4" not in summary or "..." in summary  # middle is dropped or marked


def test_short_term_with_llm_uses_llm():
    from ecomatrix.llm import StubLLM
    m = ShortTermMemory()
    for i in range(10):
        m.observe(f"obs{i}")
    summary = summarize_short_term(m, llm=StubLLM())
    # StubLLM returns a JSON; we just want it not to crash.
    assert summary  # non-empty
