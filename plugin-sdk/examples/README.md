# Enterprise Security Plugins

This directory contains production-ready security, quality, and compliance plugins for enterprise use.

## Available Plugins

### Security Scanning

1. **SonarQube SAST** (`sonarqube-sast/`)
   - Static code analysis
   - Security vulnerability detection
   - Code quality metrics
   - Technical debt tracking
   - Duplicate code detection

2. **Trivy Container Scanner** (`trivy-container-scan/`)
   - Container image vulnerability scanning
   - OS package CVE detection
   - Application dependency scanning
   - Misconfiguration detection

3. **OWASP ZAP DAST** (`owasp-zap-dast/`)
   - Dynamic application security testing
   - OWASP Top 10 vulnerability detection
   - API security testing
   - Authentication testing

4. **OWASP Dependency-Check** (`owasp-dependency-check/`)
   - Dependency vulnerability scanning
   - CVE matching
   - Multi-language support
   - CVSS scoring

5. **License Compliance Scanner** (`license-compliance/`)
   - License detection
   - Policy enforcement
   - Attribution generation
   - Multi-ecosystem support

### Testing & Quality

6. **JUnit Test Reporter** (`junit-test-reporter/`)
   - Test result aggregation
   - Coverage analysis
   - Trend reporting
   - Multi-framework support

## Quick Start

### Build All Enterprise Plugins

```bash
make -f Makefile.enterprise build-enterprise-plugins
```

### Start Enterprise Infrastructure

```bash
make -f Makefile.enterprise enterprise-start
```

This starts:
- SonarQube (code quality)
- Trivy Server (container scanning)
- OWASP ZAP (DAST)
- Nexus (artifact repository)
- Prometheus & Grafana (monitoring)
- HashiCorp Vault (secrets)

### Configure in Job

```yaml
apiVersion: ritmo.dev/v1
kind: Job
metadata:
  name: secure-build
spec:
  pipeline:
    stages:
      - name: security-scan
        steps:
          - plugin: sonarqube-sast
            config:
              server_url: http://sonarqube:9000
              token: ${SONAR_TOKEN}
              project_key: my-app
          
          - plugin: trivy-container-scan
            config:
              image: myapp:latest
              severity: [CRITICAL, HIGH]
          
          - plugin: owasp-dependency-check
            config:
              fail_on_cvss: 7.0
```

## Plugin Details

### SonarQube SAST

**Type**: Security  
**Language**: Go  
**Dependencies**: SonarScanner CLI

**Metrics Tracked**:
- Bugs, Vulnerabilities, Code Smells
- Security/Reliability/Maintainability Ratings
- Code Coverage, Duplications
- Cyclomatic Complexity, Technical Debt

**Quality Gate**: Configurable thresholds for all metrics

### Trivy Container Scan

**Type**: Security  
**Language**: Go  
**Dependencies**: Trivy CLI or Server

**Features**:
- Multi-layer scanning
- Language-specific scanners (Go, Java, Python, etc.)
- Secret detection
- SBOM generation (CycloneDX, SPDX)

**Severities**: CRITICAL, HIGH, MEDIUM, LOW

### OWASP ZAP DAST

**Type**: Security  
**Language**: Go  
**Dependencies**: OWASP ZAP Server

**Scan Types**:
- Baseline: Quick passive scan
- Full: Active attack simulation
- API: REST/GraphQL testing

**Vulnerability Coverage**: OWASP Top 10, CWE Top 25

### OWASP Dependency-Check

**Type**: Security  
**Language**: Go  
**Dependencies**: OWASP Dependency-Check Docker

**Supported Ecosystems**:
- Java (Maven, Gradle)
- JavaScript (npm, yarn)
- Python (pip)
- .NET (NuGet)
- Go modules
- Ruby gems

### License Compliance

**Type**: Security/Compliance  
**Language**: Go  
**Dependencies**: license-checker, pip-licenses, go-licenses

**Policy Options**:
- Allowed licenses (whitelist)
- Denied licenses (blacklist)
- Unknown license handling
- Attribution generation

### JUnit Test Reporter

**Type**: Testing  
**Language**: Go  
**Dependencies**: None

**Formats Supported**:
- JUnit XML
- TestNG XML
- pytest JUnit XML
- Jest JUnit XML
- Go test output

**Reports**:
- Pass/Fail/Skip counts
- Duration analysis
- Coverage metrics
- Failure details

## Integration

### With GitOps

```yaml
# jobs/my-app.yaml
apiVersion: ritmo.dev/v1
kind: Job
metadata:
  name: my-app
spec:
  plugins:
    - name: sonarqube-sast
    - name: trivy-container-scan
    - name: owasp-dependency-check
    - name: junit-test-reporter
    - name: license-compliance
```

### With API

```bash
curl -X POST http://localhost:8080/api/v1/jobs \
  -H "Content-Type: application/json" \
  -d '{
    "name": "secure-app",
    "plugins": [
      {"name": "sonarqube-sast", "config": {...}},
      {"name": "trivy-container-scan", "config": {...}}
    ]
  }'
```

## Security Best Practices

1. **Secrets Management**: Store tokens in Vault or environment variables
2. **Network Isolation**: Run security tools in isolated networks
3. **Regular Updates**: Update vulnerability databases weekly
4. **False Positives**: Maintain suppression files for known false positives
5. **Threshold Tuning**: Adjust CVSS thresholds based on risk appetite
6. **Compliance**: Document all security exceptions

## Monitoring

All plugins emit metrics to Prometheus:
- `ritmo_plugin_duration_seconds{plugin="sonarqube-sast"}`
- `ritmo_security_vulnerabilities_total{severity="high"}`
- `ritmo_quality_gate_failures_total`
- `ritmo_test_failures_total`

## Troubleshooting

### SonarQube Connection Failed
```bash
# Check SonarQube status
curl http://localhost:9090/api/system/status

# Verify token
curl -u TOKEN: http://localhost:9090/api/authentication/validate
```

### Trivy Scan Timeout
```bash
# Use server mode
trivy-server: http://localhost:8082

# Update database manually
trivy image --download-db-only
```

### ZAP Scan Incomplete
```bash
# Increase timeout
timeout: 900

# Check ZAP logs
docker logs ritmo-owasp-zap
```

## Contributing

See [Plugin Development Guide](../README.md) for creating custom security plugins.

## License

All plugins follow the same license as Ritmo core.
