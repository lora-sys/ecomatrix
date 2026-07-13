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

Local commit is possible, but push, PR creation, and remote CI are unavailable
until a Git remote is explicitly configured.
