// Package tester provides AI User Testing capabilities using Playwright MCP.
// The tester spawns Claude agents that embody user personas to navigate
// applications and identify UX issues.
package tester

import (
	"time"
)

// MCPServerConfig defines the configuration for an MCP server.
type MCPServerConfig struct {
	// Command is the command to run the MCP server.
	Command string `json:"command"`

	// Args are the arguments to pass to the command.
	Args []string `json:"args,omitempty"`

	// Env are environment variables for the MCP server.
	Env map[string]string `json:"env,omitempty"`
}

// PlaywrightConfig defines Playwright-specific settings for the tester.
type PlaywrightConfig struct {
	// Headless controls whether to run in headless mode.
	// Default: true for CI, false for interactive testing.
	Headless bool `json:"headless"`

	// Headed is the inverse of headless, for explicit control.
	Headed bool `json:"headed"`

	// Browser specifies which browser to use: chromium, firefox, webkit.
	// Default: chromium
	Browser string `json:"browser,omitempty"`

	// Viewport sets the browser viewport size.
	Viewport *Viewport `json:"viewport,omitempty"`

	// SlowMo adds delay between actions (milliseconds).
	// Useful for debugging.
	SlowMo int `json:"slow_mo,omitempty"`

	// Timeout is the default timeout for operations (milliseconds).
	// Default: 30000 (30 seconds)
	Timeout int `json:"timeout,omitempty"`
}

// Viewport defines browser window dimensions.
type Viewport struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

// RecordingConfig defines artifact recording settings.
type RecordingConfig struct {
	// Video controls video recording.
	Video bool `json:"video"`

	// VideoFormat specifies video format: webm, mp4.
	// Default: webm
	VideoFormat string `json:"video_format,omitempty"`

	// VideoQuality specifies video quality: low, medium, high.
	// Default: medium
	VideoQuality string `json:"video_quality,omitempty"`

	// Trace controls Playwright trace recording.
	Trace bool `json:"trace"`

	// Screenshots controls screenshot capture settings.
	Screenshots *ScreenshotConfig `json:"screenshots,omitempty"`
}

// ScreenshotConfig defines when to capture screenshots.
type ScreenshotConfig struct {
	// OnFailure captures screenshot when test fails.
	OnFailure bool `json:"on_failure"`

	// OnConfusion captures screenshot when agent reports confusion.
	OnConfusion bool `json:"on_confusion"`

	// RequireReview marks screenshots for human review.
	RequireReview bool `json:"require_review"`

	// KeyMoments captures screenshots at key navigation points.
	KeyMoments bool `json:"key_moments"`
}

// WaitStrategy defines how to wait for page stability.
type WaitStrategy struct {
	// NetworkIdle waits for no network requests for 500ms.
	NetworkIdle bool `json:"network_idle"`

	// AnimationComplete waits for CSS animations to finish.
	AnimationComplete bool `json:"animation_complete"`

	// MinLoadTime is the minimum time to wait after navigation (ms).
	MinLoadTime int `json:"min_load_time,omitempty"`

	// CustomSelectors are app-specific ready indicators.
	CustomSelectors []string `json:"custom_selectors,omitempty"`
}

// RetryConfig defines retry behavior for infrastructure failures.
type RetryConfig struct {
	// MaxAttempts is the maximum number of retry attempts.
	// Default: 3
	MaxAttempts int `json:"max_attempts"`

	// OnErrors lists error types that should trigger retry.
	// Common: browser_crash, timeout, network_error
	OnErrors []string `json:"on_errors,omitempty"`

	// NotOn lists error types that should NOT trigger retry.
	// Common: test_failure, blocked
	NotOn []string `json:"not_on,omitempty"`

	// Backoff specifies backoff strategy: exponential, linear, constant.
	// Default: exponential
	Backoff string `json:"backoff,omitempty"`

	// BackoffBase is the base delay in milliseconds.
	// Default: 1000
	BackoffBase int `json:"backoff_base,omitempty"`
}

// TestDataConfig defines test data isolation settings.
type TestDataConfig struct {
	// EmailPattern is the pattern for generating test emails.
	// Supports: {scenario}, {run_id}, {timestamp}
	// Example: "test+{scenario}+{run_id}@screencoach.test"
	EmailPattern string `json:"email_pattern,omitempty"`

	// EmailInbox specifies the test email service: mailhog, skip_verification.
	EmailInbox string `json:"email_inbox,omitempty"`

	// CleanupStrategy defines cleanup behavior.
	CleanupStrategy *CleanupStrategy `json:"cleanup_strategy,omitempty"`

	// Isolation settings.
	Isolation *IsolationConfig `json:"isolation,omitempty"`
}

