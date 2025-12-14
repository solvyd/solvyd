package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/vrenjith/ritmo/plugin-sdk/pkg/sdk"
)

// LicenseCompliancePlugin scans and validates software licenses
type LicenseCompliancePlugin struct {
	scanPath        string
	allowedLicenses []string
	deniedLicenses  []string
	failOnDenied    bool
	failOnUnknown   bool
	generateSBOM    bool
}

type License struct {
	Name       string `json:"name"`
	Package    string `json:"package"`
	Version    string `json:"version"`
	License    string `json:"license"`
	Repository string `json:"repository"`
	Approved   bool   `json:"approved"`
}

func (p *LicenseCompliancePlugin) Name() string {
	return "license-compliance"
}

func (p *LicenseCompliancePlugin) Version() string {
	return "1.0.0"
}

func (p *LicenseCompliancePlugin) Type() string {
	return "compliance"
}

func (p *LicenseCompliancePlugin) Initialize(config map[string]interface{}) error {
	p.scanPath = getStringConfig(config, "scan_path", ".")
	p.failOnDenied = getBoolConfig(config, "fail_on_denied", true)
	p.failOnUnknown = getBoolConfig(config, "fail_on_unknown", false)
	p.generateSBOM = getBoolConfig(config, "generate_sbom", true)

	// Parse allowed licenses
	if allowed, ok := config["allowed_licenses"].([]interface{}); ok {
		p.allowedLicenses = make([]string, len(allowed))
		for i, l := range allowed {
			p.allowedLicenses[i] = l.(string)
		}
	} else {
		p.allowedLicenses = []string{"MIT", "Apache-2.0", "BSD-3-Clause", "BSD-2-Clause", "ISC"}
	}

	// Parse denied licenses
	if denied, ok := config["denied_licenses"].([]interface{}); ok {
		p.deniedLicenses = make([]string, len(denied))
		for i, l := range denied {
			p.deniedLicenses[i] = l.(string)
		}
	} else {
		p.deniedLicenses = []string{"GPL-2.0", "GPL-3.0", "AGPL-3.0"}
	}

	return nil
}

func (p *LicenseCompliancePlugin) Execute(ctx *sdk.ExecutionContext) (*sdk.Result, error) {
	ctx.Logger.Info("Starting license compliance scan")

	licenses := make([]License, 0)

	// Scan different package managers
	npmLicenses, err := p.scanNPM(ctx)
	if err == nil {
		licenses = append(licenses, npmLicenses...)
	}

	mavenLicenses, err := p.scanMaven(ctx)
	if err == nil {
		licenses = append(licenses, mavenLicenses...)
	}

	goLicenses, err := p.scanGo(ctx)
	if err == nil {
		licenses = append(licenses, goLicenses...)
	}

	if len(licenses) == 0 {
		return &sdk.Result{
			Success:      false,
			ErrorMessage: "No licenses found",
		}, fmt.Errorf("no dependencies found")
	}

	ctx.Logger.Info(fmt.Sprintf("Found %d dependencies", len(licenses)))

	// Validate licenses
	deniedCount := 0
	unknownCount := 0
	approvedCount := 0

	for i := range licenses {
		license := &licenses[i]

		// Check if denied
		if p.isDenied(license.License) {
			license.Approved = false
			deniedCount++
			continue
		}

		// Check if allowed
		if p.isAllowed(license.License) {
			license.Approved = true
			approvedCount++
			continue
		}

		// Unknown license
		unknownCount++
		license.Approved = false
	}

	// Generate SBOM if requested
	if p.generateSBOM {
		sbomPath := filepath.Join(ctx.WorkDir, "sbom.json")
		if err := p.generateSBOMFile(licenses, sbomPath); err != nil {
			ctx.Logger.Error(fmt.Sprintf("Failed to generate SBOM: %v", err))
		} else {
			ctx.Logger.Info(fmt.Sprintf("SBOM generated: %s", sbomPath))
		}
	}

	// Build result
	result := &sdk.Result{
		Success:  deniedCount == 0 && (unknownCount == 0 || !p.failOnUnknown),
		ExitCode: 0,
		Metadata: make(map[string]interface{}),
		Output:   fmt.Sprintf("Scanned %d dependencies: %d approved, %d denied, %d unknown", len(licenses), approvedCount, deniedCount, unknownCount),
	}

	if deniedCount > 0 && p.failOnDenied {
		result.ExitCode = 1
		result.ErrorMessage = fmt.Sprintf("Found %d denied licenses", deniedCount)
	} else if unknownCount > 0 && p.failOnUnknown {
		result.ExitCode = 1
		result.ErrorMessage = fmt.Sprintf("Found %d unknown licenses", unknownCount)
	}

	result.Metadata["total_dependencies"] = len(licenses)
	result.Metadata["approved_count"] = approvedCount
	result.Metadata["denied_count"] = deniedCount
	result.Metadata["unknown_count"] = unknownCount

	ctx.Logger.Info(result.Output)

	return result, nil
}

