.PHONY: help dev dev-skaffold clean stop logs k8s-deploy k8s-delete

help: ## Show this help message
	@echo 'Solvyd CI/CD Platform - Quick Start'
	@echo ''
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

dev: ## Start complete development environment on Kubernetes
	@./scripts/setup-local-k8s.sh

dev-skaffold: ## Start development with Skaffold (hot-reload)
	@./scripts/dev-with-skaffold.sh

k8s-deploy: ## Deploy to Kubernetes cluster
	@echo "🚀 Deploying Solvyd to Kubernetes..."
	@kubectl apply -k k8s/
	@echo "✅ Deployment complete!"
	@echo ""
	@echo "🌐 Access services:"
	@echo "  Web UI:     http://localhost:30000"
	@echo "  API Server: http://localhost:30080"
	@echo ""
	@echo "📊 Port-forward additional services:"
	@echo "  Grafana:    kubectl port-forward -n solvyd svc/grafana 3001:3000"
	@echo "  MinIO:      kubectl port-forward -n solvyd svc/minio 9001:9001"

k8s-delete: ## Delete Kubernetes deployment
	@echo "🗑️  Deleting Solvyd from Kubernetes..."
	@kubectl delete namespace solvyd
	@echo "✅ Deletion complete"

stop: ## Stop all services (alias for k8s-delete)
	@make k8s-delete

clean: ## Clean up Kubernetes resources and volumes
	@echo "🗑️  Cleaning up Kubernetes resources..."
	@kubectl delete namespace solvyd --ignore-not-found=true
	@echo "✅ Cleanup complete"

logs: ## Show logs from API server
	@kubectl logs -f -n solvyd deployment/api-server

logs-worker: ## Show logs from worker agents
	@kubectl logs -f -n solvyd deployment/worker-agent

status: ## Show status of all pods
	@kubectl get pods -n solvyd

# Enterprise targets
include Makefile.enterprise

.DEFAULT_GOAL := help
