# CI/CD Separation in Ritmo

Ritmo provides first-class support for separating Continuous Integration (CI) and Continuous Deployment (CD) phases. This allows you to run independent instances for build/test and deployment, with seamless artifact handover.

## Why Separate CI and CD?

### Benefits

1. **Security**: Deployment credentials never touch CI systems
2. **Scalability**: Scale CI and CD independently based on demand
3. **Compliance**: Meet regulatory requirements for separated environments
4. **Flexibility**: Use different tools for different phases
5. **Cost**: Optimize resource allocation per phase

### Traditional Approach (Single Pipeline)

```
[Code Push] → [Build & Test] → [Deploy to Staging] → [Deploy to Production]
     ↓             ↓                  ↓                       ↓
   CI Server   CI Server         CI Server              CI Server
```

**Problems**: Single point of failure, mixed credentials, hard to scale

### Ritmo Approach (Separated)

```
[Code Push] → [Build & Test] → [Artifact Storage]
     ↓             ↓                  ↓
   CI Instance  CI Instance      Artifact Repository
                                       ↓
                              [Artifact Promotion]
                                       ↓
                           [Deploy to Staging] → [Deploy to Production]
                                  ↓                      ↓
                              CD Instance            CD Instance
```

## Architecture

### CI Phase (Ritmo CI Instance)

**Responsibilities:**
- Source code checkout
- Dependency resolution
- Build compilation
- Unit tests, integration tests
- Static code analysis
- Security scanning
- Artifact creation and storage
- Artifact metadata tagging

**Configuration:**

```yaml
# CI Instance Configuration
job:
  name: backend-service-ci
  scm:
    type: github
    url: https://github.com/myorg/backend-service
    branch: main
  
  build:
    type: maven
    goals: [clean, package, verify]
    jdk_version: "17"
  
  tests:
    - unit: mvn test
    - integration: mvn verify
    - coverage: jacoco:report
  
  artifacts:
    - name: backend-service.jar
      path: target/backend-service-*.jar
      storage: s3
      metadata:
        version: ${GIT_TAG}
        commit: ${GIT_COMMIT}
        build_number: ${BUILD_NUMBER}
  
  on_success:
    - tag_artifact:
        environment: dev
        auto_promote: true
    - notify_slack:
        message: "Build successful, ready for deployment"
    - trigger_cd_system:
        system: github_actions
        workflow: deploy-to-dev
        parameters:
          artifact_url: ${ARTIFACT_URL}
          version: ${VERSION}
```

### CD Phase (Ritmo CD Instance or External)

**Responsibilities:**
- Artifact retrieval
- Environment preparation
- Deployment execution
- Smoke tests
- Rollback capability
- Deployment verification

**Configuration:**

```yaml
# CD Instance Configuration
deployment:
  name: backend-service-deployment
  
  source:
    type: artifact_repository
    artifact_id: ${ARTIFACT_ID}
    promotion_level: staging  # Only deploy promoted artifacts
  
  environments:
    - name: staging
      target:
        type: kubernetes
        cluster: staging-cluster
        namespace: backend
      
      pre_deploy:
        - check_cluster_health
        - backup_current_version
      
      deploy:
        - apply_manifests
        - wait_for_rollout
        - run_smoke_tests
      
      post_deploy:
        - notify_slack
        - update_artifact_status
      
      rollback:
        on_failure: automatic
        strategy: previous_version
    
    - name: production
      requires_approval: true
      approvers: [ops-team, platform-team]
      
      target:
        type: kubernetes
        cluster: prod-cluster
        namespace: backend
```

## Artifact Promotion Workflow

### Stages

```
Dev → Staging → Production
 ↓       ↓          ↓
Auto   Manual    Manual + Approval
```

### 1. CI Builds and Tags Artifact

```bash
# CI phase completes successfully
Build ID: abc-123
Artifact: backend-service-1.2.3.jar
Initial Tag: dev
```

```sql
-- Database record
INSERT INTO artifacts (
  build_id, name, promotion_status, created_at
) VALUES (
  'abc-123', 
  'backend-service-1.2.3.jar', 
  'dev',
  NOW()
);
```

### 2. Promote to Staging

```bash
# Manual or automatic promotion
curl -X POST /api/v1/artifacts/{id}/promote \
  -d '{"from": "dev", "to": "staging", "approved_by": "ops-team"}'
```