func (p *LicenseCompliancePlugin) scanNPM(ctx *sdk.ExecutionContext) ([]License, error) {
	packageJSON := filepath.Join(ctx.WorkDir, p.scanPath, "package.json")
	if _, err := os.Stat(packageJSON); os.IsNotExist(err) {
		return nil, fmt.Errorf("no package.json found")
	}

	cmd := exec.Command("npm", "list", "--json", "--all")
	cmd.Dir = filepath.Join(ctx.WorkDir, p.scanPath)
	output, err := cmd.Output()
	if err != nil {
		// npm list returns non-zero even on success sometimes
		if len(output) == 0 {
			return nil, err
		}
	}

	var npmList struct {
		Dependencies map[string]struct {
			Version string `json:"version"`
			License string `json:"license,omitempty"`
		} `json:"dependencies"`
	}

	if err := json.Unmarshal(output, &npmList); err != nil {
		return nil, err
	}

	licenses := make([]License, 0)
	for name, dep := range npmList.Dependencies {
		licenses = append(licenses, License{
			Name:    name,
			Package: name,
			Version: dep.Version,
			License: dep.License,
		})
	}

	return licenses, nil
}

func (p *LicenseCompliancePlugin) scanMaven(ctx *sdk.ExecutionContext) ([]License, error) {
	pomXML := filepath.Join(ctx.WorkDir, p.scanPath, "pom.xml")
	if _, err := os.Stat(pomXML); os.IsNotExist(err) {
		return nil, fmt.Errorf("no pom.xml found")
	}

	// Use maven license plugin
	cmd := exec.Command("mvn", "license:aggregate-third-party-report", "-DoutputDirectory=target")
	cmd.Dir = filepath.Join(ctx.WorkDir, p.scanPath)
	if err := cmd.Run(); err != nil {
		return nil, err
	}

	// Parse would go here - simplified for now
	return []License{}, nil
}

func (p *LicenseCompliancePlugin) scanGo(ctx *sdk.ExecutionContext) ([]License, error) {
	goMod := filepath.Join(ctx.WorkDir, p.scanPath, "go.mod")
	if _, err := os.Stat(goMod); os.IsNotExist(err) {
		return nil, fmt.Errorf("no go.mod found")
	}

	cmd := exec.Command("go", "list", "-m", "-json", "all")
	cmd.Dir = filepath.Join(ctx.WorkDir, p.scanPath)
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(output), "\n")
	licenses := make([]License, 0)

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		var mod struct {
			Path    string `json:"Path"`
			Version string `json:"Version"`
		}

		if err := json.Unmarshal([]byte(line), &mod); err == nil {
			licenses = append(licenses, License{
				Name:    mod.Path,
				Package: mod.Path,
				Version: mod.Version,
				License: "UNKNOWN", // Would need additional lookup
			})
		}
	}

	return licenses, nil
}

func (p *LicenseCompliancePlugin) isAllowed(license string) bool {
	for _, allowed := range p.allowedLicenses {
		if strings.EqualFold(license, allowed) {
			return true
		}
	}
	return false
}

func (p *LicenseCompliancePlugin) isDenied(license string) bool {
	for _, denied := range p.deniedLicenses {
		if strings.EqualFold(license, denied) {
			return true
		}
	}
	return false
}

func (p *LicenseCompliancePlugin) generateSBOMFile(licenses []License, path string) error {
	data, err := json.MarshalIndent(licenses, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func (p *LicenseCompliancePlugin) Cleanup() error {
	return nil
}

// Helper functions
func getStringConfig(config map[string]interface{}, key, defaultValue string) string {
	if val, ok := config[key].(string); ok {
		return val
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
var Plugin LicenseCompliancePlugin

func main() {
	fmt.Println("License Compliance Plugin v1.0.0")
	fmt.Println("This plugin scans and validates software licenses")
	fmt.Println("To build: go build -o license-compliance")
}
