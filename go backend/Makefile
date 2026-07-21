.PHONY: dev build test tidy redis-up redis-down migrate-up migrate-down migrate-fresh migrate-create

# ── Ishga tushirish ────────────────────────────────────
dev:
	go run cmd/api/main.go

build:
	go build -o bin/api cmd/api/main.go

test:
	go test ./... -v -race

tidy:
	go mod tidy

# ── Redis (Docker) ─────────────────────────────────────
redis-up:
	docker compose up -d redis

redis-down:
	docker compose down

# ── Migratsiya (goose) ─────────────────────────────────
# O'rnatish: go install github.com/pressly/goose/v3/cmd/goose@latest
GOOSE_DRIVER ?= postgres
GOOSE_DBSTRING ?= "host=localhost port=5432 user=postgres password=postgres dbname=ttysi_fit_dev sslmode=disable"

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
