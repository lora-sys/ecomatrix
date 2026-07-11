SHELL := /usr/bin/env bash
BACKEND := $(CURDIR)

.PHONY: help build test test-race lint vet tidy run seed migrate-up migrate-down db-up db-down fmt

help:
	@echo "Targets: build test test-race lint vet tidy run seed migrate-up migrate-down db-up db-down fmt"

build:
	go build -o bin/server ./cmd/server
	go build -o bin/seed ./cmd/seed

test:
	go test ./...

test-race:
	go test -race -count=1 ./...

vet:
	go vet ./...

tidy:
	go mod tidy

fmt:
	gofmt -s -w .

run: build
	./bin/server

seed: build
	./bin/seed

migrate-up:
	go run ./cmd/server -h 2>/dev/null; ./bin/server &
	@echo "(server will auto-migrate on boot)"

migrate-down:
	@echo "Manual: docker exec -it <pg> psql -U repotwin -d ecomatrix -f migrations/0001_init.down.sql"

db-up:
	docker compose -f $(BACKEND)/../../docker-compose.yml up -d db

db-down:
	docker compose -f $(BACKEND)/../../docker-compose.yml down

lint:
	@command -v golangci-lint >/dev/null && golangci-lint run || echo "golangci-lint not installed; skipping"
