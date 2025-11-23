.PHONY: help build run test test-integration test-coverage lint clean migrate-up migrate-down migrate-status migrate-create migrate-reset docker-build docker-up docker-down docker-logs load-test

help:
	@echo "Available commands:"
	@echo "  make build                - Build the service"
	@echo "  make run                  - Run the service locally"
	@echo "  make test                 - Run unit tests"
	@echo "  make test-integration     - Run integration tests"
	@echo "  make lint                 - Run linter"
	@echo "  make load-test            - Run Vegeta load test"

build:
	CGO_ENABLED=0 GOOS=linux go build -o bin/service cmd/main.go

run:
	@if [ ! -f .env ]; then cp .env.example .env; fi
	go run cmd/main.go

test:
	go test -v -race -cover ./internal/...

lint:
	golangci-lint run ./...

load-test:
	go run cmd/loadtest/main.go

