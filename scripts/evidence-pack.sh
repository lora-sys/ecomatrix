#!/usr/bin/env bash
# Create a fresh evidence pack folder for an Issue.
# Usage: scripts/evidence-pack.sh ISSUE-006
set -euo pipefail

cd "$(dirname "$0")/.."
id="${1:?usage: evidence-pack.sh ISSUE-XXX}"
mkdir -p "docs/evidence/$id/test-results"
mkdir -p "docs/evidence/$id/screenshots"
cat > "docs/evidence/$id/change-summary.md" <<MD
# Change Summary — $id

## What
TODO

## Why
TODO
MD
cat > "docs/evidence/$id/verification.md" <<MD
# Verification — $id

## Commands Run
\`\`\`
$ <command>
\`\`\`

## Result
TODO
MD
cat > "docs/evidence/$id/review-report.md" <<MD
# Review Report — $id

(aggregated by review-aggregator agent)
MD
echo "Created docs/evidence/$id/"
