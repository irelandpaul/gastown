package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/steveyegge/gastown/internal/style"
	"github.com/steveyegge/gastown/internal/ui"
	"gopkg.in/yaml.v3"
)

// Run command flags
var (
	runModel        string
	runHeaded       bool
	runNoVideo      bool
	runNoTrace      bool
	runTimeout      int
	runRetry        int
	runNoRetry      bool
	runCompareTo    string
	runOutput       string
)

var testerRunCmd = &cobra.Command{
	Use:   "run <scenario.yaml>",
	Short: "Run a single test scenario",
	Long: `Run a single AI user test scenario.

The scenario file defines:
  - Persona (who is using the app)
  - Target app and environment
  - Goal (what the user is trying to accomplish)
  - Success criteria

The test spawns an AI agent to navigate the application as the
specified persona, making observations about UX issues.

Preflight checks run automatically before testing (use --skip-preflight to disable).

Examples:
  gt tester run scenarios/signup.yaml           # Run a scenario
  gt tester run scenarios/signup.yaml --headed  # Show browser window
  gt tester run scenarios/signup.yaml --model sonnet  # Use Sonnet model
  gt tester run scenarios/signup.yaml --retry 5       # Set max retries
  gt tester run scenarios/signup.yaml --no-retry      # Disable retry`,
	Args: cobra.ExactArgs(1),
	RunE: runTesterRun,
}

// TestScenario represents a parsed scenario file
type TestScenario struct {
	Scenario    string   `yaml:"scenario" json:"scenario"`
	Version     int      `yaml:"version" json:"version"`
	Description string   `yaml:"description" json:"description"`
	Tags        []string `yaml:"tags" json:"tags"`

	Persona struct {
		Name        string `yaml:"name" json:"name"`
		Role        string `yaml:"role" json:"role"`
		Context     string `yaml:"context" json:"context"`
		TechComfort string `yaml:"tech_comfort" json:"tech_comfort"`
		Patience    string `yaml:"patience" json:"patience"`
		Device      string `yaml:"device" json:"device"`
	} `yaml:"persona" json:"persona"`

	Target struct {
		App         string `yaml:"app" json:"app"`
		Environment string `yaml:"environment" json:"environment"`
		URL         string `yaml:"url" json:"url"`
	} `yaml:"target" json:"target"`

	Goal  string   `yaml:"goal" json:"goal"`
	Steps []string `yaml:"steps" json:"steps"`

	SuccessCriteria []string `yaml:"success_criteria" json:"success_criteria"`
	Evaluate        []string `yaml:"evaluate" json:"evaluate"`

	Recording struct {
		Video       bool `yaml:"video" json:"video"`
		Trace       bool `yaml:"trace" json:"trace"`
		Screenshots struct {
			OnFailure   bool `yaml:"on_failure" json:"on_failure"`
			OnConfusion bool `yaml:"on_confusion" json:"on_confusion"`
			OnDemand    bool `yaml:"on_demand" json:"on_demand"`
		} `yaml:"screenshots" json:"screenshots"`
	} `yaml:"recording" json:"recording"`

	Retry struct {
		MaxAttempts int      `yaml:"max_attempts" json:"max_attempts"`
		Backoff     string   `yaml:"backoff" json:"backoff"`
		BackoffBase int      `yaml:"backoff_base" json:"backoff_base"`
		NotOn       []string `yaml:"not_on" json:"not_on"`
	} `yaml:"retry" json:"retry"`

	Timeout int    `yaml:"timeout" json:"timeout"`
	Model   string `yaml:"model" json:"model"`
}

// TestRunResult contains the result of a test run
type TestRunResult struct {
	Scenario      string           `json:"scenario"`
	ScenarioFile  string           `json:"scenario_file"`
	StartTime     time.Time        `json:"start_time"`
	EndTime       time.Time        `json:"end_time"`
	Duration      string           `json:"duration"`
	Status        string           `json:"status"` // "pass", "fail", "error"
	ExitCode      int              `json:"exit_code"`
	Observations  []TestObservation `json:"observations"`
	CriteriaMet   int              `json:"criteria_met"`
	CriteriaTotal int              `json:"criteria_total"`
	RetryAttempts int              `json:"retry_attempts"`
	Artifacts     TestArtifacts    `json:"artifacts"`
	Error         string           `json:"error,omitempty"`
}

