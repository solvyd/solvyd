package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/vrenjith/ritmo/plugin-sdk/pkg/sdk"
)

// OWASPZAPDASTPlugin implements Dynamic Application Security Testing using OWASP ZAP
type OWASPZAPDASTPlugin struct {
	targetURL  string
	zapURL     string
	apiKey     string
	scanType   string // baseline, full, api
	timeout    int
	alertLevel string // High, Medium, Low
}

type ZAPAlert struct {
	Alert       string `json:"alert"`
	Risk        string `json:"risk"`
	Confidence  string `json:"confidence"`
	URL         string `json:"url"`
	Description string `json:"description"`
	Solution    string `json:"solution"`
	CWE         string `json:"cweid"`
}

func (p *OWASPZAPDASTPlugin) Name() string {
	return "owasp-zap-dast"
}

func (p *OWASPZAPDASTPlugin) Version() string {
	return "1.0.0"
}

func (p *OWASPZAPDASTPlugin) Type() string {
	return "security"
}

func (p *OWASPZAPDASTPlugin) Initialize(config map[string]interface{}) error {
	p.targetURL = getStringConfig(config, "target_url", "")
	p.zapURL = getStringConfig(config, "zap_url", "http://localhost:8081")
	p.apiKey = getStringConfig(config, "api_key", "ritmo-zap-api-key")
	p.scanType = getStringConfig(config, "scan_type", "baseline")
	p.timeout = getIntConfig(config, "timeout", 600)
	p.alertLevel = getStringConfig(config, "alert_level", "High")

	if p.targetURL == "" {
		return fmt.Errorf("target_url is required for DAST scanning")
	}

	return nil
}

func (p *OWASPZAPDASTPlugin) Execute(ctx *sdk.ExecutionContext) (*sdk.Result, error) {
	ctx.Logger.Info(fmt.Sprintf("Starting OWASP ZAP DAST scan on: %s", p.targetURL))

	client := &http.Client{Timeout: time.Duration(p.timeout) * time.Second}

	// Start spider scan
	ctx.Logger.Info("Starting ZAP spider scan...")
	scanID, err := p.startSpiderScan(client)
	if err != nil {
		return &sdk.Result{
			Success:      false,
			ErrorMessage: fmt.Sprintf("Failed to start spider scan: %v", err),
		}, err
	}

	// Wait for spider to complete
	if err := p.waitForScan(client, scanID, "spider"); err != nil {
		return &sdk.Result{
			Success:      false,
			ErrorMessage: fmt.Sprintf("Spider scan failed: %v", err),
		}, err
	}

	ctx.Logger.Info("Spider scan complete. Starting active scan...")

	// Start active scan
	activeScanID, err := p.startActiveScan(client)
	if err != nil {
		return &sdk.Result{
			Success:      false,
			ErrorMessage: fmt.Sprintf("Failed to start active scan: %v", err),
		}, err
	}

	// Wait for active scan to complete
	if err := p.waitForScan(client, activeScanID, "ascan"); err != nil {
		return &sdk.Result{
			Success:      false,
			ErrorMessage: fmt.Sprintf("Active scan failed: %v", err),
		}, err
	}

	ctx.Logger.Info("Active scan complete. Retrieving alerts...")

	// Get alerts
	alerts, err := p.getAlerts(client)
	if err != nil {
		return &sdk.Result{
			Success:      false,
			ErrorMessage: fmt.Sprintf("Failed to retrieve alerts: %v", err),
		}, err
	}

	// Categorize alerts
	alertCounts := make(map[string]int)
	highRiskAlerts := 0

	for _, alert := range alerts {
		alertCounts[alert.Risk]++
		if alert.Risk == "High" {
			highRiskAlerts++
		}
	}

	// Build result
	result := &sdk.Result{
		Success:  highRiskAlerts == 0,
		ExitCode: 0,
		Metadata: make(map[string]interface{}),
		Output:   fmt.Sprintf("Found %d total alerts (%d high risk)", len(alerts), highRiskAlerts),
	}

	if highRiskAlerts > 0 {
		result.ExitCode = 1
		result.ErrorMessage = fmt.Sprintf("Found %d high risk vulnerabilities", highRiskAlerts)
	}

	result.Metadata["total_alerts"] = len(alerts)
	result.Metadata["alerts_by_risk"] = alertCounts
	result.Metadata["high_risk_count"] = highRiskAlerts

	ctx.Logger.Info(fmt.Sprintf("DAST scan complete. Total alerts: %d, High risk: %d", len(alerts), highRiskAlerts))

	return result, nil
}

func (p *OWASPZAPDASTPlugin) startSpiderScan(client *http.Client) (string, error) {
	zapURL := fmt.Sprintf("%s/JSON/spider/action/scan/?apikey=%s&url=%s", p.zapURL, p.apiKey, url.QueryEscape(p.targetURL))

	resp, err := client.Get(zapURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		Scan string `json:"scan"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	return result.Scan, nil
}

func (p *OWASPZAPDASTPlugin) startActiveScan(client *http.Client) (string, error) {
	zapURL := fmt.Sprintf("%s/JSON/ascan/action/scan/?apikey=%s&url=%s", p.zapURL, p.apiKey, url.QueryEscape(p.targetURL))

	resp, err := client.Get(zapURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		Scan string `json:"scan"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	return result.Scan, nil
}

func (p *OWASPZAPDASTPlugin) waitForScan(client *http.Client, scanID, scanType string) error {
	var statusURL string
	if scanType == "spider" {
		statusURL = fmt.Sprintf("%s/JSON/spider/view/status/?apikey=%s&scanId=%s", p.zapURL, p.apiKey, scanID)
	} else {
		statusURL = fmt.Sprintf("%s/JSON/ascan/view/status/?apikey=%s&scanId=%s", p.zapURL, p.apiKey, scanID)
	}

	for i := 0; i < p.timeout/5; i++ {
		resp, err := client.Get(statusURL)
		if err != nil {
			return err
		}

		var result struct {
			Status string `json:"status"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			resp.Body.Close()
			return err
		}
		resp.Body.Close()

		if result.Status == "100" {
			return nil
		}

		time.Sleep(5 * time.Second)
	}

	return fmt.Errorf("scan timeout")
}

func (p *OWASPZAPDASTPlugin) getAlerts(client *http.Client) ([]ZAPAlert, error) {
	alertsURL := fmt.Sprintf("%s/JSON/core/view/alerts/?apikey=%s&baseurl=%s", p.zapURL, p.apiKey, url.QueryEscape(p.targetURL))

	resp, err := client.Get(alertsURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Alerts []ZAPAlert `json:"alerts"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.Alerts, nil
}

func (p *OWASPZAPDASTPlugin) Cleanup() error {
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
var Plugin OWASPZAPDASTPlugin

func main() {
	fmt.Println("OWASP ZAP DAST Plugin v1.0.0")
	fmt.Println("This plugin performs dynamic application security testing using OWASP ZAP")
	fmt.Println("To build: go build -o owasp-zap-dast")
}
