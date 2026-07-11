# EcoMatrix root Makefile — orchestrates the three apps.
# Run `make help` to see targets.

SHELL := /usr/bin/env bash
ROOT  := $(CURDIR)

DB_DSN ?= postgres://repotwin:repotwin@localhost:5432/ecomatrix?sslmode=disable
ADMIN_TOKEN ?= dev-admin-token

BACKEND_HTTP ?= :8080
FRONTEND_HTTP ?= :3100
WS_URL ?= ws://localhost:8080/v1/stream

export ECOMATRIX_DB_DSN := $(DB_DSN)
export ECOMATRIX_ADMIN_TOKEN := $(ADMIN_TOKEN)
export ECOMATRIX_HTTP_ADDR := $(BACKEND_HTTP)
export NEXT_PUBLIC_BACKEND_URL := http://localhost$(BACKEND_HTTP)
export NEXT_PUBLIC_WS_URL := $(WS_URL)

.DEFAULT_GOAL := help

.PHONY: help db-up db-down seed backend frontend agent demo test fmt lint clean

help: ## show this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nTargets:\n"} /^[a-zA-Z_-]+:.*##/ { printf "  %-12s %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

db-up: ## start Postgres via docker compose
	docker compose up -d db

db-down: ## stop Postgres
	docker compose down

seed: ## seed the database with 11 deterministic agents
	cd apps/backend && go build -o bin/seed ./cmd/seed && ./bin/seed

backend: ## build + run the Go backend (foreground)
	cd apps/backend && go build -o bin/server ./cmd/server && ./bin/server

frontend: ## run the Next.js dashboard on :3100 (foreground)
	cd apps/frontend && npm install && PORT=3100 npx next dev -p 3100

agent: ## run the multi-agent scenario for 5 ticks
	cd apps/agent && uv venv --python 3.12 .venv && . .venv/bin/activate && uv pip install -e '.[dev]'
	cd apps/agent && . .venv/bin/activate && python -m ecomatrix.runner --scenario multi --ticks 5 --tick-seconds 0.3

demo: ## one-shot: db + migrate + seed + backend (bg) + frontend (bg) + multi-agent (bg)
	bash scripts/demo.sh

test: ## run all tests across the three apps
	cd apps/backend && go test -race -count=1 ./...
	cd apps/agent && . .venv/bin/activate && pytest -q
	cd apps/frontend && npx tsc --noEmit && npx playwright test

fmt: ## gofmt + ruff format
	cd apps/backend && gofmt -s -w .
	cd apps/agent && . .venv/bin/activate && ruff format ecomatrix tests 2>/dev/null || true

lint: ## go vet + ruff check + next lint
	cd apps/backend && go vet ./...
	cd apps/agent && . .venv/bin/activate && ruff check ecomatrix tests 2>/dev/null || true
	cd apps/frontend && npx next lint 2>/dev/null || true

clean: ## remove build artifacts
	rm -rf apps/backend/bin
	rm -rf apps/frontend/.next
	rm -rf apps/frontend/test-results
	rm -rf apps/agent/.venv
