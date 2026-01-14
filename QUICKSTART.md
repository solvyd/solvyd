# Solvyd Quick Start - Testing Current Implementation

This guide helps you test the **currently implemented** features of Solvyd.

## Prerequisites

- Docker Desktop with Kubernetes enabled
- kubectl configured
- Skaffold installed (`brew install skaffold`)

## 1. Start Solvyd Infrastructure

```bash
cd solvyd

# Start all services with hot-reload
skaffold dev
```

**What this does**:
- Deploys PostgreSQL database
- Deploys Redis cache
- Deploys MinIO object storage
- Deploys Prometheus monitoring
- Deploys Grafana dashboards
- Builds and deploys API Server
- Builds and deploys Worker Agent
- Builds and deploys Web UI (placeholder)

**Wait for**: All pods to be running (2-3 minutes)

```bash
# Check status
kubectl get pods -n solvyd

# Expected output:
NAME                           READY   STATUS    RESTARTS   AGE
api-server-xxx                 1/1     Running   0          2m
worker-agent-xxx               1/1     Running   0          2m
postgres-xxx                   1/1     Running   0          2m
redis-xxx                      1/1     Running   0          2m
minio-xxx                      1/1     Running   0          2m
prometheus-xxx                 1/1     Running   0          2m
grafana-xxx                    1/1     Running   0          2m
```

---

## 2. Verify Worker Registration

```bash
# Watch worker logs
kubectl logs -f deployment/worker-agent -n solvyd

# Look for:
# ✅ "Worker registered successfully: {worker_id}"
# ✅ "Heartbeat sent" (every 30 seconds)
# ✅ "Checking for builds..." (every 5 seconds)
```

**What's happening**:
1. Worker agent starts
2. Calls `POST /api/v1/workers/register`
3. Receives worker UUID
4. Starts heartbeat loop (30s interval)
5. Starts build polling loop (5s interval)

---

## 3. Check API Server

```bash
# Watch API server logs
kubectl logs -f deployment/api-server -n solvyd

# Look for:
# ✅ "Worker registered: {worker_id}"
# ✅ "Heartbeat received from worker: {worker_id}"
```

---

## 4. Create Test Job

First, get the worker ID:

```bash
# Get worker ID from logs
WORKER_ID=$(kubectl logs deployment/worker-agent -n solvyd | grep "Worker registered successfully" | awk '{print $NF}')
echo "Worker ID: $WORKER_ID"
```

Connect to PostgreSQL:

```bash
kubectl exec -it deployment/postgres -n solvyd -- psql -U solvyd
```

Create a test job:

```sql
-- Create test job
INSERT INTO jobs (
  name, description, scm_type, scm_url, scm_branch,
  build_config, enabled, timeout_minutes
)
VALUES (
  'hello-world-test',
  'Simple hello world test build',
  'git',
  'https://github.com/octocat/Hello-World.git',
  'master',
  '{
    "image": "ubuntu:22.04",
    "commands": [
      "echo Hello from Solvyd!",
      "echo Current directory:",
      "pwd",
      "echo Files:",
      "ls -la",
      "echo Build completed!"
    ]
  }'::jsonb,
  true,
  10
)
RETURNING id;
```

**Copy the returned `id` (job UUID)**

---

## 5. Trigger Test Build

Use the job ID from previous step:

```sql
-- Replace <JOB_ID> and <WORKER_ID> with actual values
INSERT INTO builds (
  job_id,
  build_number,
  status,
  worker_id,
  scm_commit_sha,
  branch,
  triggered_by,
  build_config
)
VALUES (
  '<JOB_ID>',
  1,
  'queued',
  '<WORKER_ID>',
  'master',
  'master',
  'manual:test',
  '{
    "image": "ubuntu:22.04",
    "commands": [
      "echo Hello from Solvyd!",
      "echo Current directory:",
      "pwd",
      "echo Files:",
      "ls -la",
      "echo Build completed!"
    ]
  }'::jsonb
)
RETURNING id;
```

**Copy the returned build `id`**

Exit psql: `\q`

---

## 6. Watch Build Execution

```bash
# Watch worker logs in real-time
kubectl logs -f deployment/worker-agent -n solvyd
```

**You should see**:

```
[INFO] Found pending builds: 1
[INFO] Starting build execution: {build_id}
[INFO] Cloning repository: https://github.com/octocat/Hello-World.git
[INFO] Repository cloned successfully
[INFO] Using Docker image: ubuntu:22.04
[INFO] Running: docker run --rm ...
Hello from Solvyd!
Current directory:
/workspace
Files:
total 8
drwxr-xr-x 3 root root 4096 Jan 15 10:30 .
drwxr-xr-x 1 root root 4096 Jan 15 10:30 ..
drwxr-xr-x 8 root root  256 Jan 15 10:30 .git
[INFO] Build completed successfully
[INFO] Build status updated
```

---

## 7. Verify Build Status

```bash
kubectl exec -it deployment/postgres -n solvyd -- psql -U solvyd

-- Check build status
SELECT 
  id, 
  build_number, 
  status, 
  exit_code,
  started_at, 
  completed_at,
  duration_seconds
FROM builds
ORDER BY queued_at DESC
LIMIT 1;

-- Expected:
--   status = 'success'
--   exit_code = 0
--   started_at and completed_at populated
--   duration_seconds = a few seconds
```

---

## 8. Test Different Build Scenarios

### Test 8.1: Failing Build

