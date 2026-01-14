# Solvyd - Next-Generation CI/CD Platform

Solvyd is a highly scalable, plugin-based CI/CD platform designed to overcome the limitations of traditional single-controller systems like Jenkins.

🚧 **Status**: Active Development - Core build execution pipeline implemented  
📊 **Progress**: ~25% complete - See [IMPLEMENTATION_PROGRESS.md](./IMPLEMENTATION_PROGRESS.md)  
🎯 **Target**: MVP in 2 weeks, Production launch in 6 weeks

## 🎉 What's Working Now

✅ **Worker Registration & Heartbeat** - Workers self-register and maintain heartbeat  
✅ **Docker-based Build Execution** - Full Docker container execution with git cloning  
✅ **Build Polling & Assignment** - Workers fetch and execute assigned builds  
✅ **Status Reporting** - Real-time build status updates to API server  
✅ **Kubernetes Deployment** - Complete k8s manifests with Skaffold hot-reload  
✅ **Comprehensive Documentation** - Full docs at [solvyd-docs](../solvyd-docs)

## 📚 Quick Links

- **[Implementation Progress](./IMPLEMENTATION_PROGRESS.md)** - Detailed status and next steps
- **[Quick Start Guide](./QUICKSTART.md)** - Testing current implementation
- **[Full Documentation](../solvyd-docs/docs/)** - Architecture, guides, deployment, API reference
- **[Getting Started Guide](../solvyd-docs/docs/getting-started/)** - Setup instructions

## Architecture

### Core Components

1. **API Server** (`/api-server`)
   - Horizontally scalable stateless servers
   - gRPC and REST endpoints
   - Job scheduling and orchestration
   - Worker lifecycle management
   - Binary-level plugin system

2. **Database** (`/database`)
   - PostgreSQL for reliable, scalable storage
   - Stores jobs, builds, workers, artifacts, deployments
   - Supports multiple API server instances

3. **Worker Agent** (`/worker-agent`)
   - Lightweight agent running on worker nodes
   - Executes jobs in isolated environments
   - Reports real-time status to API server
   - On-demand scaling capability

4. **Web UI** (`/web-ui`)
   - Modern React-based dashboard
   - Real-time job monitoring
   - Build analytics and metrics
   - Worker fleet management
   - Job configuration interface

5. **Plugin SDK** (`/plugin-sdk`)
   - Binary plugin interface
   - SCM integrations (Git, GitHub, GitLab)
   - Build tool plugins (Maven, Gradle, npm, etc.)
   - Artifact storage (S3, Artifactory, etc.)
   - Notification plugins (Slack, email, etc.)
   - Deployment target plugins

6. **CLI** (`/cli`)
   - Command-line interface for job management
   - Pipeline definition and triggering
   - Configuration management

## Key Features

### Scalability
- **Horizontal API server scaling**: Run unlimited API server instances
- **Shared database backend**: All servers use same PostgreSQL cluster
- **On-demand worker provisioning**: Scale workers based on job queue
- **Distributed job execution**: No single point of failure

### CI/CD Separation
- **Distinct CI and CD phases**: Clear separation of concerns
- **Artifact promotion**: Seamless handover between CI and CD
- **External CD integration**: Native support for GitOps tools (ArgoCD, Flux)
- **Artifact versioning**: Track and promote artifacts across environments

### GitOps Configuration
- **Declarative configuration**: Define jobs, credentials, webhooks in Git
- **Automatic synchronization**: Continuous sync from Git repository
- **Version control**: Full audit trail of configuration changes
- **Pull Request workflow**: Review and approve configuration changes
- **Encrypted secrets**: Secure credential management with external secret managers

### Plugin Architecture
- **Binary plugins**: Native performance, no JVM overhead
- **Hot-reloadable**: Update plugins without server restart
- **Sandboxed execution**: Isolated plugin runtime for security
- **Plugin marketplace**: Extensible ecosystem

### Observability
- **Metrics**: Prometheus-compatible metrics
- **Distributed tracing**: OpenTelemetry integration
- **Centralized logging**: Structured logging with correlation IDs
- **Real-time dashboards**: Live build status and system health

## Quick Start

### Prerequisites

- **Kubernetes cluster** (minikube, kind, or Docker Desktop with Kubernetes enabled)
- **kubectl** installed and configured
- **Docker** for building images
- **Skaffold** (optional, for hot-reload development)

### Option 1: Quick Setup (Recommended)

```bash
# Clone repository
git clone https://github.com/solvyd/solvyd.git
cd solvyd

# Setup complete environment on Kubernetes
make dev
```

This will:
- Create the `solvyd` namespace
- Deploy PostgreSQL, Redis, MinIO, Prometheus, and Grafana
- Build and deploy API server, worker agents, and web UI
- Initialize the database with schema and seed data
- Expose services via NodePort

### Option 2: Development with Hot-Reload (Skaffold)

```bash
# Install Skaffold if not already installed
# macOS: brew install skaffold
# Linux: curl -Lo skaffold https://storage.googleapis.com/skaffold/releases/latest/skaffold-linux-amd64 && chmod +x skaffold && sudo mv skaffold /usr/local/bin

# Start development mode with auto-rebuild on file changes
make dev-skaffold
```

### Option 3: Manual Kubernetes Deployment

```bash
# Build Docker images
docker build -t solvyd/api-server:latest ./api-server
docker build -t solvyd/worker-agent:latest ./worker-agent
docker build -t solvyd/web-ui:latest ./web-ui

# Load images to cluster (for minikube)
minikube image load solvyd/api-server:latest
minikube image load solvyd/worker-agent:latest
minikube image load solvyd/web-ui:latest

# Deploy to Kubernetes
make k8s-deploy
```

Access at:
- **Web UI**: http://localhost:30000
- **API**: http://localhost:30080
- **Grafana**: `kubectl port-forward -n solvyd svc/grafana 3001:3000` then http://localhost:3001
- **MinIO Console**: `kubectl port-forward -n solvyd svc/minio 9001:9001` then http://localhost:9001

### Useful Commands

```bash
# View logs
make logs              # API server logs
make logs-worker       # Worker agent logs

# Check status
make status            # Show all pod statuses
kubectl get all -n solvyd

# Port forward services
kubectl port-forward -n solvyd svc/grafana 3001:3000
kubectl port-forward -n solvyd svc/minio 9001:9001

# Clean up
make clean             # Delete all resources
```

## 📖 Documentation

**Complete documentation available at**: **https://solvyd.github.io**

- [Quick Start Guide](https://solvyd.github.io/getting-started/quickstart/)
- [Installation Guide](https://solvyd.github.io/getting-started/installation/)
- [Kubernetes Setup Guide](https://solvyd.github.io/deployment/kubernetes-setup/)
- [Kubernetes Quick Reference](https://solvyd.github.io/deployment/kubernetes-quickref/)
- [Architecture Overview](https://solvyd.github.io/architecture/overview/)
- [Plugin Development](https://solvyd.github.io/plugins/introduction/)
- [Enterprise Security](https://solvyd.github.io/security/overview/)
- [API Reference](https://solvyd.github.io/api-reference/rest-api/)

## Project Status

🚧 **Active Development** - Building foundation components

## Technology Stack

- **Backend**: Go 1.21+
- **Database**: PostgreSQL 15+
- **Frontend**: React 18 + TypeScript + Vite
- **Communication**: gRPC + Protocol Buffers
- **Containerization**: Docker
- **Orchestration**: Kubernetes
- **Development Tools**: Skaffold, Kustomize

## License

MIT License