// CleanupStrategy defines when and how to clean up test data.
type CleanupStrategy struct {
	// OnSuccess action: delete_account, keep, mark_complete.
	OnSuccess string `json:"on_success,omitempty"`

	// OnFailure action: mark_for_review, delete_account, keep.
	OnFailure string `json:"on_failure,omitempty"`

	// OnCrash action: cleanup_job, keep.
	OnCrash string `json:"on_crash,omitempty"`
}

// IsolationConfig defines data isolation settings.
type IsolationConfig struct {
	// UniqueSuffix appends UUID to all created data.
	UniqueSuffix bool `json:"unique_suffix"`
}

// Observation represents a UX observation from the test.
type Observation struct {
	// Type is the observation category.
	// Values: confusion, friction, error, success, suggestion
	Type string `json:"type"`

	// Severity is the priority level: P0, P1, P2, P3.
	Severity string `json:"severity"`

	// Confidence is the agent's self-assessment: high, medium, low.
	Confidence string `json:"confidence"`

	// Timestamp is when the observation occurred (mm:ss format).
	Timestamp string `json:"timestamp"`

	// Location is where in the app the observation occurred.
	Location string `json:"location"`

	// Description is the observation detail.
	Description string `json:"description"`

	// Screenshot is the filename of the associated screenshot.
	Screenshot string `json:"screenshot,omitempty"`

	// Validated indicates human validation status (null if pending).
	Validated *bool `json:"validated,omitempty"`

	// FalsePositive indicates if marked as false positive.
	FalsePositive *bool `json:"false_positive,omitempty"`
}

// TestResult represents the outcome of a test run.
type TestResult struct {
	// Scenario is the scenario name.
	Scenario string `json:"scenario"`

	// Persona is the persona name used.
	Persona string `json:"persona"`

	// Completed indicates if the test finished without errors.
	Completed bool `json:"completed"`

	// Status is the overall result: passed, failed, error, blocked.
	Status string `json:"status"`

	// DurationSeconds is the test duration.
	DurationSeconds int `json:"duration_seconds"`

	// Observations are the UX observations recorded.
	Observations []Observation `json:"observations"`

	// SuccessCriteriaMet lists criteria that passed.
	SuccessCriteriaMet []string `json:"success_criteria_met"`

	// SuccessCriteriaFailed lists criteria that failed.
	SuccessCriteriaFailed []string `json:"success_criteria_failed"`

	// OverallExperience is a summary of the user experience.
	OverallExperience string `json:"overall_experience"`

	// RetryCount is the number of retries attempted.
	RetryCount int `json:"retry_count"`

	// InfrastructureErrors lists any infrastructure issues.
	InfrastructureErrors []string `json:"infrastructure_errors"`

	// Artifacts contains paths to recorded artifacts.
	Artifacts *Artifacts `json:"artifacts,omitempty"`

	// StartedAt is when the test started.
	StartedAt time.Time `json:"started_at"`

	// CompletedAt is when the test completed.
	CompletedAt time.Time `json:"completed_at"`
}

// Artifacts contains paths to test artifacts.
type Artifacts struct {
	// VideoPath is the path to the recorded video.
	VideoPath string `json:"video_path,omitempty"`

	// TracePath is the path to the Playwright trace archive.
	TracePath string `json:"trace_path,omitempty"`

	// ScreenshotsDir is the directory containing screenshots.
	ScreenshotsDir string `json:"screenshots_dir,omitempty"`

	// ObservationsPath is the path to observations.json.
	ObservationsPath string `json:"observations_path,omitempty"`

	// SummaryPath is the path to summary.md.
	SummaryPath string `json:"summary_path,omitempty"`
}

// PreflightResult represents the result of preflight checks.
type PreflightResult struct {
	// Passed indicates all checks passed.
	Passed bool `json:"passed"`

	// Checks lists individual check results.
	Checks []PreflightCheck `json:"checks"`

	// Warnings lists non-blocking issues.
	Warnings []string `json:"warnings,omitempty"`

	// Errors lists blocking issues.
	Errors []string `json:"errors,omitempty"`
}

// PreflightCheck represents a single preflight check.
type PreflightCheck struct {
	// Name is the check name.
	Name string `json:"name"`

	// Passed indicates if the check passed.
	Passed bool `json:"passed"`

	// Message is the check result message.
	Message string `json:"message"`

	// Details provides additional information.
	Details string `json:"details,omitempty"`
}
