#!/usr/bin/env bash
# Bootstrap script — re-run safe; idempotent.
set -euo pipefail

cd "$(dirname "$0")/.."

mkdir -p \
  docs/{product,architecture,design,decisions,evidence,sessions,.index} \
  memory tasks sessions templates checklists skills scripts \
  .github/{ISSUE_TEMPLATE,workflows} \
  apps/{backend,frontend,agent}

echo "Bootstrap verified."
