package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/solvyd/solvyd/plugin-sdk/pkg/sdk"
)

// OWASPDependencyCheckPlugin scans project dependencies for known vulnerabilities
type OWASPDependencyCheckPlugin struct {
	projectPath        string
	scanPath           string
	failOnCVSS         float64
	format             string
	suppressionFile    string
	enableExperimental bool
	timeout            int
}

type DependencyCheckReport struct {
	Dependencies []struct {
		FileName        string `json:"fileName"`
		FilePath        string `json:"filePath"`
		Vulnerabilities []struct {
			Name        string  `json:"name"`
			CVSSV2      float64 `json:"cvssv2"`
			CVSSV3      float64 `json:"cvssv3"`
			Severity    string  `json:"severity"`
			Description string  `json:"description"`
			References  []struct {
				Source string `json:"source"`
				URL    string `json:"url"`
			} `json:"references"`
		} `json:"vulnerabilities"`
	} `json:"dependencies"`
}

func (p *OWASPDependencyCheckPlugin) Name() string {
	return "owasp-dependency-check"
}

func (p *OWASPDependencyCheckPlugin) Version() string {
	return "1.0.0"
}

func (p *OWASPDependencyCheckPlugin) Type() string {
	return "security"
}

func (p *OWASPDependencyCheckPlugin) Initialize(config map[string]interface{}) error {
	p.projectPath = getStringConfig(config, "project_path", ".")
	p.scanPath = getStringConfig(config, "scan_path", ".")
	p.failOnCVSS = getFloatConfig(config, "fail_on_cvss", 7.0)
	p.format = getStringConfig(config, "format", "JSON")
	p.suppressionFile = getStringConfig(config, "suppression_file", "")
	p.enableExperimental = getBoolConfig(config, "enable_experimental", false)
	p.timeout = getIntConfig(config, "timeout", 600)

	return nil
}

func (p *OWASPDependencyCheckPlugin) Execute(ctx *sdk.ExecutionContext) (*sdk.Result, error) {
	ctx.Logger.Info("Starting OWASP Dependency-Check scan")

	outputDir := filepath.Join(ctx.WorkDir, "dependency-check-report")
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return &sdk.Result{
			Success:      false,
			ErrorMessage: fmt.Sprintf("Failed to create output directory: %v", err),
		}, err
	}

	// Build dependency-check command
	args := []string{
		"run", "--rm",
		"-v", fmt.Sprintf("%s:/src:ro", filepath.Join(ctx.WorkDir, p.scanPath)),
		"-v", fmt.Sprintf("%s:/report", outputDir),
		"owasp/dependency-check",
		"--scan", "/src",
		"--format", p.format,
		"--out", "/report",
		"--project", ctx.JobID,
	}

	if p.suppressionFile != "" {
		args = append(args, "--suppression", p.suppressionFile)
	}

	if p.enableExperimental {
		args = append(args, "--enableExperimental")
	}

	cmd := exec.Command("docker", args...)
	cmd.Dir = ctx.WorkDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	ctx.Logger.Info("Running dependency-check in Docker container...")
	if err := cmd.Run(); err != nil {
		return &sdk.Result{
			Success:      false,
			ErrorMessage: fmt.Sprintf("Dependency-check failed: %v", err),
		}, err
	}

	// Parse results
	reportFile := filepath.Join(outputDir, "dependency-check-report.json")
	data, err := os.ReadFile(reportFile)
	if err != nil {
		return &sdk.Result{
			Success:      false,
			ErrorMessage: fmt.Sprintf("Failed to read report: %v", err),
		}, err
	}

	var report DependencyCheckReport
	if err := json.Unmarshal(data, &report); err != nil {
		return &sdk.Result{
			Success:      false,
			ErrorMessage: fmt.Sprintf("Failed to parse report: %v", err),
		}, err
	}

	// Analyze vulnerabilities
	totalVulns := 0
	highSeverityVulns := 0
	vulnsByCVSS := make(map[string]int)

	for _, dep := range report.Dependencies {
		for _, vuln := range dep.Vulnerabilities {
			totalVulns++
			cvss := vuln.CVSSV3
			if cvss == 0 {
				cvss = vuln.CVSSV2
			}

			if cvss >= p.failOnCVSS {
				highSeverityVulns++
			}

			// Categorize by severity
			if cvss >= 9.0 {
				vulnsByCVSS["CRITICAL"]++
			} else if cvss >= 7.0 {
				vulnsByCVSS["HIGH"]++
			} else if cvss >= 4.0 {
				vulnsByCVSS["MEDIUM"]++
			} else {
				vulnsByCVSS["LOW"]++
			}
		}
	}

	// Build result
	result := &sdk.Result{
		Success:  highSeverityVulns == 0,
		ExitCode: 0,
		Metadata: make(map[string]interface{}),
		Output:   fmt.Sprintf("Found %d vulnerabilities (%d above CVSS %.1f)", totalVulns, highSeverityVulns, p.failOnCVSS),
	}

	if highSeverityVulns > 0 {
		result.ExitCode = 1
		result.ErrorMessage = fmt.Sprintf("Found %d high-severity vulnerabilities", highSeverityVulns)
	}

	result.Metadata["total_vulnerabilities"] = totalVulns
	result.Metadata["high_severity_count"] = highSeverityVulns
	result.Metadata["vulnerabilities_by_severity"] = vulnsByCVSS
	result.Metadata["cvss_threshold"] = p.failOnCVSS

	ctx.Logger.Info(result.Output)
	for severity, count := range vulnsByCVSS {
		ctx.Logger.Info(fmt.Sprintf("  %s: %d", severity, count))
	}

	return result, nil
}

func (p *OWASPDependencyCheckPlugin) Cleanup() error {
	return nil
}

// Helper functions
func getStringConfig(config map[string]interface{}, key, defaultValue string) string {
	if val, ok := config[key].(string); ok {
		return val
	}
	return defaultValue
}

func getFloatConfig(config map[string]interface{}, key string, defaultValue float64) float64 {
	if val, ok := config[key].(float64); ok {
		return val
	}
	return defaultValue
}

func getIntConfig(config map[string]interface{}, key string, defaultValue int) int {
	if val, ok := config[key].(float64); ok {
		return int(val)
	}
	return defaultValue
}

func getBoolConfig(config map[string]interface{}, key string, defaultValue bool) bool {
	if val, ok := config[key].(bool); ok {
		return val
	}
	return defaultValue
}

// Export the plugin
var Plugin OWASPDependencyCheckPlugin

func main() {
	fmt.Println("OWASP Dependency-Check Plugin v1.0.0")
	fmt.Println("This plugin scans project dependencies for known vulnerabilities")
	fmt.Println("To build: go build -o owasp-dependency-check")
}