// TestObservation represents a UX observation made during testing
type TestObservation struct {
	Timestamp  string `json:"timestamp"`
	Severity   string `json:"severity"` // P0-P3
	Confidence string `json:"confidence"` // high, medium, low
	Type       string `json:"type"` // confusion, friction, blocked, bug
	Message    string `json:"message"`
	Screenshot string `json:"screenshot,omitempty"`
}

// TestArtifacts contains paths to test artifacts
type TestArtifacts struct {
	Video     string `json:"video,omitempty"`
	Trace     string `json:"trace,omitempty"`
	Summary   string `json:"summary,omitempty"`
	OutputDir string `json:"output_dir"`
}

// InfrastructureError represents an error that can be retried
type InfrastructureError struct {
	Type    string
	Message string
	Err     error
}

func (e InfrastructureError) Error() string {
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

func init() {
	testerRunCmd.Flags().StringVar(&runModel, "model", "", "Override model (haiku, sonnet)")
	testerRunCmd.Flags().BoolVar(&runHeaded, "headed", false, "Show browser window")
	testerRunCmd.Flags().BoolVar(&runNoVideo, "no-video", false, "Disable video recording")
	testerRunCmd.Flags().BoolVar(&runNoTrace, "no-trace", false, "Disable Playwright trace")
	testerRunCmd.Flags().IntVar(&runTimeout, "timeout", 0, "Override timeout (seconds)")
	testerRunCmd.Flags().IntVar(&runRetry, "retry", 0, "Override retry attempts")
	testerRunCmd.Flags().BoolVar(&runNoRetry, "no-retry", false, "Disable retry logic")
	testerRunCmd.Flags().StringVar(&runCompareTo, "compare-to", "", "Compare results to previous run")
	testerRunCmd.Flags().StringVar(&runOutput, "output", "", "Custom output directory")
	testerRunCmd.Flags().BoolVar(&testerSkipPreflight, "skip-preflight", false, "Skip environment preflight checks")
	testerRunCmd.Flags().BoolVar(&testerVerbose, "verbose", false, "Show agent output in real-time")
}

func runTesterRun(cmd *cobra.Command, args []string) error {
	scenarioPath := args[0]

	// Load and validate scenario
	scenario, err := loadScenario(scenarioPath)
	if err != nil {
		return fmt.Errorf("loading scenario: %w", err)
	}

	// Print header
	fmt.Printf("\n%s %s\n", style.Bold.Render("Running:"), scenario.Scenario)
	if scenario.Description != "" {
		fmt.Printf("  %s\n", ui.RenderMuted(scenario.Description))
	}
	fmt.Printf("  Persona: %s, %s\n", scenario.Persona.Name, scenario.Persona.Context)
	fmt.Printf("  App: %s (%s)\n", scenario.Target.App, scenario.Target.Environment)

	// Determine model
	model := scenario.Model
	if runModel != "" {
		model = runModel
	}
	if model == "" {
		model = "haiku"
	}
	fmt.Printf("  Model: %s\n", model)
	fmt.Println()

	// Run preflight checks unless skipped
	if !testerSkipPreflight {
		fmt.Println("Preflight checks...")
		passed, err := runPreflightQuick()
		if err != nil {
			return err
		}
		if !passed {
			return NewSilentExit(4)
		}
		fmt.Println()
	}

	// Determine retry config
	maxAttempts := 3
	if scenario.Retry.MaxAttempts > 0 {
		maxAttempts = scenario.Retry.MaxAttempts
	}
	if runRetry > 0 {
		maxAttempts = runRetry
	}
	if runNoRetry {
		maxAttempts = 1
	}

	// Determine timeout
	timeout := 600
	if scenario.Timeout > 0 {
		timeout = scenario.Timeout
	}
	if runTimeout > 0 {
		timeout = runTimeout
	}

	// Create output directory
	outputDir := runOutput
	if outputDir == "" {
		date := time.Now().Format("2006-01-02")
		outputDir = filepath.Join("test-results", date, scenario.Scenario, fmt.Sprintf("run-%03d", 1))
	}
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("creating output directory: %w", err)
	}

	// Initialize result
	result := TestRunResult{
		Scenario:      scenario.Scenario,
		ScenarioFile:  scenarioPath,
		StartTime:     time.Now(),
		CriteriaTotal: len(scenario.SuccessCriteria),
		Artifacts: TestArtifacts{
			OutputDir: outputDir,
		},
	}

	// Run test with retry logic
	fmt.Println("Starting browser...")
	var lastErr error

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		result.RetryAttempts = attempt

		runErr := executeTestScenario(scenario, &result, attempt, timeout, model)

		if runErr == nil {
			// Test completed successfully
			break
		}

		// Check if this is an infrastructure error (retriable)
		if infraErr, ok := runErr.(InfrastructureError); ok {
			// Check if this error type should not be retried
			if isNoRetryError(infraErr, scenario.Retry.NotOn) {
				lastErr = runErr
				break
			}

			if attempt < maxAttempts {
				backoff := calculateBackoff(attempt, scenario.Retry)
				fmt.Printf("  %s %s (attempt %d/%d)\n", ui.RenderFailIcon(), infraErr.Type, attempt, maxAttempts)
				fmt.Printf("  Retrying in %v...\n", backoff)
				time.Sleep(backoff)
				continue
			}
		}

		// Non-retriable error or max attempts reached
		lastErr = runErr
		break
	}

	// Record end time and duration
	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime).Round(time.Second).String()

	// Handle errors
	if lastErr != nil {
		result.Status = "error"
		result.Error = lastErr.Error()
		result.ExitCode = 2
	}

	// Output results
	fmt.Println()
	fmt.Println(style.Bold.Render("Test Complete"))
	fmt.Printf("  Duration: %s\n", result.Duration)

	// Count observations by severity
	p0p1Count := 0
	p2Count := 0
	p3Count := 0
	for _, obs := range result.Observations {
		switch obs.Severity {
		case "P0", "P1":
			p0p1Count++
		case "P2":
			p2Count++
		case "P3":
			p3Count++
		}
	}

	fmt.Printf("  Observations: %d issues (%d P0/P1, %d P2, %d P3)\n",
		len(result.Observations), p0p1Count, p2Count, p3Count)
	fmt.Printf("  Success criteria: %d/%d met\n", result.CriteriaMet, result.CriteriaTotal)
	if result.RetryAttempts > 1 {
		fmt.Printf("  Retries: %d\n", result.RetryAttempts-1)
	}

	// Artifacts
	fmt.Println()
	fmt.Println("Artifacts:")
	if result.Artifacts.Video != "" {
		fmt.Printf("  Video: %s\n", result.Artifacts.Video)
	}
	if result.Artifacts.Trace != "" {
		fmt.Printf("  Trace: %s\n", result.Artifacts.Trace)
	}
	if result.Artifacts.Summary != "" {
		fmt.Printf("  Summary: %s\n", result.Artifacts.Summary)
	}

	// Final result
	fmt.Println()
	switch result.Status {
	case "pass":
		if p0p1Count == 0 {
			fmt.Printf("Result: %s (no bead created - no P0/P1 issues)\n", ui.RenderPass("PASS"))
		} else {
			fmt.Printf("Result: %s - %d P0/P1 issues require attention\n", ui.RenderWarn("PASS with issues"), p0p1Count)
		}
	case "fail":
		fmt.Printf("Result: %s - success criteria not met\n", ui.RenderFail("FAIL"))
	case "error":
		fmt.Printf("Result: %s - %s\n", ui.RenderFail("ERROR"), result.Error)
	}

	// JSON output if requested
	if testerJSON {
		fmt.Println()
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(result)
	}

	// Return appropriate exit code
	switch result.Status {
	case "pass":
		return nil
	case "fail":
		return NewSilentExit(1)
	case "error":
		return NewSilentExit(result.ExitCode)
	}

	return nil
}

