#!/usr/bin/env bash
# Start a new multi-agent session log.
# Usage: scripts/new-session.sh "<goal>"
set -euo pipefail

cd "$(dirname "$0")/.."
goal="${1:-unnamed-session}"
ts=$(date +%Y%m%d-%H%M%S)
slug=$(echo "$goal" | tr '[:upper:] ' '[:lower:]-' | tr -cd 'a-z0-9-' | head -c 40)
file="sessions/${ts}-${slug}.md"
cat > "$file" <<HDR
# Session — ${goal}

- Started: $(date -u +"%Y-%m-%dT%H:%M:%SZ")
- Coordinator: Codex
- Phase: Phase 1

## Plan
1.

## Agents Spawned
- (none yet)

## Decisions
- (none yet)

## Outcomes
- (none yet)
HDR
echo "$file"
