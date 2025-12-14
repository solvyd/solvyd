.PHONY: help build run test clean docker-up docker-down db-migrate db-seed

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the API server
	go build -o bin/ritmo-server cmd/server/main.go

run: ## Run the API server
	go run cmd/server/main.go

test: ## Run tests
	go test -v ./...

test-coverage: ## Run tests with coverage
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

clean: ## Clean build artifacts
	rm -rf bin/
	rm -f coverage.out

deps: ## Download dependencies
	go mod download
	go mod tidy

docker-up: ## Start infrastructure (PostgreSQL, Redis, MinIO, etc.)
	docker-compose up -d
	@echo "Waiting for PostgreSQL to be ready..."
	@sleep 5

docker-down: ## Stop infrastructure
	docker-compose down

docker-clean: ## Stop and remove all containers and volumes
	docker-compose down -v

db-migrate: ## Run database migrations
	@echo "Running database migrations..."
	PGPASSWORD=ritmo_dev_password psql -h localhost -U ritmo -d ritmo -f ../database/schema.sql

db-seed: ## Seed database with sample data
	@echo "Seeding database..."
	PGPASSWORD=ritmo_dev_password psql -h localhost -U ritmo -d ritmo -f ../database/seed.sql

dev: docker-up db-migrate db-seed ## Start development environment
	@echo "Development environment ready!"
	@echo "PostgreSQL: localhost:5432"
	@echo "MinIO: http://localhost:9000 (console: http://localhost:9001)"
	@echo "Prometheus: http://localhost:9091"
	@echo "Grafana: http://localhost:3001 (admin/admin)"

dev-run: dev run ## Start dev environment and run server

format: ## Format Go code
	go fmt ./...
	goimports -w .

lint: ## Run linter
	golangci-lint run

.DEFAULT_GOAL := help
