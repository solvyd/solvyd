package main

import (
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"

	"github.com/vrenjith/ritmo/plugin-sdk/pkg/sdk"
)

// JUnitTestReporterPlugin processes JUnit XML test reports
type JUnitTestReporterPlugin struct {
	reportPath     string
	coverageMin    float64
	failOnError    bool
	includeSkipped bool
}

type TestSuites struct {
	XMLName    xml.Name    `xml:"testsuites"`
	TestSuites []TestSuite `xml:"testsuite"`
}

type TestSuite struct {
	XMLName   xml.Name   `xml:"testsuite"`
	Name      string     `xml:"name,attr"`
	Tests     int        `xml:"tests,attr"`
	Failures  int        `xml:"failures,attr"`
	Errors    int        `xml:"errors,attr"`
	Skipped   int        `xml:"skipped,attr"`
	Time      float64    `xml:"time,attr"`
	TestCases []TestCase `xml:"testcase"`
}

type TestCase struct {
	Name      string   `xml:"name,attr"`
	ClassName string   `xml:"classname,attr"`
	Time      float64  `xml:"time,attr"`
	Failure   *Failure `xml:"failure,omitempty"`
	Error     *Error   `xml:"error,omitempty"`
	Skipped   *Skipped `xml:"skipped,omitempty"`
}

type Failure struct {
	Message string `xml:"message,attr"`
	Type    string `xml:"type,attr"`
	Content string `xml:",chardata"`
}

type Error struct {
	Message string `xml:"message,attr"`
	Type    string `xml:"type,attr"`
	Content string `xml:",chardata"`
}

type Skipped struct {
	Message string `xml:"message,attr"`
}

func (p *JUnitTestReporterPlugin) Name() string {
	return "junit-test-reporter"
}

func (p *JUnitTestReporterPlugin) Version() string {
	return "1.0.0"
}

func (p *JUnitTestReporterPlugin) Type() string {
	return "test"
}

func (p *JUnitTestReporterPlugin) Initialize(config map[string]interface{}) error {
	p.reportPath = getStringConfig(config, "report_path", "**/test-results/**/*.xml")
	p.coverageMin = getFloatConfig(config, "coverage_min", 0.0)
	p.failOnError = getBoolConfig(config, "fail_on_error", true)
	p.includeSkipped = getBoolConfig(config, "include_skipped", false)

	return nil
}

func (p *JUnitTestReporterPlugin) Execute(ctx *sdk.ExecutionContext) (*sdk.Result, error) {
	ctx.Logger.Info("Processing JUnit test reports")

	// Find all test report files
	files, err := filepath.Glob(filepath.Join(ctx.WorkDir, p.reportPath))
	if err != nil {
		return &sdk.Result{
			Success:      false,
			ErrorMessage: fmt.Sprintf("Failed to find test reports: %v", err),
		}, err
	}

	if len(files) == 0 {
		return &sdk.Result{
			Success:      false,
			ErrorMessage: "No test report files found",
		}, fmt.Errorf("no test reports found at %s", p.reportPath)
	}

	ctx.Logger.Info(fmt.Sprintf("Found %d test report files", len(files)))

	// Parse all test reports
	totalTests := 0
	totalFailures := 0
	totalErrors := 0
	totalSkipped := 0
	totalTime := 0.0

	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			ctx.Logger.Error(fmt.Sprintf("Failed to read %s: %v", file, err))
			continue
		}

		var suites TestSuites
		if err := xml.Unmarshal(data, &suites); err != nil {
			// Try parsing as single suite
			var suite TestSuite
			if err := xml.Unmarshal(data, &suite); err != nil {
				ctx.Logger.Error(fmt.Sprintf("Failed to parse %s: %v", file, err))
				continue
			}
			suites.TestSuites = []TestSuite{suite}
		}

		// Aggregate results
		for _, suite := range suites.TestSuites {
			totalTests += suite.Tests
			totalFailures += suite.Failures
			totalErrors += suite.Errors
			totalSkipped += suite.Skipped
			totalTime += suite.Time
		}
	}

	// Calculate pass rate
	totalPassed := totalTests - totalFailures - totalErrors - totalSkipped
	passRate := 0.0
	if totalTests > 0 {
		passRate = float64(totalPassed) / float64(totalTests) * 100
	}

	// Build result
	result := &sdk.Result{
		Success:  totalFailures == 0 && totalErrors == 0,
		ExitCode: 0,
		Metadata: make(map[string]interface{}),
		Output:   fmt.Sprintf("Tests: %d, Passed: %d, Failed: %d, Errors: %d, Skipped: %d, Pass Rate: %.2f%%", totalTests, totalPassed, totalFailures, totalErrors, totalSkipped, passRate),
	}

	if (totalFailures > 0 || totalErrors > 0) && p.failOnError {
		result.ExitCode = 1
		result.ErrorMessage = fmt.Sprintf("%d tests failed, %d errors", totalFailures, totalErrors)
	}

	result.Metadata["total_tests"] = totalTests
	result.Metadata["passed"] = totalPassed
	result.Metadata["failures"] = totalFailures
	result.Metadata["errors"] = totalErrors
	result.Metadata["skipped"] = totalSkipped
	result.Metadata["pass_rate"] = passRate
	result.Metadata["total_time"] = totalTime

	ctx.Logger.Info(result.Output)

	return result, nil
}

func (p *JUnitTestReporterPlugin) Cleanup() error {
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

func getBoolConfig(config map[string]interface{}, key string, defaultValue bool) bool {
	if val, ok := config[key].(bool); ok {
		return val
	}
	return defaultValue
}

// Export the plugin
var Plugin JUnitTestReporterPlugin

func main() {
	fmt.Println("JUnit Test Reporter Plugin v1.0.0")
	fmt.Println("This plugin processes and reports JUnit XML test results")
	fmt.Println("To build: go build -o junit-test-reporter")
}
