package tester

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// Manager handles tester agent lifecycle and test execution.
type Manager struct {
	// WorkDir is the working directory for test execution.
	WorkDir string

	// OutputDir is the directory for test artifacts.
	OutputDir string

	// Config is the default tester configuration.
	Config *Config
}

// Config holds the tester configuration.
type Config struct {
	// Playwright is the Playwright-specific configuration.
	Playwright *PlaywrightConfig `json:"playwright,omitempty"`

	// Recording is the artifact recording configuration.
	Recording *RecordingConfig `json:"recording,omitempty"`

	// WaitStrategy is the default wait strategy.
	WaitStrategy *WaitStrategy `json:"wait_strategy,omitempty"`

	// Retry is the default retry configuration.
	Retry *RetryConfig `json:"retry,omitempty"`

	// TestData is the test data isolation configuration.
	TestData *TestDataConfig `json:"test_data,omitempty"`

	// DefaultModel is the default Claude model for testing.
	// Options: haiku, sonnet, gemini
	DefaultModel string `json:"default_model,omitempty"`

	// DefaultTimeout is the default test timeout in seconds.
	DefaultTimeout int `json:"default_timeout,omitempty"`

	// Environment is the target environment (staging, production).
	Environment string `json:"environment,omitempty"`
}

// NewManager creates a new tester manager.
func NewManager(workDir, outputDir string) *Manager {
	return &Manager{
		WorkDir:   workDir,
		OutputDir: outputDir,
		Config:    DefaultConfig(),
	}
}

// DefaultConfig returns sensible defaults for testing.
func DefaultConfig() *Config {
	return &Config{
		Playwright: &PlaywrightConfig{
			Headless: true,
			Browser:  "chromium",
			Timeout:  30000,
		},
		Recording: &RecordingConfig{
			Video:        true,
			VideoFormat:  "webm",
			VideoQuality: "medium",
			Trace:        true,
			Screenshots: &ScreenshotConfig{
				OnFailure:     true,
				OnConfusion:   true,
				RequireReview: true,
				KeyMoments:    true,
			},
		},
		WaitStrategy: &WaitStrategy{
			NetworkIdle:       true,
			AnimationComplete: true,
			MinLoadTime:       1500,
		},
		Retry: &RetryConfig{
			MaxAttempts: 3,
			OnErrors:    []string{"browser_crash", "timeout", "network_error"},
			NotOn:       []string{"test_failure", "blocked"},
			Backoff:     "exponential",
			BackoffBase: 1000,
		},
		DefaultModel:   "haiku",
		DefaultTimeout: 600,
		Environment:    "staging",
	}
}

// EnsureOutputDir creates the output directory structure for a test run.
func (m *Manager) EnsureOutputDir(scenario string, runID string) (string, error) {
	date := time.Now().Format("2006-01-02")
	runDir := filepath.Join(m.OutputDir, date, scenario, runID)

	dirs := []string{
		runDir,
		filepath.Join(runDir, "screenshots"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return "", fmt.Errorf("creating output directory %s: %w", dir, err)
		}
	}

	return runDir, nil
}

// CheckPlaywrightInstalled verifies Playwright is available.
func (m *Manager) CheckPlaywrightInstalled() error {
	cmd := exec.Command("npx", "@playwright/mcp@latest", "--version")
	if err := cmd.Run(); err != nil {
		// Try alternative check - just verify npx is available
		cmd = exec.Command("npx", "--version")
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("npx not available: %w", err)
		}
		// npx is available, so Playwright MCP can be installed on demand
	}
	return nil
}

// CheckEnvironmentReachable verifies the target environment is accessible.
func (m *Manager) CheckEnvironmentReachable(url string) error {
	// Simple check - just verify we can reach the URL
	cmd := exec.Command("curl", "-s", "-o", "/dev/null", "-w", "%{http_code}", url)
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("cannot reach %s: %w", url, err)
	}

	statusCode := strings.TrimSpace(string(output))
	if statusCode == "000" {
		return fmt.Errorf("cannot connect to %s", url)
	}

	return nil
}

