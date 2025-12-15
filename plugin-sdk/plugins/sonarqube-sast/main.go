package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/solvyd/solvyd/plugin-sdk/pkg/sdk"
)

// SonarQubeSASTPlugin implements SAST plugin for SonarQube code analysis
type SonarQubeSASTPlugin struct {
	serverURL      string
	token          string
	projectKey     string
	qualityGate    string
	sources        string
	timeout        int
	scannerVersion string
}

func (p *SonarQubeSASTPlugin) Name() string {
	return "sonarqube-sast"
}

func (p *SonarQubeSASTPlugin) Version() string {
	return "1.0.0"
}

func (p *SonarQubeSASTPlugin) Type() string {
	return "security"
}

func (p *SonarQubeSASTPlugin) Initialize(config map[string]interface{}) error {
	p.serverURL = getStringConfig(config, "server_url", "http://localhost:9000")
	p.token = getStringConfig(config, "token", os.Getenv("SONAR_TOKEN"))
	p.projectKey = getStringConfig(config, "project_key", "")
	p.qualityGate = getStringConfig(config, "quality_gate", "Sonar way")
	p.sources = getStringConfig(config, "sources", ".")
	p.timeout = getIntConfig(config, "timeout", 300)
	p.scannerVersion = getStringConfig(config, "scanner_version", "5.0.1.3006")

	if p.token == "" {
		return fmt.Errorf("sonarqube token is required (set token in config or SONAR_TOKEN env var)")
	}

	return nil
}

func (p *SonarQubeSASTPlugin) Execute(ctx *sdk.ExecutionContext) (*sdk.Result, error) {
	ctx.Logger.Info("Starting SonarQube SAST analysis")

	// Generate project key if not provided
	if p.projectKey == "" {
		p.projectKey = fmt.Sprintf("ritmo-%s-%s", ctx.JobID, ctx.BuildID)
	}

	// Ensure sonar-scanner is available
	scannerPath, err := p.ensureSonarScanner()
	if err != nil {
		return &sdk.Result{
			Success:      false,
			ErrorMessage: fmt.Sprintf("Failed to setup sonar-scanner: %v", err),
		}, err
	}

	// Run SonarQube analysis
	if err := p.runSonarScan(ctx, scannerPath); err != nil {
		return &sdk.Result{
			Success:      false,
			ErrorMessage: fmt.Sprintf("SonarQube scan failed: %v", err),
		}, err
	}

	// Wait for analysis to complete and check quality gate
	ctx.Logger.Info("Waiting for SonarQube analysis to complete...")
	passed, metrics, err := p.waitForAnalysisAndCheckQualityGate()
	if err != nil {
		return &sdk.Result{
			Success:      false,
			ErrorMessage: fmt.Sprintf("Failed to check quality gate: %v", err),
		}, err
	}

	// Build result
	result := &sdk.Result{
		Success:  passed,
		ExitCode: 0,
		Metadata: make(map[string]interface{}),
	}

	if !passed {
		result.ExitCode = 1
		result.ErrorMessage = "Quality gate failed"
	}

	// Add metrics to result
	for key, value := range metrics {
		result.Metadata[key] = value
	}

	result.Output = fmt.Sprintf("SonarQube analysis complete. Quality Gate: %s", map[bool]string{true: "PASSED", false: "FAILED"}[passed])
	ctx.Logger.Info(result.Output)

	return result, nil
}

func (p *SonarQubeSASTPlugin) ensureSonarScanner() (string, error) {
	// Check if sonar-scanner is already in PATH
	if path, err := exec.LookPath("sonar-scanner"); err == nil {
		return path, nil
	}

	// Download and setup sonar-scanner
	scannerDir := filepath.Join(os.TempDir(), "sonar-scanner")
	scannerBin := filepath.Join(scannerDir, "bin", "sonar-scanner")

	if _, err := os.Stat(scannerBin); err == nil {
		return scannerBin, nil
	}

	// TODO: Implement scanner download logic
	// For now, assume it's installed
	return "sonar-scanner", nil
}

func (p *SonarQubeSASTPlugin) runSonarScan(ctx *sdk.ExecutionContext, scannerPath string) error {
	args := []string{
		fmt.Sprintf("-Dsonar.projectKey=%s", p.projectKey),
		fmt.Sprintf("-Dsonar.sources=%s", p.sources),
		fmt.Sprintf("-Dsonar.host.url=%s", p.serverURL),
		fmt.Sprintf("-Dsonar.login=%s", p.token),
		fmt.Sprintf("-Dsonar.qualitygate.wait=true"),
		fmt.Sprintf("-Dsonar.qualitygate.timeout=%d", p.timeout),
	}

	cmd := exec.Command(scannerPath, args...)
	cmd.Dir = ctx.WorkDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = append(os.Environ(), fmt.Sprintf("SONAR_TOKEN=%s", p.token))

	return cmd.Run()
}

func (p *SonarQubeSASTPlugin) waitForAnalysisAndCheckQualityGate() (bool, map[string]interface{}, error) {
	client := &http.Client{Timeout: time.Duration(p.timeout) * time.Second}

	// Get project status
	url := fmt.Sprintf("%s/api/qualitygates/project_status?projectKey=%s", p.serverURL, p.projectKey)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return false, nil, err
	}

	req.SetBasicAuth(p.token, "")

	// Poll for results
	maxAttempts := p.timeout / 5
	for i := 0; i < maxAttempts; i++ {
		resp, err := client.Do(req)
		if err != nil {
			time.Sleep(5 * time.Second)
			continue
		}

		if resp.StatusCode == http.StatusOK {
			var result struct {
				ProjectStatus struct {
					Status     string `json:"status"`
					Conditions []struct {
						Status         string `json:"status"`
						MetricKey      string `json:"metricKey"`
						Comparator     string `json:"comparator"`
						ErrorThreshold string `json:"errorThreshold"`
						ActualValue    string `json:"actualValue"`
					} `json:"conditions"`
				} `json:"projectStatus"`
			}

			if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
				resp.Body.Close()
				return false, nil, err
			}
			resp.Body.Close()

			// Extract metrics
			metrics := make(map[string]interface{})
			metrics["quality_gate_status"] = result.ProjectStatus.Status
			metrics["conditions"] = result.ProjectStatus.Conditions

			passed := result.ProjectStatus.Status == "OK"
			return passed, metrics, nil
		}

		resp.Body.Close()
		time.Sleep(5 * time.Second)
	}

	return false, nil, fmt.Errorf("timeout waiting for analysis results")
}

func (p *SonarQubeSASTPlugin) Cleanup() error {
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

// Export the plugin
var Plugin SonarQubeSASTPlugin

func main() {
	fmt.Println("SonarQube SAST Plugin v1.0.0")
	fmt.Println("This plugin performs static application security testing using SonarQube")
	fmt.Println("To build: go build -o sonarqube-sast")
}
