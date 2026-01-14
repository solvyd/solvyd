#!/bin/bash
set -e

echo "🚀 Setting up Solvyd on local Kubernetes..."

# Check if kubectl is installed
if ! command -v kubectl &> /dev/null; then
    echo "❌ kubectl is not installed. Please install kubectl first."
    exit 1
fi

# Check if a Kubernetes cluster is running
if ! kubectl cluster-info &> /dev/null; then
    echo "❌ No Kubernetes cluster detected. Please start minikube, kind, or Docker Desktop Kubernetes."
    exit 1
fi

# Create namespace
echo "📦 Creating namespace..."
kubectl apply -f k8s/namespace.yaml

# Deploy infrastructure
echo "🔧 Deploying infrastructure..."
kubectl apply -f k8s/postgres.yaml
kubectl apply -f k8s/redis.yaml
kubectl apply -f k8s/minio.yaml
kubectl apply -f k8s/prometheus.yaml
kubectl apply -f k8s/grafana.yaml

# Wait for infrastructure to be ready
echo "⏳ Waiting for infrastructure pods to be ready..."
kubectl wait --for=condition=ready pod -l app=postgres -n solvyd --timeout=120s
kubectl wait --for=condition=ready pod -l app=redis -n solvyd --timeout=60s
kubectl wait --for=condition=ready pod -l app=minio -n solvyd --timeout=60s

# Initialize database
echo "🗄️  Initializing database..."
kubectl exec -n solvyd deployment/postgres -- psql -U solvyd -d solvyd -c "SELECT 1" &> /dev/null || {
    echo "Setting up database schema..."
    kubectl exec -n solvyd deployment/postgres -i -- psql -U solvyd -d solvyd < database/schema.sql
    kubectl exec -n solvyd deployment/postgres -i -- psql -U solvyd -d solvyd < database/seed.sql
}

# Build and deploy applications
echo "🔨 Building application images..."
if command -v skaffold &> /dev/null; then
    echo "Using Skaffold for build and deploy..."
    skaffold dev --port-forward=true
else
    echo "Skaffold not found. Building with docker..."
    
    # Build images
    echo "Building API server..."
    docker build -t solvyd/api-server:latest ./api-server
    
    echo "Building worker agent..."
    docker build -t solvyd/worker-agent:latest ./worker-agent
    
    echo "Building web UI..."
    docker build -t solvyd/web-ui:latest ./web-ui
    
    # Load images to cluster (for minikube/kind)
    if command -v minikube &> /dev/null && minikube status &> /dev/null; then
        echo "Loading images to minikube..."
        minikube image load solvyd/api-server:latest
        minikube image load solvyd/worker-agent:latest
        minikube image load solvyd/web-ui:latest
    elif command -v kind &> /dev/null; then
        CLUSTER_NAME=$(kubectl config current-context | sed 's/kind-//')
        echo "Loading images to kind cluster: $CLUSTER_NAME"
        kind load docker-image solvyd/api-server:latest --name $CLUSTER_NAME
        kind load docker-image solvyd/worker-agent:latest --name $CLUSTER_NAME
        kind load docker-image solvyd/web-ui:latest --name $CLUSTER_NAME
    fi
    
    # Deploy applications
    echo "Deploying applications..."
    kubectl apply -f k8s/api-server.yaml
    kubectl apply -f k8s/worker-agent.yaml
    kubectl apply -f k8s/web-ui.yaml
    
    echo "⏳ Waiting for application pods to be ready..."
    kubectl wait --for=condition=ready pod -l app=api-server -n solvyd --timeout=120s
    kubectl wait --for=condition=ready pod -l app=worker-agent -n solvyd --timeout=120s
    kubectl wait --for=condition=ready pod -l app=web-ui -n solvyd --timeout=120s
    
    echo ""
    echo "✅ Solvyd is ready!"
    echo ""
    echo "🌐 Access the services:"
    echo "   Web UI:     http://localhost:30000"
    echo "   API Server: http://localhost:30080"
    echo "   Grafana:    kubectl port-forward -n solvyd svc/grafana 3001:3000"
    echo "   MinIO:      kubectl port-forward -n solvyd svc/minio 9001:9001"
    echo ""
    echo "📊 View logs:"
    echo "   kubectl logs -f -n solvyd deployment/api-server"
    echo "   kubectl logs -f -n solvyd deployment/worker-agent"
    echo ""
    echo "🧹 To clean up:"
    echo "   kubectl delete namespace solvyd"
fi