// RunPreflight performs preflight checks before testing.
func (m *Manager) RunPreflight(targetURL string) (*PreflightResult, error) {
	result := &PreflightResult{
		Passed: true,
		Checks: []PreflightCheck{},
	}

	// Check 1: Playwright available
	check := PreflightCheck{Name: "Playwright MCP"}
	if err := m.CheckPlaywrightInstalled(); err != nil {
		check.Passed = false
		check.Message = "Not available"
		check.Details = err.Error()
		result.Passed = false
		result.Errors = append(result.Errors, "Playwright MCP not available")
	} else {
		check.Passed = true
		check.Message = "Available via npx"
	}
	result.Checks = append(result.Checks, check)

	// Check 2: Environment reachable
	if targetURL != "" {
		check = PreflightCheck{Name: "Target Environment"}
		if err := m.CheckEnvironmentReachable(targetURL); err != nil {
			check.Passed = false
			check.Message = "Unreachable"
			check.Details = err.Error()
			result.Passed = false
			result.Errors = append(result.Errors, fmt.Sprintf("Cannot reach %s", targetURL))
		} else {
			check.Passed = true
			check.Message = fmt.Sprintf("Reachable (%s)", targetURL)
		}
		result.Checks = append(result.Checks, check)
	}

	// Check 3: Disk space
	check = PreflightCheck{Name: "Disk Space"}
	freeGB, err := m.CheckDiskSpace(m.OutputDir)
	if err != nil {
		check.Passed = true // Non-blocking
		check.Message = "Could not check"
		check.Details = err.Error()
		result.Warnings = append(result.Warnings, "Could not verify disk space")
	} else if freeGB < 5 {
		check.Passed = false
		check.Message = fmt.Sprintf("Low: %.1f GB free", freeGB)
		result.Warnings = append(result.Warnings, fmt.Sprintf("Low disk space: %.1f GB", freeGB))
	} else {
		check.Passed = true
		check.Message = fmt.Sprintf("%.1f GB free", freeGB)
	}
	result.Checks = append(result.Checks, check)

	return result, nil
}

// CheckDiskSpace returns free disk space in GB.
func (m *Manager) CheckDiskSpace(path string) (float64, error) {
	// Use df to check disk space
	cmd := exec.Command("df", "-BG", path)
	output, err := cmd.Output()
	if err != nil {
		return 0, err
	}

	// Parse df output (second line, fourth column for available space)
	lines := strings.Split(string(output), "\n")
	if len(lines) < 2 {
		return 0, fmt.Errorf("unexpected df output")
	}

	fields := strings.Fields(lines[1])
	if len(fields) < 4 {
		return 0, fmt.Errorf("unexpected df output format")
	}

	// Remove 'G' suffix and parse
	availStr := strings.TrimSuffix(fields[3], "G")
	var avail float64
	if _, err := fmt.Sscanf(availStr, "%f", &avail); err != nil {
		return 0, fmt.Errorf("parsing available space: %w", err)
	}

	return avail, nil
}

// GenerateRunID creates a unique run identifier.
func GenerateRunID() string {
	return fmt.Sprintf("run-%03d", time.Now().UnixNano()%1000)
}

// SaveTestResult saves a test result to the output directory.
func (m *Manager) SaveTestResult(runDir string, result *TestResult) error {
	// Save observations.json
	observationsPath := filepath.Join(runDir, "observations.json")
	observationsData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling result: %w", err)
	}
	if err := os.WriteFile(observationsPath, observationsData, 0644); err != nil {
		return fmt.Errorf("writing observations: %w", err)
	}
	result.Artifacts.ObservationsPath = observationsPath

	// Generate summary.md
	summaryPath := filepath.Join(runDir, "summary.md")
	summary := m.GenerateSummary(result)
	if err := os.WriteFile(summaryPath, []byte(summary), 0644); err != nil {
		return fmt.Errorf("writing summary: %w", err)
	}
	result.Artifacts.SummaryPath = summaryPath

	return nil
}

