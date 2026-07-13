# ISS-029 Verification

## Local Gates

- `cd apps/backend && go test -race -count=1 ./...` — all packages green.
- `cd apps/agent && uv run pytest -q` — 101 passed.
- `cd apps/frontend && npm run typecheck && npm run lint && npm run build` —
  typecheck and lint clean; `build` emits 255 KB First Load JS.

## CI

- Filled in after PR is opened and the first CI run completes.
