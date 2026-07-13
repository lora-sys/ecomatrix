# ISSUE-028 Verification

Date: 2026-07-13
Branch: `feature/ISSUE-028-hierarchical-supervisor`

## Python

```text
$ cd apps/agent
$ ECOMATRIX_AGENT_TRACES=0 uv run pytest -q
........................................................................ [ 71%]
.............................                                            [100%]
101 passed in 0.44s

$ uv run ruff check ecomatrix/supervisor.py ecomatrix/runner.py tests/test_supervisor.py tests/test_runner.py
All checks passed!
```

Focused supervisor/runner suite before final expansion:

```text
$ uv run pytest -q tests/test_supervisor.py tests/test_runner.py
......................                                                   [100%]
22 passed in 0.54s
```

## Backend

The first parallel race run exposed that repo and service test binaries both
truncated the public schema. After introducing the isolated service schema:

```text
$ cd apps/backend
$ go test -race -count=1 ./...
?    github.com/ecomatrix/backend/cmd/seed                    [no test files]
?    github.com/ecomatrix/backend/cmd/server                  [no test files]
ok   github.com/ecomatrix/backend/internal/auth               2.227s
ok   github.com/ecomatrix/backend/internal/domain             1.013s
ok   github.com/ecomatrix/backend/internal/repo               1.600s
ok   github.com/ecomatrix/backend/internal/service            2.486s
ok   github.com/ecomatrix/backend/internal/transport/http     1.025s
ok   github.com/ecomatrix/backend/pkg/a2a                     1.016s
```

`TestReceipt_JSONShape_MatchesA2AProtocol` pins `tx_id`, `from`, `to`,
`amount`, `currency_type`, and lowercase nested `balance_after` keys.

## Live CLI

Setup:

```text
$ ECOMATRIX_DB_DSN=postgres://.../ecomatrix?sslmode=disable ./bin/seed
{"level":"INFO","msg":"seed complete","agents":11}

$ ECOMATRIX_HTTP_ADDR=127.0.0.1:8081 ECOMATRIX_DB_DSN=postgres://... ./bin/server
{"level":"INFO","msg":"listening","addr":"127.0.0.1:8081"}

$ curl http://127.0.0.1:8081/healthz
{"status":"ok"}
```

Command:

```text
$ ECOMATRIX_AGENT_BACKEND_URL=http://127.0.0.1:8081 \
  ECOMATRIX_AGENT_LLM_PROVIDER=stub ECOMATRIX_AGENT_TRACES=0 \
  uv run python -m ecomatrix.runner \
  --backend http://127.0.0.1:8081 \
  --scenario supervisor \
  --goal 'delegate one safe trade and summarize the receipt' \
  --max-subtasks 1
exit: 0
```

Relevant JSON result:

```json
{
  "worker_results": [
    {
      "agent_id": "agent_hacker_01",
      "final_receipt": {
        "tx_id": "tx_c4025330ef84b6ab9819c8394144df25",
        "to": "agent_merchant_01",
        "amount": 10
      },
      "error": null
    }
  ],
  "final_summary": "Completed 1 subtask(s): 1 succeeded, 0 failed, and 1 produced a receipt.",
  "error": null,
  "cost": {
    "tick_used": 3596,
    "tick_budget": 12000,
    "cumulative_used": 3596,
    "cumulative_budget": 12000
  }
}
```

The isolated backend process was stopped after verification.

## Repository Gate

```text
$ git remote -v
(no output)

$ gh auth status
Logged in to github.com as lora-sys; repo/workflow scopes available.

$ gh repo list lora-sys ... | select(name == "ecomatrix")
(no matching repository)
```

At this checkpoint, push, PR creation, and remote CI were unavailable. The
user later authorized repository creation; the resolution is recorded below.

## Remote CI Recovery

The user authorized creating a remote after the initial local handoff.

```text
Repository: https://github.com/lora-sys/ecomatrix (private)
Issue:      https://github.com/lora-sys/ecomatrix/issues/1
Draft PR:   https://github.com/lora-sys/ecomatrix/pull/2
First run:  https://github.com/lora-sys/ecomatrix/actions/runs/29218912703
```

First-run result:

```text
agent (Python pytest)             pass
backend (Go -race)               fail: six files reported by gofmt -l
frontend (Next.js + Playwright)  fail: next lint requested interactive setup
e2e (Playwright)                 skipped because dependencies failed
```

Recovery verification before push:

```text
$ cd apps/backend
$ go vet ./... && test -z "$(gofmt -l .)" && go test -race -count=1 ./...
all packages passed

$ cd apps/frontend
$ npm run lint
✔ No ESLint warnings or errors

$ npx tsc --noEmit
exit: 0

$ npm run build
✓ Compiled successfully
✓ Generating static pages (3/3)
```
