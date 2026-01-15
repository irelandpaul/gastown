// Package batch provides batch execution for AI User Testing scenarios.
// It handles running multiple scenarios in parallel, aggregating results,
// and managing quarantined tests.
package batch

import (
	"time"
)

// RunStatus represents the outcome of a single scenario run.
type RunStatus string

const (
	// StatusPending means the scenario has not started yet.
	StatusPending RunStatus = "pending"

	// StatusRunning means the scenario is currently executing.
	StatusRunning RunStatus = "running"

	// StatusPassed means the scenario completed successfully.
	StatusPassed RunStatus = "passed"

	// StatusFailed means the scenario failed (criteria not met).
	StatusFailed RunStatus = "failed"

	// StatusError means the scenario had an error (crash, timeout).
	StatusError RunStatus = "error"

	// StatusSkipped means the scenario was skipped (quarantined).
	StatusSkipped RunStatus = "skipped"

	// StatusRetrying means the scenario is retrying after an error.
	StatusRetrying RunStatus = "retrying"
)

// Config defines the configuration for a batch run.
type Config struct {
	// Pattern is the glob pattern for scenario files.
	Pattern string `json:"pattern" yaml:"pattern"`

	// Parallel is the number of scenarios to run simultaneously.
	Parallel int `json:"parallel" yaml:"parallel"`

	// StopOnFail stops the batch on the first failure.
	StopOnFail bool `json:"stop_on_fail" yaml:"stop_on_fail"`

	// ConvoyName is the name for the convoy bead (optional).
	ConvoyName string `json:"convoy_name,omitempty" yaml:"convoy_name,omitempty"`

	// Model overrides the model for all scenarios.
	Model string `json:"model,omitempty" yaml:"model,omitempty"`

	// Environment is the target environment.
	Environment string `json:"environment" yaml:"environment"`

	// FilterTags includes only scenarios with these tags.
	FilterTags []string `json:"filter_tags,omitempty" yaml:"filter_tags,omitempty"`

	// ExcludeTags excludes scenarios with these tags.
	ExcludeTags []string `json:"exclude_tags,omitempty" yaml:"exclude_tags,omitempty"`

	// IncludeQuarantined includes quarantined tests.
	IncludeQuarantined bool `json:"include_quarantined" yaml:"include_quarantined"`

	// CompareTo is the previous batch run to compare against.
	CompareTo string `json:"compare_to,omitempty" yaml:"compare_to,omitempty"`

	// SkipPreflight skips the preflight checks.
	SkipPreflight bool `json:"skip_preflight" yaml:"skip_preflight"`

	// OutputDir is the output directory for results.
	OutputDir string `json:"output_dir" yaml:"output_dir"`
}

// DefaultConfig returns the default batch configuration.
func DefaultConfig() Config {
	return Config{
		Parallel:           1,
		StopOnFail:         false,
		Environment:        "staging",
		IncludeQuarantined: false,
		SkipPreflight:      false,
		OutputDir:          "test-results",
	}
}

// ScenarioResult holds the result of a single scenario execution.
type ScenarioResult struct {
	// Scenario is the scenario name.
	Scenario string `json:"scenario"`

	// Path is the path to the scenario file.
	Path string `json:"path"`

	// Status is the run status.
	Status RunStatus `json:"status"`

	// Duration is how long the scenario took.
	Duration time.Duration `json:"duration"`

	// Observations is the count of observations by severity.
	Observations map[string]int `json:"observations"`

	// SuccessCriteriaMet is the number of criteria met.
	SuccessCriteriaMet int `json:"success_criteria_met"`

	// SuccessCriteriaTotal is the total number of criteria.
	SuccessCriteriaTotal int `json:"success_criteria_total"`

	// RetryCount is how many retries were needed.
	RetryCount int `json:"retry_count"`

	// Error contains the error message if failed.
	Error string `json:"error,omitempty"`

	// ArtifactDir is the directory containing artifacts.
	ArtifactDir string `json:"artifact_dir,omitempty"`

	// Quarantined indicates if this scenario is quarantined.
	Quarantined bool `json:"quarantined"`

	// SkipReason explains why the scenario was skipped.
	SkipReason string `json:"skip_reason,omitempty"`
}

