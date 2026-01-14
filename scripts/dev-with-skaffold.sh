#!/bin/bash
set -e

echo "🚀 Starting Solvyd development with Skaffold..."

# Check if skaffold is installed
if ! command -v skaffold &> /dev/null; then
    echo "❌ Skaffold is not installed."
    echo "📥 Install it with:"
    echo "   macOS: brew install skaffold"
    echo "   Linux: curl -Lo skaffold https://storage.googleapis.com/skaffold/releases/latest/skaffold-linux-amd64 && chmod +x skaffold && sudo mv skaffold /usr/local/bin"
    exit 1
fi

# Create namespace if it doesn't exist
kubectl get namespace solvyd &> /dev/null || kubectl create namespace solvyd

# Deploy infrastructure first
echo "🔧 Deploying infrastructure..."
kubectl apply -f k8s/postgres.yaml
kubectl apply -f k8s/redis.yaml
kubectl apply -f k8s/minio.yaml
kubectl apply -f k8s/prometheus.yaml
kubectl apply -f k8s/grafana.yaml

echo "⏳ Waiting for infrastructure..."
kubectl wait --for=condition=ready pod -l app=postgres -n solvyd --timeout=120s
kubectl wait --for=condition=ready pod -l app=redis -n solvyd --timeout=60s

# Initialize database if needed
kubectl exec -n solvyd deployment/postgres -- psql -U solvyd -d solvyd -c "SELECT 1" &> /dev/null || {
    echo "🗄️  Initializing database..."
    kubectl exec -n solvyd deployment/postgres -i -- psql -U solvyd -d solvyd < database/schema.sql
    kubectl exec -n solvyd deployment/postgres -i -- psql -U solvyd -d solvyd < database/seed.sql
}

# Start Skaffold in dev mode (auto-rebuild and deploy on changes)
echo "🔥 Starting Skaffold dev mode..."
echo "   - Watching for file changes"
echo "   - Auto-rebuilding and redeploying"
echo "   - Port-forwarding services"
echo ""
skaffold dev --port-forward=true
