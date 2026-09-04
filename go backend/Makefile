.PHONY: dev build test tidy db-up db-down redis-up redis-down migrate-up migrate-down migrate-fresh migrate-create

# ── Ishga tushirish ────────────────────────────────────
dev:
	go run cmd/api/main.go

build:
	go build -o bin/api cmd/api/main.go

test:
	go test ./... -v -race

tidy:
	go mod tidy

# ── Postgres + Redis (Docker) ──────────────────────────
# Compose o'zgaruvchilari .env.local dan (default .env bo'sh).
DC := docker compose --env-file .env.local

db-up:
	$(DC) up -d postgres redis

db-down:
	$(DC) down

redis-up:
	$(DC) up -d redis

redis-down:
	$(DC) down

# ── Migratsiya (goose) ─────────────────────────────────
# O'rnatish: go install github.com/pressly/goose/v3/cmd/goose@latest
GOOSE_DRIVER ?= postgres
# Parol .env.local da — Makefile ga hardcode qilinmaydi (§17.2).
include .env.local
export
GOOSE_DBSTRING ?= "host=$(DB_HOST) port=$(DB_PORT) user=$(DB_USER) password=$(DB_PASSWORD) dbname=$(DB_NAME) sslmode=$(DB_SSLMODE)"

migrate-up:
	goose -dir migrations $(GOOSE_DRIVER) $(GOOSE_DBSTRING) up

migrate-down:
	goose -dir migrations $(GOOSE_DRIVER) $(GOOSE_DBSTRING) down

migrate-fresh:
	goose -dir migrations $(GOOSE_DRIVER) $(GOOSE_DBSTRING) reset
	goose -dir migrations $(GOOSE_DRIVER) $(GOOSE_DBSTRING) up

# make migrate-create name=create_activities_table
migrate-create:
	goose -dir migrations create $(name) sql