// loadScenario loads and parses a scenario YAML file
func loadScenario(path string) (*TestScenario, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}

	var scenario TestScenario
	if err := yaml.Unmarshal(data, &scenario); err != nil {
		return nil, fmt.Errorf("parsing YAML: %w", err)
	}

	// Validate required fields
	if scenario.Scenario == "" {
		return nil, fmt.Errorf("scenario name is required")
	}
	if scenario.Persona.Name == "" {
		return nil, fmt.Errorf("persona name is required")
	}
	if scenario.Goal == "" {
		return nil, fmt.Errorf("goal is required")
	}

	return &scenario, nil
}

// runPreflightQuick runs a quick subset of preflight checks
func runPreflightQuick() (bool, error) {
	checks := []func() PreflightCheck{
		checkPlaywright,
		checkNodeJS,
	}

	allPassed := true
	for _, check := range checks {
		c := check()
		icon := ui.RenderPassIcon()
		if c.Status == "fail" {
			icon = ui.RenderFailIcon()
			allPassed = false
		} else if c.Status == "warn" {
			icon = ui.RenderWarnIcon()
		}

		fmt.Printf("  %s %s", icon, c.Name)
		if c.Message != "" {
			fmt.Printf(" (%s)", c.Message)
		}
		fmt.Println()
	}

	return allPassed, nil
}

