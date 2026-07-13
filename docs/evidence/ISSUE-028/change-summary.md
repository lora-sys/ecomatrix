# ISSUE-028 Change Summary

## Shipped

- Added a bounded hierarchical supervisor that validates one goal, normalizes
  at most four LLM-generated subtasks, routes deterministically, delegates via
  the existing ReAct workflow, and aggregates receipts and failures.
- Added one shared `CostLedger` wrapper across decomposition, every worker LLM
  call, reflection, and final aggregation. The result exposes the consumed and
  configured token budget.
- Built the worker registry from live seeded agents and included each formal
  Agent Contract's mission, capabilities, and limitations in decomposition.
- Added `--scenario supervisor --goal ... --max-subtasks ...` without changing
  `two_agent` or `multi` behavior. Any worker error produces a non-zero exit
  after all delegated results have been emitted.
- Restored the Go receipt's A2A v1.1 snake_case JSON tags. Live verification
  found this Phase 5 regression when the new CLI received PascalCase fields.
- Isolated service tests in a per-process Postgres schema so repo and service
  packages can pass the documented parallel race-detector command reliably.

## Risk And Rollback

- Supervisor execution is deliberately sequential and capped at four subtasks
  and three ReAct iterations per worker.
- Cost estimates use a conservative character-to-token approximation and
  charge attempted provider calls because timed-out upstream work may still be
  billable.
- Rollback is code-only: remove the supervisor scenario/module and revert the
  receipt tags/test schema helper. No production migration or data rewrite was
  introduced.
