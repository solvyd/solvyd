# Solvyd API Server

The core API server for Solvyd CI/CD platform.

## Features

- RESTful API for job, build, and worker management
- Job scheduling and worker assignment
- Real-time build status via WebSocket
- Prometheus metrics integration
- PostgreSQL database backend
- Plugin system for extensibility
- CI/CD separation with artifact promotion
- Deployment orchestration

## Quick Start

### Prerequisites

- Go 1.21+
- PostgreSQL 15+
- Docker & Docker Compose (for infrastructure)

### Development Setup

1. **Start infrastructure**:
   ```bash
   make docker-up
   ```

2. **Initialize database**:
   ```bash
   make db-migrate
   make db-seed
   ```

3. **Run the server**:
   ```bash
   make run
   ```

Or run everything in one command:
```bash
make dev-run
```

The API server will be available at `http://localhost:8080`.

## API Endpoints

### Health & Status
- `GET /health` - Health check
- `GET /ready` - Readiness check
- `GET /metrics` - Prometheus metrics

### Jobs
- `GET /api/v1/jobs` - List all jobs
- `POST /api/v1/jobs` - Create a new job
- `GET /api/v1/jobs/{id}` - Get job details
- `PUT /api/v1/jobs/{id}` - Update a job
- `DELETE /api/v1/jobs/{id}` - Delete a job
- `POST /api/v1/jobs/{id}/trigger` - Trigger a manual build

### Builds
- `GET /api/v1/builds` - List all builds
- `GET /api/v1/builds/{id}` - Get build details
- `POST /api/v1/builds/{id}/cancel` - Cancel a build
- `GET /api/v1/builds/{id}/logs` - Get build logs
- `GET /api/v1/builds/{id}/artifacts` - List build artifacts

### Workers
- `GET /api/v1/workers` - List all workers
- `GET /api/v1/workers/{id}` - Get worker details
- `PUT /api/v1/workers/{id}` - Update worker configuration
- `POST /api/v1/workers/{id}/drain` - Drain a worker

### Deployments
- `GET /api/v1/deployments` - List deployments
- `POST /api/v1/deployments` - Create a deployment
- `GET /api/v1/deployments/{id}` - Get deployment details
- `POST /api/v1/deployments/{id}/rollback` - Rollback a deployment

### Plugins
- `GET /api/v1/plugins` - List installed plugins
- `GET /api/v1/plugins/{id}` - Get plugin details
- `POST /api/v1/plugins` - Install a plugin

### WebSocket
- `GET /ws` - WebSocket connection for real-time updates

## Configuration

Configuration can be provided via:
1. `config.yaml` file
2. Environment variables
3. Command-line flags

See `config.yaml` for all available options.

## Architecture

```
cmd/
  server/
    main.go          # Application entry point

internal/
  config/            # Configuration management
  database/          # Database connection and helpers
  handlers/          # HTTP request handlers
  models/            # Data models
  scheduler/         # Job scheduling logic
  worker/            # Worker management
  metrics/           # Prometheus metrics
  plugin/            # Plugin system (TODO)
```

## Development

```bash
# Install dependencies
make deps

# Format code
make format

# Run tests
make test

# Run with coverage
make test-coverage

# Build binary
make build
```

## Metrics

The server exposes Prometheus metrics at `/metrics`:

- `ritmo_builds_total` - Total builds by status
- `ritmo_builds_queued` - Current queued builds
- `ritmo_builds_running` - Current running builds
- `ritmo_build_duration_seconds` - Build duration histogram
- `ritmo_workers_total` - Workers by status
- `ritmo_worker_utilization` - Worker utilization
- `ritmo_deployments_total` - Total deployments
- `ritmo_api_requests_total` - API request count
- `ritmo_api_request_duration_seconds` - API request duration

## Next Steps

- [ ] Implement plugin system with binary plugin loading
- [ ] Add authentication and authorization (JWT, OAuth)
- [ ] Implement webhook handlers for GitHub, GitLab, Bitbucket
- [ ] Add artifact storage integration (S3, GCS, Artifactory)
- [ ] Implement build log streaming
- [ ] Add support for pipeline stages
- [ ] Create deployment plugins (Kubernetes, Docker, SSH)
- [ ] Add audit logging
- [ ] Implement secrets management (Vault integration)
