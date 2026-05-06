.PHONY: run build test test-coverage docker-up docker-down migrate lint

run:
	go run ./cmd/api

build:
	go build -ldflags="-w -s" -o bin/api ./cmd/api

test:
	go test ./... -v

test-coverage:
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

migrate:
	@echo "Running migrations..."
	psql $(DATABASE_URL) -f migrations/001_create_payments.sql

lint:
	golangci-lint run ./...

.DEFAULT_GOAL := run
