# Getting Started with Ritmo

This guide will help you get Ritmo up and running on your local machine.

## Prerequisites

- **Go 1.21+**: For API server and worker agent
- **Node.js 18+** and **npm**: For web UI
- **PostgreSQL 15+**: Database
- **Docker**: For running infrastructure and build isolation
- **Git**: Version control

## Quick Start (5 minutes)

### 1. Clone the Repository

```bash
git clone https://github.com/vrenjith/ritmo.git
cd ritmo
```

### 2. Start Infrastructure

```bash
# Start PostgreSQL, Redis, MinIO, Prometheus, Grafana
docker-compose up -d

# Wait for services to be ready
sleep 10
```

### 3. Initialize Database

```bash
# Create schema
PGPASSWORD=ritmo_dev_password psql -h localhost -U ritmo -d ritmo -f database/schema.sql

# Seed with sample data
PGPASSWORD=ritmo_dev_password psql -h localhost -U ritmo -d ritmo -f database/seed.sql
```

### 4. Start API Server

```bash
cd api-server
go mod download
go run cmd/server/main.go
```

The API server will start on http://localhost:8080

### 5. Start Worker Agent (in a new terminal)

```bash
cd worker-agent
go run cmd/agent/main.go --api-server=localhost:8080 --name=worker-local
```

### 6. Start Web UI (in a new terminal)

```bash
cd web-ui
npm install
npm run dev
```

The web UI will start on http://localhost:3000

### 7. Access Ritmo

Open your browser and navigate to:
- **Web UI**: http://localhost:3000
- **API Server**: http://localhost:8080/api/v1/jobs
- **Metrics**: http://localhost:8080/metrics
- **Grafana**: http://localhost:3001 (admin/admin)
- **MinIO Console**: http://localhost:9001 (ritmo/ritmo_minio_password)

## What's Next?

### Create Your First Job

1. Go to http://localhost:3000/jobs
2. Click "New Job"
3. Configure:
   - Name: `hello-world`
   - SCM URL: `https://github.com/your-repo/hello-world`
   - Branch: `main`
   - Build Config: Maven, npm, or custom script

### Trigger a Build

1. Go to the Jobs page
2. Click "Trigger" on your job
3. Watch it run in real-time on the Builds page

### View Metrics

1. Go to http://localhost:3001 (Grafana)
2. Login with admin/admin
3. Explore build metrics, worker utilization, and system health

## Configuration

### API Server

Configuration is in `api-server/config.yaml` or via environment variables:

```yaml
port: 8080
database_url: "postgres://ritmo:ritmo_dev_password@localhost:5432/ritmo"
log_level: "info"
```

### Worker Agent

Command-line flags:

```bash
./ritmo-agent \
  --api-server=localhost:8080 \
  --name=worker-01 \
  --max-concurrent=4 \
  --isolation=docker
```

## Troubleshooting

### Database Connection Issues

```bash
# Check PostgreSQL is running
docker ps | grep postgres

# Test connection
PGPASSWORD=ritmo_dev_password psql -h localhost -U ritmo -d ritmo -c "SELECT 1;"
```

### Worker Not Registering

```bash
# Check API server logs
cd api-server
tail -f logs/ritmo.log

# Check worker agent logs
cd worker-agent
./ritmo-agent --log-level=debug
```

### API Server Won't Start

```bash
# Check if port 8080 is available
lsof -i :8080

# Run with debug logging
cd api-server
go run cmd/server/main.go --log-level=debug
```

## Development Workflow

### Running Tests

```bash
# API Server tests
cd api-server
make test

# Web UI tests
cd web-ui
npm test
```

### Building for Production

```bash
# Build API server
cd api-server
make build

# Build worker agent
cd worker-agent
go build -o bin/ritmo-agent cmd/agent/main.go

# Build web UI
cd web-ui
npm run build
```

### Using Make Commands

API Server:
```bash
make help          # Show all available commands
make dev           # Start development environment
make dev-run       # Start dev environment and run server
make test          # Run tests
make docker-up     # Start infrastructure
make db-migrate    # Run database migrations
```

## Next Steps

- Read the [Architecture Guide](ARCHITECTURE.md)
- Explore [Plugin Development](plugin-sdk/README.md)
- Check out [Example Jobs](docs/examples/)
- Learn about [CI/CD Separation](docs/ci-cd-separation.md)