// GenerateSummary creates a markdown summary of the test result.
func (m *Manager) GenerateSummary(result *TestResult) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("# Test Summary: %s\n\n", result.Scenario))
	sb.WriteString(fmt.Sprintf("**Persona**: %s\n", result.Persona))
	sb.WriteString(fmt.Sprintf("**Status**: %s\n", result.Status))
	sb.WriteString(fmt.Sprintf("**Duration**: %ds\n", result.DurationSeconds))
	sb.WriteString(fmt.Sprintf("**Retries**: %d\n\n", result.RetryCount))

	sb.WriteString("## Success Criteria\n\n")
	for _, c := range result.SuccessCriteriaMet {
		sb.WriteString(fmt.Sprintf("- [x] %s\n", c))
	}
	for _, c := range result.SuccessCriteriaFailed {
		sb.WriteString(fmt.Sprintf("- [ ] %s\n", c))
	}
	sb.WriteString("\n")

	if len(result.Observations) > 0 {
		sb.WriteString("## Observations\n\n")
		for _, obs := range result.Observations {
			sb.WriteString(fmt.Sprintf("### %s/%s: %s\n\n", obs.Severity, obs.Confidence, obs.Type))
			sb.WriteString(fmt.Sprintf("- **Location**: %s\n", obs.Location))
			sb.WriteString(fmt.Sprintf("- **Timestamp**: %s\n", obs.Timestamp))
			sb.WriteString(fmt.Sprintf("- **Description**: %s\n", obs.Description))
			if obs.Screenshot != "" {
				sb.WriteString(fmt.Sprintf("- **Screenshot**: %s\n", obs.Screenshot))
			}
			sb.WriteString("\n")
		}
	}

	if result.OverallExperience != "" {
		sb.WriteString("## Overall Experience\n\n")
		sb.WriteString(result.OverallExperience + "\n\n")
	}

	if len(result.InfrastructureErrors) > 0 {
		sb.WriteString("## Infrastructure Errors\n\n")
		for _, e := range result.InfrastructureErrors {
			sb.WriteString(fmt.Sprintf("- %s\n", e))
		}
		sb.WriteString("\n")
	}

	if result.Artifacts != nil {
		sb.WriteString("## Artifacts\n\n")
		if result.Artifacts.VideoPath != "" {
			sb.WriteString(fmt.Sprintf("- Video: %s\n", result.Artifacts.VideoPath))
		}
		if result.Artifacts.TracePath != "" {
			sb.WriteString(fmt.Sprintf("- Trace: %s\n", result.Artifacts.TracePath))
		}
		if result.Artifacts.ScreenshotsDir != "" {
			sb.WriteString(fmt.Sprintf("- Screenshots: %s\n", result.Artifacts.ScreenshotsDir))
		}
	}

	return sb.String()
}

// CalculateBackoff calculates retry delay based on attempt number and config.
func CalculateBackoff(attempt int, config *RetryConfig) time.Duration {
	base := config.BackoffBase
	if base == 0 {
		base = 1000
	}

	switch config.Backoff {
	case "exponential":
		// 1s, 2s, 4s, 8s, ...
		return time.Duration(base*(1<<(attempt-1))) * time.Millisecond
	case "linear":
		// 1s, 2s, 3s, 4s, ...
		return time.Duration(base*attempt) * time.Millisecond
	default:
		// constant
		return time.Duration(base) * time.Millisecond
	}
}

// ShouldRetry determines if an error should trigger a retry.
func ShouldRetry(errorType string, config *RetryConfig) bool {
	// Check if error is in the not-on list
	for _, e := range config.NotOn {
		if e == errorType {
			return false
		}
	}

	// Check if error is in the on-list (or if on-list is empty, retry by default)
	if len(config.OnErrors) == 0 {
		return true
	}

	for _, e := range config.OnErrors {
		if e == errorType {
			return true
		}
	}

	return false
}
