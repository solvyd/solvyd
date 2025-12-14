# Ritmo Worker Agent

The worker agent runs on worker nodes and executes CI/CD builds.

## Features

- Registers with API server
- Receives and executes build jobs
- Multiple isolation strategies (Docker, process, VM)
- Real-time log streaming
- Artifact upload
- Health monitoring and heartbeat
- Graceful shutdown with build draining

## Quick Start

```bash
# Run with default settings (Docker isolation)
go run cmd/agent/main.go --api-server=localhost:8080

# Run with custom settings
go run cmd/agent/main.go \
  --api-server=localhost:8080 \
  --name=worker-01 \
  --max-concurrent=4 \
  --label=zone=us-west-1a \
  --label=type=linux \
  --isolation=docker
```

## Configuration

Command-line flags:
- `--api-server`: API server address (default: localhost:8080)
- `--name`: Worker name (default: auto-generated)
- `--max-concurrent`: Maximum concurrent builds (default: 2)
- `--label`: Worker labels for job targeting (can be repeated)
- `--log-level`: Log level (debug, info, warn, error)
- `--isolation`: Build isolation type (docker, process, vm)

## Build Isolation

### Docker (recommended)
- Runs builds in Docker containers
- Maximum isolation and security
- Requires Docker daemon

### Process
- Runs builds as separate processes
- Lightweight, minimal overhead
- Less isolation than Docker

### VM (future)
- Runs builds in virtual machines
- Maximum security for untrusted code
- Higher resource usage

## Architecture

```
cmd/
  agent/
    main.go          # Agent entry point

internal/
  agent/             # Agent core logic
  config/            # Configuration
  executor/          # Build execution engines
```

## Development

```bash
# Build
go build -o bin/ritmo-agent cmd/agent/main.go

# Run
./bin/ritmo-agent --api-server=localhost:8080
```

## Next Steps

- [ ] Implement Docker executor with actual Docker API
- [ ] Implement process executor with proper isolation
- [ ] Add artifact upload to S3/MinIO
- [ ] Stream logs to API server in real-time
- [ ] Implement build cancellation
- [ ] Add support for custom Docker images per job
- [ ] Implement secrets injection
- [ ] Add resource limits (CPU, memory, disk)
