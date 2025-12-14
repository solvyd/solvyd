package main

import (
	"encoding/json"
	"fmt"
	"os/exec"

	"github.com/vrenjith/ritmo/plugin-sdk/pkg/sdk"
)

// TrivyContainerScanPlugin implements container security scanning using Trivy
type TrivyContainerScanPlugin struct {
	image         string
	severity      []string
	trivyServer   string
	ignoreUnfixed bool
	timeout       int
	exitCode      int
}

type TrivyReport struct {
	Results []struct {
		Target          string `json:"Target"`
		Class           string `json:"Class"`
		Type            string `json:"Type"`
		Vulnerabilities []struct {
			VulnerabilityID  string `json:"VulnerabilityID"`
			PkgName          string `json:"PkgName"`
			InstalledVersion string `json:"InstalledVersion"`
			FixedVersion     string `json:"FixedVersion"`
			Severity         string `json:"Severity"`
			Title            string `json:"Title"`
			Description      string `json:"Description"`
			PrimaryURL       string `json:"PrimaryURL"`
		} `json:"Vulnerabilities"`
	} `json:"Results"`
}

func (p *TrivyContainerScanPlugin) Name() string {
	return "trivy-container-scan"
}

func (p *TrivyContainerScanPlugin) Version() string {
	return "1.0.0"
}

func (p *TrivyContainerScanPlugin) Type() string {
	return "security"
}

func (p *TrivyContainerScanPlugin) Initialize(config map[string]interface{}) error {
	p.image = getStringConfig(config, "image", "")
	p.trivyServer = getStringConfig(config, "trivy_server", "")
	p.ignoreUnfixed = getBoolConfig(config, "ignore_unfixed", false)
	p.timeout = getIntConfig(config, "timeout", 300)
	p.exitCode = getIntConfig(config, "exit_code", 1)

	// Parse severity levels
	if sev, ok := config["severity"].([]interface{}); ok {
		p.severity = make([]string, len(sev))
		for i, s := range sev {
			p.severity[i] = s.(string)
		}
	} else {
		p.severity = []string{"CRITICAL", "HIGH"}
	}

	if p.image == "" {
		return fmt.Errorf("image is required for container scanning")
	}

	return nil
}

func (p *TrivyContainerScanPlugin) Execute(ctx *sdk.ExecutionContext) (*sdk.Result, error) {
	ctx.Logger.Info(fmt.Sprintf("Starting Trivy container scan for image: %s", p.image))

	// Build trivy command
	args := []string{"image", "--format", "json"}

	if p.trivyServer != "" {
		args = append(args, "--server", p.trivyServer)
	}

	if p.ignoreUnfixed {
		args = append(args, "--ignore-unfixed")
	}

	// Add severity filters
	if len(p.severity) > 0 {
		severityStr := ""
		for i, s := range p.severity {
			if i > 0 {
				severityStr += ","
			}
			severityStr += s
		}
		args = append(args, "--severity", severityStr)
	}

	args = append(args, p.image)

	// Run trivy
	cmd := exec.Command("trivy", args...)
	cmd.Dir = ctx.WorkDir
	output, err := cmd.CombinedOutput()

	// Parse results even if command failed
	var report TrivyReport
	if len(output) > 0 {
		if parseErr := json.Unmarshal(output, &report); parseErr != nil {
			ctx.Logger.Error(fmt.Sprintf("Failed to parse Trivy output: %v", parseErr))
		}
	}

	// Count vulnerabilities by severity
	vulnCounts := make(map[string]int)
	totalVulns := 0

	for _, result := range report.Results {
		for _, vuln := range result.Vulnerabilities {
			vulnCounts[vuln.Severity]++
			totalVulns++
		}
	}

	// Build result
	result := &sdk.Result{
		Success:  err == nil && totalVulns == 0,
		ExitCode: 0,
		Metadata: make(map[string]interface{}),
		Output:   string(output),
	}

	if totalVulns > 0 {
		result.ExitCode = p.exitCode
		result.ErrorMessage = fmt.Sprintf("Found %d vulnerabilities", totalVulns)
	}

	// Add vulnerability counts to metadata
	result.Metadata["total_vulnerabilities"] = totalVulns
	result.Metadata["vulnerabilities_by_severity"] = vulnCounts
	result.Metadata["scanned_image"] = p.image

	ctx.Logger.Info(fmt.Sprintf("Trivy scan complete. Found %d vulnerabilities", totalVulns))
	for severity, count := range vulnCounts {
		ctx.Logger.Info(fmt.Sprintf("  %s: %d", severity, count))
	}

	return result, nil
}

func (p *TrivyContainerScanPlugin) Cleanup() error {
	return nil
}

// Helper functions
func getStringConfig(config map[string]interface{}, key, defaultValue string) string {
	if val, ok := config[key].(string); ok {
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
var Plugin TrivyContainerScanPlugin

func main() {
	fmt.Println("Trivy Container Scan Plugin v1.0.0")
	fmt.Println("This plugin scans container images for vulnerabilities using Trivy")
	fmt.Println("To build: go build -o trivy-container-scan")
}
