.PHONY: help dev clean stop logs

help: ## Show this help message
	@echo 'Ritmo CI/CD Platform - Quick Start'
	@echo ''
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

dev: ## Start complete development environment
	@echo "🚀 Starting Ritmo development environment..."
	@docker-compose up -d
	@echo "⏳ Waiting for services to be ready..."
	@sleep 10
	@echo "📊 Initializing database..."
	@PGPASSWORD=ritmo_dev_password psql -h localhost -U ritmo -d ritmo -f database/schema.sql 2>/dev/null || true
	@PGPASSWORD=ritmo_dev_password psql -h localhost -U ritmo -d ritmo -f database/seed.sql 2>/dev/null || true
	@echo ""
	@echo "✅ Development environment ready!"
	@echo ""
	@echo "Services:"
	@echo "  📊 PostgreSQL:     localhost:5432"
	@echo "  🗄️  MinIO:          http://localhost:9000 (console: http://localhost:9001)"
	@echo "  📈 Prometheus:     http://localhost:9091"
	@echo "  📉 Grafana:        http://localhost:3001 (admin/admin)"
	@echo ""
	@echo "Next steps:"
	@echo "  1. Start API server:    cd api-server && make run"
	@echo "  2. Start worker agent:  cd worker-agent && go run cmd/agent/main.go"
	@echo "  3. Start web UI:        cd web-ui && npm run dev"
	@echo ""

stop: ## Stop all services
	@echo "🛑 Stopping all services..."
	@docker-compose down
	@echo "✅ All services stopped"

clean: ## Stop services and remove volumes
	@echo "🗑️  Cleaning up..."
	@docker-compose down -v
	@echo "✅ Cleanup complete"

logs: ## Show logs from all services
	@docker-compose logs -f

status: ## Show status of all services
	@docker-compose ps

.DEFAULT_GOAL := help
