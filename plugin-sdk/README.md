# Solvyd Plugin SDK

Plugin SDK for developing Solvyd plugins.

## Plugin Types

1. **SCM Plugins**: Source control management (Git, GitHub, GitLab, Bitbucket, SVN)
2. **Build Plugins**: Build tools (Maven, Gradle, npm, Go, Python, etc.)
3. **Artifact Plugins**: Artifact storage (S3, GCS, Artifactory, Nexus)
4. **Notification Plugins**: Notifications (Slack, Email, Teams, PagerDuty)
5. **Deployment Plugins**: Deployment targets (Kubernetes, Docker, SSH, ArgoCD)
6. **Test Plugins**: Test runners and reporters (JUnit, pytest, Jest)
7. **Security Plugins**: Security scanning (SonarQube, Snyk, Trivy)

## Plugin Interface

All plugins must implement the base `Plugin` interface:

```go
type Plugin interface {
    // Name returns the plugin name
    Name() string
    
    // Version returns the plugin version
    Version() string
    
    // Type returns the plugin type (scm, build, artifact, etc.)
    Type() string
    
    // Initialize initializes the plugin with configuration
    Initialize(config map[string]interface{}) error
    
    // Execute executes the plugin
    Execute(context *ExecutionContext) (*Result, error)
    
    // Cleanup performs cleanup after execution
    Cleanup() error
}
```

## Execution Context

```go
type ExecutionContext struct {
    BuildID       string
    JobID         string
    WorkDir       string
    EnvVars       map[string]string
    Parameters    map[string]interface{}
    Secrets       map[string]string
    Logger        Logger
}
```

## Plugin Result

```go
type Result struct {
    Success      bool
    ExitCode     int
    ErrorMessage string
    Output       string
    Artifacts    []Artifact
    Metadata     map[string]interface{}
}
```

## Creating a Plugin

### 1. Implement the Plugin Interface

```go
package main

import "github.com/solvyd/solvyd/plugin-sdk/pkg/sdk"

type MyPlugin struct {
    config map[string]interface{}
}

func (p *MyPlugin) Name() string {
    return "my-plugin"
}

func (p *MyPlugin) Version() string {
    return "1.0.0"
}

func (p *MyPlugin) Type() string {
    return "build"
}

func (p *MyPlugin) Initialize(config map[string]interface{}) error {
    p.config = config
    return nil
}

func (p *MyPlugin) Execute(ctx *sdk.ExecutionContext) (*sdk.Result, error) {
    // Your plugin logic here
    return &sdk.Result{
        Success: true,
        ExitCode: 0,
    }, nil
}

func (p *MyPlugin) Cleanup() error {
    return nil
}

// Export the plugin
var Plugin MyPlugin
```

### 2. Build the Plugin

```bash
# Build as Go plugin (.so)
go build -buildmode=plugin -o my-plugin.so

# Or build as standalone binary
go build -o my-plugin
```

## Core Plugins

See the `plugins/` directory for official plugin implementations:

### SCM Plugins
- `git-scm/` - Git SCM plugin

### Notification Plugins
- `slack-notify/` - Slack notification plugin

### Security Plugins (Enterprise)
- `sonarqube-sast/` - SonarQube static analysis (code quality, bugs, vulnerabilities, duplications)
- `trivy-container-scan/` - Container image vulnerability scanning
- `owasp-zap-dast/` - Dynamic application security testing
- `owasp-dependency-check/` - Dependency vulnerability scanning
- `license-compliance/` - License compliance and attribution

### Test Plugins
- `junit-test-reporter/` - JUnit/TestNG test result parser and reporter

For comprehensive enterprise security setup, see [Enterprise Security Guide](../docs/ENTERPRISE-SECURITY.md).

## Plugin Configuration

Plugins receive configuration via the `Initialize()` method:

```yaml
plugins:
  - name: git-scm
    config:
      depth: 1
      submodules: true
      
  - name: maven-build
    config:
      goals: ["clean", "package"]
      profiles: ["production"]
      jdk_version: "17"
```

## Best Practices

1. **Error Handling**: Always return meaningful error messages
2. **Logging**: Use the provided logger for structured logging
3. **Cleanup**: Always clean up resources in `Cleanup()`
4. **Validation**: Validate configuration in `Initialize()`
5. **Idempotency**: Make operations idempotent where possible
6. **Secrets**: Never log secrets or sensitive data
7. **Timeouts**: Respect context cancellation
8. **Versioning**: Follow semantic versioning

## Testing

```go
func TestMyPlugin(t *testing.T) {
    plugin := &MyPlugin{}
    
    err := plugin.Initialize(map[string]interface{}{
        "key": "value",
    })
    assert.NoError(t, err)
    
    ctx := &sdk.ExecutionContext{
        BuildID: "test-build",
        WorkDir: "/tmp/test",
    }
    
    result, err := plugin.Execute(ctx)
    assert.NoError(t, err)
    assert.True(t, result.Success)
}
```

## Publishing

Plugins can be published to:
1. Solvyd Plugin Marketplace (future)
2. GitHub releases
3. Private repositories
4. Container registries (for containerized plugins)
