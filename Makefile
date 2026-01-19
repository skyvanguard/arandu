.PHONY: all build dev test lint clean install help
.PHONY: backend-dev frontend-dev backend-test frontend-test
.PHONY: docker-build docker-up docker-down docker-logs
.PHONY: generate codegen migrate

# Default target
all: install build

# =============================================================================
# INSTALLATION
# =============================================================================

install: backend-install frontend-install ## Install all dependencies

backend-install: ## Install backend dependencies
	cd backend && go mod download

frontend-install: ## Install frontend dependencies
	cd frontend && yarn install

# =============================================================================
# DEVELOPMENT
# =============================================================================

dev: ## Run both backend and frontend in development mode
	@echo "Starting development servers..."
	@make -j2 backend-dev frontend-dev

backend-dev: ## Run backend in development mode
	cd backend && go run .

frontend-dev: ## Run frontend in development mode
	cd frontend && yarn dev

# =============================================================================
# BUILD
# =============================================================================

build: backend-build frontend-build ## Build both backend and frontend

backend-build: ## Build backend binary
	cd backend && go build -o arandu .

frontend-build: ## Build frontend for production
	cd frontend && yarn build

# =============================================================================
# TESTING
# =============================================================================

test: backend-test frontend-test ## Run all tests

backend-test: ## Run backend tests
	cd backend && go test -v ./...

backend-test-cover: ## Run backend tests with coverage
	cd backend && go test -coverprofile=coverage.out ./...
	cd backend && go tool cover -html=coverage.out -o coverage.html

frontend-test: ## Run frontend tests
	cd frontend && yarn test

frontend-test-cover: ## Run frontend tests with coverage
	cd frontend && yarn test:coverage

# =============================================================================
# LINTING & FORMATTING
# =============================================================================

lint: backend-lint frontend-lint ## Run all linters

backend-lint: ## Run backend linter
	cd backend && golangci-lint run

frontend-lint: ## Run frontend linter
	cd frontend && yarn lint

format: backend-format frontend-format ## Format all code

backend-format: ## Format backend code
	cd backend && go fmt ./...

frontend-format: ## Format frontend code
	cd frontend && yarn format:fix

# =============================================================================
# CODE GENERATION
# =============================================================================

generate: backend-generate frontend-generate ## Generate all code

backend-generate: ## Generate backend code (GraphQL, SQL)
	cd backend && go generate ./...

frontend-generate: ## Generate frontend GraphQL types
	cd frontend && yarn codegen

codegen: generate ## Alias for generate

# =============================================================================
# DATABASE
# =============================================================================

migrate: ## Run database migrations
	cd backend && goose -dir migrations sqlite3 database.db up

migrate-down: ## Rollback last migration
	cd backend && goose -dir migrations sqlite3 database.db down

migrate-status: ## Show migration status
	cd backend && goose -dir migrations sqlite3 database.db status

# =============================================================================
# DOCKER
# =============================================================================

docker-build: ## Build Docker image
	docker build -t arandu .

docker-up: ## Start all services with Docker Compose
	docker-compose up -d

docker-down: ## Stop all services
	docker-compose down

docker-logs: ## Show logs from all services
	docker-compose logs -f

docker-dev: ## Start development environment with Docker Compose
	docker-compose --profile dev up -d

# =============================================================================
# OLLAMA
# =============================================================================

ollama-pull: ## Pull recommended Ollama model
	ollama pull qwen2.5-coder:14b

ollama-pull-small: ## Pull smaller Ollama model (for limited hardware)
	ollama pull qwen2.5-coder:7b

# =============================================================================
# CLEANUP
# =============================================================================

clean: ## Clean build artifacts
	rm -f backend/arandu backend/arandu.exe
	rm -rf frontend/dist
	rm -f backend/coverage.out backend/coverage.html

clean-all: clean ## Clean everything including dependencies
	rm -rf frontend/node_modules
	rm -rf backend/database.db

# =============================================================================
# HELP
# =============================================================================

help: ## Show this help message
	@echo "Arandu - AI Coding Assistant"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'
