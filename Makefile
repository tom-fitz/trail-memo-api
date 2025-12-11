.PHONY: help run build test clean migrate docker-build docker-run dev-setup dev-start dev-stop dev-logs dev-db dev-reset

# Load environment variables from .env file if it exists
ifneq (,$(wildcard .env))
    include .env
    export
endif

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

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
		echo "Error: DATABASE_URL environment variable is not set"; \
		exit 1; \
	fi
	@echo "DATABASE_URL: $(DATABASE_PUBLIC_URL)"
	@echo "🗄️  Running database migrations..."
	@for file in migrations/*.sql; do \
		echo "  📄 Applying $$file..."; \
		psql $(DATABASE_PUBLIC_URL) -f $$file || exit 1; \
	done
	@echo "✅ All migrations applied successfully!"

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
	@echo "🔥 Starting development server with hot reload..."
	@echo "📍 Make sure Docker Compose is running: make dev-start"
	air

# ============================================
# Development Environment Commands
# ============================================

dev-setup: ## Setup local development environment
	@echo "🚀 Setting up development environment..."
	@if [ ! -f .env ]; then \
		echo "📝 Creating .env from template..."; \
		cp env.development.example .env; \
		echo "⚠️  Please edit .env with your Firebase credentials"; \
	else \
		echo "✅ .env file already exists"; \
	fi
	@echo "🐳 Starting Docker containers..."
	docker-compose up -d
	@echo "⏳ Waiting for PostgreSQL to be ready..."
	@sleep 5
	@echo "🗄️  Running database migrations..."
	$(MAKE) migrate
	@echo "✅ Development environment ready!"
	@echo ""
	@echo "Next steps:"
	@echo "  1. Edit .env with your Firebase credentials"
	@echo "  2. Run: make dev"

dev-start: ## Start development Docker containers
	@echo "🐳 Starting Docker containers..."
	docker-compose up -d
	@echo "✅ Containers started!"
	@echo "📊 PostgreSQL: localhost:5432"
	@echo "🔍 Run 'make dev-logs' to see logs"

dev-stop: ## Stop development Docker containers
	@echo "🛑 Stopping Docker containers..."
	docker-compose down
	@echo "✅ Containers stopped!"

dev-logs: ## Show Docker container logs
	docker-compose logs -f

dev-db: ## Connect to development database
	@echo "🗄️  Connecting to development database..."
	@echo "Password: trailmemo_dev_password"
	psql postgresql://trailmemo:trailmemo_dev_password@localhost:5432/trailmemo_dev

dev-reset: ## Reset development database (WARNING: deletes all data)
	@echo "⚠️  This will delete all development data!"
	@read -p "Are you sure? [y/N] " -n 1 -r; \
	echo; \
	if [[ $$REPLY =~ ^[Yy]$$ ]]; then \
		echo "🗑️  Stopping containers and removing data..."; \
		docker-compose down -v; \
		echo "🚀 Restarting containers..."; \
		docker-compose up -d; \
		echo "⏳ Waiting for PostgreSQL..."; \
		sleep 5; \
		echo "🗄️  Running migrations..."; \
		$(MAKE) migrate; \
		echo "✅ Database reset complete!"; \
	else \
		echo "❌ Cancelled"; \
	fi

dev-pgadmin: ## Start pgAdmin for database management
	@echo "🔧 Starting pgAdmin..."
	docker-compose --profile tools up -d pgadmin
	@echo "✅ pgAdmin started at http://localhost:5050"
	@echo "📧 Email: admin@trailmemo.local"
	@echo "🔑 Password: admin"

dev-seed: ## Seed database with test data
	@if [ -z "$(TOKEN)" ]; then \
		echo "❌ Error: TOKEN not set"; \
		echo "Get a token first:"; \
		echo "  1. Sign in to get token"; \
		echo "  2. export TOKEN=\"your_token\""; \
		echo "  3. make dev-seed"; \
		exit 1; \
	fi
	@./seed-data.sh

.DEFAULT_GOAL := help

