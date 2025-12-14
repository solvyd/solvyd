# Ritmo Enterprise Security - Quick Reference

## 🚀 Quick Start

```bash
# Start enterprise infrastructure (SonarQube, Trivy, OWASP ZAP, Nexus, etc.)
make enterprise-start

# Build all security plugins
make build-enterprise-plugins

# Full setup
make enterprise-setup
```

## 🔒 Security Plugins

| Plugin | Purpose | Industry Tool |
|--------|---------|---------------|
| SonarQube SAST | Code quality, bugs, vulnerabilities, duplications | SonarQube |
| Trivy Container Scan | Container image vulnerabilities | Aqua Trivy |
| OWASP ZAP DAST | Dynamic security testing | OWASP ZAP |
| OWASP Dependency-Check | Dependency vulnerabilities | OWASP Dependency-Check |
| License Compliance | License policy enforcement | OSS License Tools |
| JUnit Test Reporter | Test results & coverage | JUnit/TestNG/pytest/Jest |

## 📊 Enterprise Services

| Service | URL | Purpose |
|---------|-----|---------|
| SonarQube | http://localhost:9090 | SAST & code quality |
| Trivy Server | http://localhost:8082 | Container scanning |
| OWASP ZAP | http://localhost:8081 | DAST testing |
| Nexus | http://localhost:8083 | Artifact repository |
| Grafana | http://localhost:3001 | Metrics dashboard |
| Prometheus | http://localhost:9091 | Metrics collection |
| Vault | http://localhost:8200 | Secrets management |

## 🎯 Security Pipeline Example

```yaml
apiVersion: ritmo.dev/v1
kind: Job
metadata:
  name: enterprise-app
spec:
  pipeline:
    stages:
      # SAST
      - name: code-quality
        steps:
          - plugin: sonarqube-sast
            config:
              server_url: http://sonarqube:9000
              token: ${SONAR_TOKEN}
              quality_gate: "Enterprise"
      
      # Dependency Security
      - name: dependency-scan
        steps:
          - plugin: owasp-dependency-check
            config:
              fail_on_cvss: 7.0
          - plugin: license-compliance
            config:
              fail_on_denied: true
      
      # Build & Test
      - name: test
        steps:
          - plugin: junit-test-reporter
            config:
              coverage_min: 80.0
      
      # Container Security
      - name: container-scan
        steps:
          - plugin: trivy-container-scan
            config:
              image: myapp:${BUILD_NUMBER}
              severity: [CRITICAL, HIGH]
      
      # DAST
      - name: dynamic-scan
        steps:
          - plugin: owasp-zap-dast
            config:
              target_url: https://staging.myapp.com
              scan_type: full
```

## 🛡️ Quality Gates

### Default Thresholds

| Metric | Threshold | Blocker |
|--------|-----------|---------|
| Critical Security Issues | 0 | ✅ Yes |
| High Security Issues | 0 | ✅ Yes |
| Container CVE (Critical) | 0 | ✅ Yes |
| Code Coverage | ≥ 80% | ✅ Yes |
| License Violations | 0 | ✅ Yes |
| Code Duplications | ≤ 3% | ❌ No |
| Maintainability Rating | A or B | ❌ No |

## 📈 Metrics Tracked

### Security
- Vulnerabilities by severity (Critical/High/Medium/Low)
- CVSS scores
- Mean Time to Remediation (MTTR)
- Security gate pass rate

### Quality
- Code coverage %
- Bugs and code smells
- Technical debt
- Duplicate code %

### Compliance
- License compliance rate
- Policy violations
- SBOM generation status

## 🔧 Configuration

### Environment Variables

```bash
# SonarQube
export SONAR_HOST_URL=http://localhost:9090
export SONAR_TOKEN=your-token

# OWASP ZAP
export ZAP_API_KEY=ritmo-zap-api-key

# Trivy
export TRIVY_SERVER=http://localhost:8082

# Nexus
export NEXUS_URL=http://localhost:8083
export NEXUS_USERNAME=admin
export NEXUS_PASSWORD=admin123
```

### Plugin Configuration Files

Place in GitOps repository (`ritmo-config/plugins/`):

```yaml
# sonarqube.yaml
apiVersion: ritmo.dev/v1
kind: Plugin
metadata:
  name: sonarqube-sast
spec:
  default_config:
    server_url: http://sonarqube:9000
    quality_gate: "Sonar way"
    timeout: 300
```

## 🚨 Vulnerability Severity Guide

| CVSS Score | Severity | SLA |
|------------|----------|-----|
| 9.0 - 10.0 | CRITICAL | 24 hours |
| 7.0 - 8.9 | HIGH | 7 days |
| 4.0 - 6.9 | MEDIUM | 30 days |
| 0.1 - 3.9 | LOW | Next release |

## 📚 Documentation

- [Complete Enterprise Security Guide](docs/ENTERPRISE-SECURITY.md)
- [Plugin Development](plugin-sdk/README.md)
- [GitOps Configuration](docs/GITOPS-CONFIG.md)
- [Architecture Overview](ARCHITECTURE.md)

## 🔍 Troubleshooting

### Services Not Starting

```bash
# Check logs
make enterprise-logs

# Restart specific service
docker-compose -f docker-compose.enterprise.yml restart sonarqube

# Check service health
docker-compose -f docker-compose.enterprise.yml ps
```

### Plugin Build Failures

```bash
# Clean and rebuild
cd plugin-sdk/examples/sonarqube-sast
go clean
go mod tidy
go build
```

### Scan Failures

```bash
# Increase timeout in job config
timeout: 600

# Check service connectivity
curl http://localhost:9090/api/system/status  # SonarQube
curl http://localhost:8082/healthz            # Trivy
curl http://localhost:8081                     # ZAP
```

## 💡 Best Practices

1. **Run security scans on every commit**
2. **Block builds on critical/high vulnerabilities**
3. **Update vulnerability databases weekly**
4. **Review and tune quality gates monthly**
5. **Maintain suppression files for false positives**
6. **Generate and track SBOMs**
7. **Rotate credentials regularly**
8. **Monitor security metrics dashboards**

## 🎓 Training Resources

- [OWASP Top 10](https://owasp.org/www-project-top-ten/)
- [CWE Top 25](https://cwe.mitre.org/top25/)
- [NIST CVE Database](https://nvd.nist.gov/)
- [SonarQube Documentation](https://docs.sonarqube.org/)
- [Trivy Documentation](https://aquasecurity.github.io/trivy/)
- [OWASP ZAP Documentation](https://www.zaproxy.org/docs/)
