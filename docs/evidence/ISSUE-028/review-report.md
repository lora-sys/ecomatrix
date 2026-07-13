# ISSUE-028 Review Report

Reviewers ran as independent read-only Codex sessions against the working-tree
diff. `paseo` was unavailable in `PATH`, so the local Codex CLI was used with a
read-only sandbox.

## Bug Hunter

Initial result: no Critical/High findings. Medium findings requested explicit
partial-failure exit semantics, empty aggregation tests, and a `multi` scenario
regression test.

Disposition:

- Fixed: partial worker failure is documented and tested as exit code 1 after
  all results are emitted.
- Fixed: aggregation error, empty summary, and missing summary all use a
  deterministic non-empty fallback.
- Fixed: both `two_agent` and `multi` CLI dispatch have regression tests.
- Fixed after live discovery: PascalCase Go receipts are restored to A2A v1.1
  snake_case and pinned by a serialization test.
- Final casing concern rejected with stronger evidence: live `/v1/agents`
  returned `StringID`/`JobType`, the registry built 13 supported workers, and
  the real CLI completed with exit 0.

## Architecture Reviewer

Initial High findings:

- Missing shared cost ceiling across supervisor and worker calls.
- Decomposition saw only worker mission, not capabilities/limitations.

Disposition:

- Fixed: `_BudgetedLLM` wraps decomposition, all ReAct calls, reflections, and
  aggregation with one 12,000-token `CostLedger`.
- Fixed: `WorkerSpec` carries contract capabilities and limitations and renders
  them in the decomposition prompt.
- Fixed: registry construction is separate and routing ties are stable by
  agent ID.
- Fixed: service tests migrate once in a per-process schema and clean it in a
  deferred teardown; the original parallel race command passes.
- Rejected: refunding a failed provider call. Timeout/rate-limit failures may
  still consume upstream tokens, so charging attempted calls is the safe
  bounded-cost policy and prevents retries from bypassing the ceiling.
- Rejected: parsing natural-language limitations in the router. All current
  contracts expose the same three tools; hard trade invariants remain enforced
  by A2A/backend, while the supervisor supplies full contract context to the
  LLM and delegates execution to the contract-aware ReAct workflow.

## Final Gate

- Critical: 0 open
- High: 0 open
- Medium: 0 blocking; residual suggestions above are either fixed or
  dispositioned with current runtime evidence.