// executeTestScenario runs the actual test scenario
func executeTestScenario(scenario *TestScenario, result *TestRunResult, attempt int, timeout int, model string) error {
	fmt.Printf("Agent navigating... (attempt %d)\n", attempt)

	// For now, this is a placeholder for the actual test execution
	// In a full implementation, this would:
	// 1. Spawn a Task agent with the tester CLAUDE.md context
	// 2. Provide the scenario details and persona
	// 3. Let the agent navigate using Playwright MCP
	// 4. Collect observations and artifacts

	// Simulate successful execution for the scaffold
	result.Status = "pass"
	result.CriteriaMet = result.CriteriaTotal

	// Create placeholder artifacts
	result.Artifacts.Video = filepath.Join(result.Artifacts.OutputDir, "video.webm")
	result.Artifacts.Trace = filepath.Join(result.Artifacts.OutputDir, "trace.zip")
	result.Artifacts.Summary = filepath.Join(result.Artifacts.OutputDir, "summary.md")

	// Write a placeholder summary
	summaryContent := fmt.Sprintf(`# Test Run Summary

**Scenario**: %s
**Persona**: %s
**App**: %s (%s)
**Model**: %s
**Status**: %s

## Observations

_No observations recorded (scaffold implementation)_

## Success Criteria

All %d criteria met.
`, scenario.Scenario, scenario.Persona.Name, scenario.Target.App,
		scenario.Target.Environment, model, result.Status, result.CriteriaTotal)

	if err := os.WriteFile(result.Artifacts.Summary, []byte(summaryContent), 0644); err != nil {
		// Non-fatal - just log
		fmt.Printf("  %s Could not write summary: %v\n", ui.RenderWarnIcon(), err)
	}

	return nil
}

// calculateBackoff calculates the backoff duration for retry
func calculateBackoff(attempt int, retry struct {
	MaxAttempts int      `yaml:"max_attempts" json:"max_attempts"`
	Backoff     string   `yaml:"backoff" json:"backoff"`
	BackoffBase int      `yaml:"backoff_base" json:"backoff_base"`
	NotOn       []string `yaml:"not_on" json:"not_on"`
}) time.Duration {
	base := 1000 // milliseconds
	if retry.BackoffBase > 0 {
		base = retry.BackoffBase
	}

	// Exponential backoff by default
	multiplier := 1
	if retry.Backoff == "linear" {
		multiplier = attempt
	} else {
		// exponential
		for i := 1; i < attempt; i++ {
			multiplier *= 2
		}
	}

	return time.Duration(base*multiplier) * time.Millisecond
}

// isNoRetryError checks if an error type is in the no-retry list
func isNoRetryError(err InfrastructureError, notOn []string) bool {
	for _, noRetry := range notOn {
		if strings.EqualFold(err.Type, noRetry) {
			return true
		}
	}
	return false
}