```sql
INSERT INTO builds (
  job_id, build_number, status, worker_id,
  scm_commit_sha, branch, triggered_by, build_config
)
VALUES (
  '<JOB_ID>',
  2,
  'queued',
  '<WORKER_ID>',
  'master',
  'master',
  'manual:test-failure',
  '{
    "image": "ubuntu:22.04",
    "commands": [
      "echo This build will fail",
      "false",
      "echo This should not print"
    ]
  }'::jsonb
);
```

**Expected**: status='failure', exit_code=1

### Test 8.2: Go Project Build

```sql
INSERT INTO builds (
  job_id, build_number, status, worker_id,
  scm_commit_sha, branch, triggered_by, build_config
)
VALUES (
  '<JOB_ID>',
  3,
  'queued',
  '<WORKER_ID>',
  'master',
  'master',
  'manual:test-go',
  '{
    "image": "golang:1.24",
    "commands": [
      "go version",
      "echo package main > main.go",
      "echo func main() { println(\"Hello Go!\") } >> main.go",
      "go run main.go"
    ]
  }'::jsonb
);
```

**Expected**: status='success', "Hello Go!" in logs

### Test 8.3: Python Project Build

```sql
INSERT INTO builds (
  job_id, build_number, status, worker_id,
  scm_commit_sha, branch, triggered_by, build_config
)
VALUES (
  '<JOB_ID>',
  4,
  'queued',
  '<WORKER_ID>',
  'master',
  'master',
  'manual:test-python',
  '{
    "image": "python:3.11",
    "commands": [
      "python --version",
      "echo print(\"Hello Python!\") > test.py",
      "python test.py"
    ]
  }'::jsonb
);
```

**Expected**: status='success', "Hello Python!" in logs

---

## 9. Monitor with Prometheus

```bash
# Port-forward Prometheus
kubectl port-forward -n solvyd svc/prometheus 9090:9090
```

Open browser: http://localhost:9090

**Query Examples**:
- `up{namespace="solvyd"}` - Check all services are up
- `go_goroutines{job="api-server"}` - API server goroutines
- `process_cpu_seconds_total` - CPU usage

---

## 10. Monitor with Grafana

```bash
# Port-forward Grafana
kubectl port-forward -n solvyd svc/grafana 3000:3000
```

Open browser: http://localhost:3000

**Login**: admin / admin (change on first login)

**Explore**:
- Add Prometheus data source: http://prometheus:9090
- Import dashboards or create custom ones

---

## 11. Direct API Testing

```bash
# Port-forward API server
kubectl port-forward -n solvyd svc/api-server 8080:8080
```

### List Workers
```bash
curl http://localhost:8080/api/v1/workers | jq
```

### Get Worker Details
```bash
curl http://localhost:8080/api/v1/workers/<WORKER_ID> | jq
```

### List Builds
```bash
curl http://localhost:8080/api/v1/builds | jq
```

### Get Build Details
```bash
curl http://localhost:8080/api/v1/builds/<BUILD_ID> | jq
```

---

## Troubleshooting

### Worker Not Registering

```bash
# Check worker logs
kubectl logs deployment/worker-agent -n solvyd

# Check API server connectivity
kubectl exec -it deployment/worker-agent -n solvyd -- wget -O- http://api-server:8080/health
```

### Build Not Starting

```bash
# Check if build is in database
kubectl exec -it deployment/postgres -n solvyd -- psql -U solvyd -c "SELECT * FROM builds WHERE status='queued';"

# Check worker ID matches
echo "Worker ID: $WORKER_ID"

# Verify worker can poll builds
kubectl logs deployment/worker-agent -n solvyd | grep "Checking for builds"
```

### Docker Errors in Worker

```bash
# Check if Docker socket is mounted
kubectl exec -it deployment/worker-agent -n solvyd -- docker ps

# If Docker not available, worker needs Docker-in-Docker or host Docker socket
```

**Current Limitation**: Worker agent in Kubernetes needs Docker access. Options:
1. Use Docker-in-Docker sidecar
2. Mount host Docker socket (security risk)
3. Use Kubernetes Jobs instead of Docker (future)

---

## What's NOT Working Yet

❌ Manual job trigger endpoint (POST /api/v1/jobs/{id}/trigger)
❌ Log upload to API server
❌ Artifact upload to MinIO
❌ Web UI build viewer
❌ Build cancellation
❌ Webhook triggers
❌ Authentication
❌ Real-time WebSocket logs

See [IMPLEMENTATION_PROGRESS.md](./IMPLEMENTATION_PROGRESS.md) for roadmap.

---

## Clean Up

```bash
# Stop Skaffold (Ctrl+C in terminal)

# Delete namespace
kubectl delete namespace solvyd

# Or delete all resources
cd solvyd
skaffold delete
```

---

## Next Steps

1. **Implement Log Upload** - Workers upload logs to API
2. **Build Trigger Endpoint** - Manual job triggering via API
3. **Basic Web UI** - View jobs and builds
4. **Worker Assignment** - Automatic build assignment scheduler

**ETA for working demo**: 3-5 days

---

## Need Help?

- Check logs: `kubectl logs -f deployment/<service> -n solvyd`
- Describe pod: `kubectl describe pod/<pod-name> -n solvyd`
- Check events: `kubectl get events -n solvyd --sort-by='.lastTimestamp'`
- Database access: `kubectl exec -it deployment/postgres -n solvyd -- psql -U solvyd`

---

**Last Updated**: Current session  
**Tested On**: macOS, Docker Desktop, Kubernetes 1.28
