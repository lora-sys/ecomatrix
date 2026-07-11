#!/usr/bin/env bash
# scripts/demo.sh — bring up the full EcoMatrix stack in one command.
#
# What it does:
#   1. Ensures Postgres is running (docker compose up -d db if not).
#   2. Applies migrations + seeds 11 agents.
#   3. Starts the Go backend on :8080.
#   4. Installs frontend deps + starts Next.js on :3100.
#   5. Spawns the multi-agent scenario in the background (continuous trading).
#   6. Prints the URLs to visit and tails the backend log.
#
# Stop everything with Ctrl-C; the script traps SIGINT and kills children.

set -e
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
cd "$ROOT"

DB_DSN="${ECOMATRIX_DB_DSN:-postgres://repotwin:repotwin@localhost:5432/ecomatrix?sslmode=disable}"
ADMIN_TOKEN="${ECOMATRIX_ADMIN_TOKEN:-dev-admin-token}"
BACKEND_HTTP="${ECOMATRIX_HTTP_ADDR:-:8080}"
FRONTEND_PORT="${ECOMATRIX_FRONTEND_PORT:-3100}"

export ECOMATRIX_DB_DSN="$DB_DSN"
export ECOMATRIX_ADMIN_TOKEN="$ADMIN_TOKEN"
export ECOMATRIX_HTTP_ADDR="$BACKEND_HTTP"
export NEXT_PUBLIC_BACKEND_URL="http://localhost$BACKEND_HTTP"
export NEXT_PUBLIC_WS_URL="ws://localhost$BACKEND_HTTP/v1/stream"

LOG_DIR="$ROOT/.demo-logs"
mkdir -p "$LOG_DIR"

PIDS=()
cleanup() {
  echo
  echo "demo.sh: shutting down..."
  for pid in "${PIDS[@]}"; do
    kill -TERM "$pid" 2>/dev/null || true
  done
  sleep 0.5
  for pid in "${PIDS[@]}"; do
    kill -KILL "$pid" 2>/dev/null || true
  done
}
trap cleanup INT TERM EXIT

# 1. Postgres.
if ! docker ps --format '{{.Names}}' | grep -q "ecomatrix-postgres"; then
  echo "[1/5] starting Postgres via docker compose..."
  docker compose up -d db
  for i in $(seq 1 30); do
    if docker exec ecomatrix-postgres pg_isready -U repotwin >/dev/null 2>&1; then
      break
    fi
    sleep 1
  done
else
  echo "[1/5] Postgres already running."
fi

# 2. Build backend + seed.
echo "[2/5] building backend + seeding..."
mkdir -p apps/backend/bin
(cd apps/backend && go build -o bin/server ./cmd/server && go build -o bin/seed ./cmd/seed)
apps/backend/bin/seed

# 3. Start backend in background.
echo "[3/5] starting backend on $BACKEND_HTTP..."
(cd apps/backend && ./bin/server) > "$LOG_DIR/backend.log" 2>&1 &
PIDS+=($!)
for i in $(seq 1 15); do
  if curl -fs "http://localhost$BACKEND_HTTP/healthz" >/dev/null 2>&1; then
    break
  fi
  sleep 1
done

# 4. Start frontend in background.
echo "[4/5] starting frontend on :$FRONTEND_PORT..."
if [ ! -d apps/frontend/node_modules ]; then
  (cd apps/frontend && npm install --no-audit --no-fund)
fi
(cd apps/frontend && PORT=$FRONTEND_PORT npx next dev -p $FRONTEND_PORT) > "$LOG_DIR/frontend.log" 2>&1 &
PIDS+=($!)
for i in $(seq 1 60); do
  if curl -fs "http://localhost:$FRONTEND_PORT/" -o /dev/null 2>&1; then
    break
  fi
  sleep 1
done

# 5. Start multi-agent scenario in background.
echo "[5/5] spawning multi-agent scenario (continuous)..."
if [ ! -d apps/agent/.venv ]; then
  (cd apps/agent && uv venv --python 3.12 .venv && \
   . .venv/bin/activate && \
   uv pip install -e '.[dev]')
fi
(cd apps/agent && \
  . .venv/bin/activate && \
  python -m ecomatrix.runner --scenario multi --ticks 999 --tick-seconds 0.5) \
  > "$LOG_DIR/agent.log" 2>&1 &
PIDS+=($!)

echo
echo "================================================================"
echo "  EcoMatrix is LIVE"
echo "================================================================"
echo "  Dashboard:  http://localhost:$FRONTEND_PORT"
echo "  Backend:    http://localhost$BACKEND_HTTP/v1/metrics"
echo "  Health:     http://localhost$BACKEND_HTTP/healthz"
echo
echo "  Logs (tail -f):"
echo "    $LOG_DIR/backend.log"
echo "    $LOG_DIR/frontend.log"
echo "    $LOG_DIR/agent.log"
echo "================================================================"
echo "  Press Ctrl-C to stop everything."
echo

# Wait forever; cleanup() handles termination.
wait
