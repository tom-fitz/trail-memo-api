.PHONY: help run build test clean migrate docker-build docker-run

# Load environment variables from .env file if it exists
ifneq (,$(wildcard .env))
    include .env
    export
endif

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

run: ## Run the application locally
	go run cmd/server/main.go

build: ## Build the application binary
	go build -o bin/server cmd/server/main.go

test: ## Run tests
	go test -v ./...

clean: ## Clean build artifacts
	rm -rf bin/
	rm -rf dist/

migrate: ## Run database migrations
	@if [ -z "$(DATABASE_PUBLIC_URL)" ]; then \
		echo "Error: DATABASE_PUBLIC_URL environment variable is not set"; \
		exit 1; \
	fi
	psql $(DATABASE_PUBLIC_URL) -f migrations/001_init.sql

docker-build: ## Build Docker image
	docker build -t trailmemo-api .

docker-run: ## Run Docker container
	docker run -p 8080:8080 --env-file .env trailmemo-api

install: ## Install dependencies
	go mod download
	go mod tidy

lint: ## Run linter
	golangci-lint run

dev: ## Run with hot reload (requires air)
	air

.DEFAULT_GOAL := help