```sql
UPDATE artifacts 
SET promotion_status = 'staging',
    promoted_at = NOW(),
    promoted_by = 'ops-team'
WHERE id = 'artifact-id';
```

### 3. Deploy to Staging

```bash
# CD system picks up promoted artifact
Artifact Status: staging
Trigger: automatic
Target: staging-cluster
```

### 4. Promote to Production

```bash
# Require approval
curl -X POST /api/v1/artifacts/{id}/promote \
  -d '{
    "from": "staging", 
    "to": "production",
    "approved_by": ["alice@company.com", "bob@company.com"],
    "notes": "Tested successfully in staging"
  }'
```

## Integration Patterns

### Pattern 1: Ritmo CI → Ritmo CD

Both CI and CD use Ritmo instances.

```yaml
# CI Instance
on_success:
  - webhook:
      url: https://cd-instance.company.com/api/v1/deployments
      method: POST
      body:
        artifact_id: ${ARTIFACT_ID}
        artifact_url: ${ARTIFACT_URL}
        environment: staging
        version: ${VERSION}
```

**Benefits**: Unified platform, consistent API, shared metadata

### Pattern 2: Ritmo CI → GitHub Actions CD

CI in Ritmo, deployment via GitHub Actions.

```yaml
# CI Instance
on_success:
  - trigger_github_workflow:
      repository: myorg/backend-service
      workflow: deploy.yml
      inputs:
        artifact_url: ${ARTIFACT_URL}
        version: ${VERSION}
        environment: staging
```

```yaml
# .github/workflows/deploy.yml
name: Deploy
on:
  workflow_dispatch:
    inputs:
      artifact_url:
        required: true
      version:
        required: true
      environment:
        required: true

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - name: Download Artifact
        run: |
          curl -o app.jar ${{ github.event.inputs.artifact_url }}
      
      - name: Deploy to Kubernetes
        run: |
          kubectl set image deployment/backend \
            backend=myregistry/backend:${{ github.event.inputs.version }}
```

**Benefits**: Leverage GitHub's security, compliance, approvals

### Pattern 3: Ritmo CI → ArgoCD

CI in Ritmo, GitOps deployment via ArgoCD.

```yaml
# CI Instance
on_success:
  - update_gitops_repo:
      repository: myorg/gitops-manifests
      path: apps/backend/overlays/staging/kustomization.yaml
      update:
        - key: images[0].newTag
          value: ${VERSION}
      commit_message: "Update backend to ${VERSION}"
      create_pr: true
```

ArgoCD automatically detects the Git change and deploys.

**Benefits**: GitOps best practices, audit trail, declarative config

### Pattern 4: Ritmo CI → Spinnaker

CI in Ritmo, orchestrated deployment via Spinnaker.

```yaml
# CI Instance
on_success:
  - trigger_spinnaker_pipeline:
      application: backend-service
      pipeline: deploy-to-staging
      parameters:
        artifactUrl: ${ARTIFACT_URL}
        version: ${VERSION}
```

**Benefits**: Advanced deployment strategies, canary, blue-green

## API Endpoints for CI/CD Separation

### Artifact Management

```bash
# Create artifact record (CI phase)
POST /api/v1/artifacts
{
  "build_id": "abc-123",
  "name": "backend-service-1.2.3.jar",
  "storage_url": "s3://artifacts/backend-service-1.2.3.jar",
  "checksum": "sha256:abcd1234...",
  "metadata": {
    "version": "1.2.3",
    "commit": "abc123",
    "branch": "main"
  }
}

# Promote artifact
POST /api/v1/artifacts/{id}/promote
{
  "from_environment": "dev",
  "to_environment": "staging",
  "approved_by": "ops-team",
  "notes": "Ready for staging"
}

# Get artifact by promotion level
GET /api/v1/artifacts?promotion_status=staging&name=backend-service

# Get artifact metadata
GET /api/v1/artifacts/{id}
```

### Deployment Triggers

```bash
# Create deployment (CD phase)
POST /api/v1/deployments
{
  "artifact_id": "artifact-uuid",
  "environment": "staging",
  "target_type": "kubernetes",
  "target_url": "https://staging-cluster.company.com",
  "deployed_by": "cd-system"
}

# Get deployment status
GET /api/v1/deployments/{id}

# Rollback deployment
POST /api/v1/deployments/{id}/rollback
```

## Example: Full Workflow

### Step 1: Developer Pushes Code