// BatchResult holds the aggregated results of a batch run.
type BatchResult struct {
	// ID is the unique identifier for this batch run.
	ID string `json:"id"`

	// Config is the batch configuration used.
	Config Config `json:"config"`

	// StartedAt is when the batch started.
	StartedAt time.Time `json:"started_at"`

	// CompletedAt is when the batch completed.
	CompletedAt *time.Time `json:"completed_at,omitempty"`

	// TotalDuration is the total elapsed time.
	TotalDuration time.Duration `json:"total_duration"`

	// ScenariosFound is the total number of scenarios found.
	ScenariosFound int `json:"scenarios_found"`

	// ScenariosRun is the number of scenarios actually run.
	ScenariosRun int `json:"scenarios_run"`

	// ScenariosSkipped is the number of scenarios skipped.
	ScenariosSkipped int `json:"scenarios_skipped"`

	// Results holds individual scenario results.
	Results []ScenarioResult `json:"results"`

	// Summary holds aggregated statistics.
	Summary BatchSummary `json:"summary"`

	// ConvoyID is the convoy bead ID if created.
	ConvoyID string `json:"convoy_id,omitempty"`

	// OutputDir is where results are stored.
	OutputDir string `json:"output_dir"`

	// Comparison holds the comparison to a baseline batch (if --compare-to was used).
	Comparison *Comparison `json:"comparison,omitempty"`
}

// BatchSummary holds aggregated statistics for a batch run.
type BatchSummary struct {
	// Passed is the count of passed scenarios.
	Passed int `json:"passed"`

	// Failed is the count of failed scenarios.
	Failed int `json:"failed"`

	// Errors is the count of errored scenarios.
	Errors int `json:"errors"`

	// Skipped is the count of skipped scenarios.
	Skipped int `json:"skipped"`

	// TotalObservations is the total observation count by severity.
	TotalObservations map[string]int `json:"total_observations"`

	// TotalRetries is the sum of all retries.
	TotalRetries int `json:"total_retries"`

	// FlakeRate is the calculated flake rate for this batch.
	FlakeRate float64 `json:"flake_rate"`

	// NewQuarantineCandidates are scenarios that might need quarantining.
	NewQuarantineCandidates []string `json:"new_quarantine_candidates,omitempty"`

	// AutoQuarantined are scenarios that were auto-quarantined during this batch.
	AutoQuarantined []string `json:"auto_quarantined,omitempty"`

	// AutoUnquarantined are scenarios that were auto-unquarantined during this batch.
	AutoUnquarantined []string `json:"auto_unquarantined,omitempty"`

	// FlakyScenarios are scenarios detected as flaky (but not yet quarantined).
	FlakyScenarios []string `json:"flaky_scenarios,omitempty"`
}

// PreflightResult holds the result of preflight checks.
type PreflightResult struct {
	// Passed indicates if all checks passed.
	Passed bool `json:"passed"`

	// Checks holds individual check results.
	Checks []PreflightCheck `json:"checks"`

	// Warnings holds non-fatal issues.
	Warnings []string `json:"warnings,omitempty"`
}

// PreflightCheck represents a single preflight check.
type PreflightCheck struct {
	// Name is the check name.
	Name string `json:"name"`

	// Passed indicates if the check passed.
	Passed bool `json:"passed"`

	// Message is the check result message.
	Message string `json:"message"`

	// Error contains the error if failed.
	Error string `json:"error,omitempty"`

	// Fix suggests how to fix the issue.
	Fix string `json:"fix,omitempty"`
}

// Comparison holds the result of comparing to a previous batch.
type Comparison struct {
	// BaselineID is the batch being compared to.
	BaselineID string `json:"baseline_id"`

	// Fixed are issues that were fixed since baseline.
	Fixed []ComparisonItem `json:"fixed,omitempty"`

	// NewIssues are issues that appeared since baseline.
	NewIssues []ComparisonItem `json:"new_issues,omitempty"`

	// Recurring are issues that persist across runs.
	Recurring []ComparisonItem `json:"recurring,omitempty"`

	// RegressionScore is the overall change (+improvement, -regression).
	RegressionScore int `json:"regression_score"`
}

// ComparisonItem represents a single item in a comparison.
type ComparisonItem struct {
	// Scenario is the scenario name.
	Scenario string `json:"scenario"`

	// Description describes the issue.
	Description string `json:"description"`

	// Severity is the issue severity (P0-P3).
	Severity string `json:"severity"`

	// RunCount is how many consecutive runs this has appeared.
	RunCount int `json:"run_count,omitempty"`
}
