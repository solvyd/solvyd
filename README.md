# Solvyd - Next-Generation CI/CD Platform

Solvyd is a highly scalable, plugin-based CI/CD platform designed to overcome the limitations of traditional single-controller systems like Jenkins.

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

```bash
# Clone repository
git clone https://github.com/solvyd/solvyd.git
cd solvyd

# Start the infrastructure
docker-compose up -d

# Start API server
cd api-server
go run cmd/server/main.go

# Start worker agent (new terminal)
cd worker-agent
go run cmd/agent/main.go --api-server=localhost:8080

# Start web UI (new terminal)
cd web-ui
npm install
npm run dev
```

Access at:
- **Web UI**: http://localhost:3000
- **API**: http://localhost:8080
- **Grafana**: http://localhost:3001

## 📖 Documentation

**Complete documentation available at**: **https://solvyd.github.io**

- [Quick Start Guide](https://solvyd.github.io/getting-started/quickstart/)
- [Installation Guide](https://solvyd.github.io/getting-started/installation/)
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
- **Orchestration**: Kubernetes (optional)

## License

MIT License