```bash
git commit -m "Add new feature"
git push origin main
```

### Step 2: CI Instance Triggered

```
[Webhook] → [Ritmo CI] → [Build & Test] → [Create Artifact]
                             ↓
                    ✅ Build #142 Success
                    📦 Artifact: backend-1.2.3.jar
                    🏷️  Tag: dev
                    🔗 URL: s3://artifacts/backend-1.2.3.jar
```

### Step 3: Auto-Promote to Staging

```bash
# CI completes successfully, auto-promote
POST /api/v1/artifacts/art-456/promote
{
  "from": "dev",
  "to": "staging",
  "auto": true
}

Response:
{
  "status": "promoted",
  "environment": "staging",
  "promoted_at": "2024-01-15T10:30:00Z"
}
```

### Step 4: Trigger CD System

```bash
# Ritmo sends webhook to GitHub Actions
POST https://api.github.com/repos/myorg/backend/dispatches
{
  "event_type": "deploy",
  "client_payload": {
    "artifact_url": "s3://artifacts/backend-1.2.3.jar",
    "version": "1.2.3",
    "environment": "staging"
  }
}
```

### Step 5: GitHub Actions Deploys

```yaml
# GitHub Actions workflow executes
- Download artifact from S3
- Push to container registry
- Deploy to Kubernetes staging
- Run smoke tests
- Notify team
```

### Step 6: Manual Promotion to Production

```bash
# Ops team approves after testing
POST /api/v1/artifacts/art-456/promote
{
  "from": "staging",
  "to": "production",
  "approved_by": ["alice@co.com", "bob@co.com"],
  "approval_ticket": "JIRA-1234",
  "notes": "All tests passed in staging"
}
```

### Step 7: Production Deployment

```bash
# CD system deploys to production
- Canary deployment (10% traffic)
- Monitor metrics
- Gradual rollout to 100%
- Mark artifact as production-stable
```

## Security Considerations

### Credential Separation

**CI Instance** has access to:
- ✅ Source code repositories
- ✅ Build tools
- ✅ Artifact storage (write)
- ❌ Production infrastructure
- ❌ Production secrets
- ❌ Deployment credentials

**CD Instance** has access to:
- ✅ Artifact storage (read)
- ✅ Deployment targets
- ✅ Production secrets (via Vault)
- ❌ Source code repositories
- ❌ Build tools

### Artifact Verification

```bash
# CI signs artifacts
gpg --detach-sign backend-service.jar

# CD verifies signatures before deployment
gpg --verify backend-service.jar.sig backend-service.jar
```

### Approval Workflows

```yaml
promotion_rules:
  - from: dev
    to: staging
    requires_approval: false
    auto_promote: true
  
  - from: staging
    to: production
    requires_approval: true
    min_approvers: 2
    approver_groups: [ops-team, security-team]
    approval_timeout: 7d
```

## Monitoring and Observability

### Metrics

- Artifact promotion time (dev → staging → production)
- Deployment success rate per environment
- Time to production (commit to deployment)
- Rollback frequency
- Approval wait time

### Tracing

```
CI Build [span:build-142]
  ├─ Checkout Code
  ├─ Run Tests
  ├─ Build Artifact [artifact:backend-1.2.3]
  └─ Upload to S3

Artifact Promotion [span:promote-art-456]
  ├─ Validate Artifact
  ├─ Update Metadata
  └─ Trigger CD [correlation-id:xyz]

CD Deployment [span:deploy-staging-89]
  ├─ Download Artifact [artifact:backend-1.2.3]
  ├─ Deploy to K8s
  ├─ Run Smoke Tests
  └─ Update Status
```

## Best Practices

1. **Immutable Artifacts**: Never modify artifacts, only promote them
2. **Metadata Rich**: Tag artifacts with commit, version, tests run
3. **Automated Dev/Staging**: Auto-promote and deploy to lower environments
4. **Manual Production**: Require human approval for production
5. **Rollback Ready**: Always keep previous versions available
6. **Audit Trail**: Log all promotions and deployments
7. **Environment Parity**: Keep environments as similar as possible
8. **Progressive Delivery**: Use canary, blue-green for production

## Next Steps

- See [Plugin Development](../plugin-sdk/README.md) for creating deployment plugins
- Check [API Documentation](API.md) for full API reference
- Read [GitOps Integration](gitops.md) for ArgoCD/Flux patterns
