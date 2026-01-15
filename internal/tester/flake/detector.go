// Package flake provides flake detection and quarantine management for AI User Testing.
// It tracks test run history, calculates flake rates, and automatically quarantines
// tests that exceed configurable thresholds.
package flake

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

// Config defines the configuration for flake detection.
type Config struct {
	// WindowSize is the number of recent runs to consider for flake detection.
	// Default: 10
	WindowSize int `json:"window_size" yaml:"window_size"`

	// FlakeThreshold is the failure rate above which a test is considered flaky.
	// Range: 0.0 to 1.0. Default: 0.3 (30% failures)
	FlakeThreshold float64 `json:"flake_threshold" yaml:"flake_threshold"`

	// MinRuns is the minimum number of runs required before making flake decisions.
	// Default: 3
	MinRuns int `json:"min_runs" yaml:"min_runs"`

	// AutoQuarantine enables automatic quarantining of flaky tests.
	// Default: true
	AutoQuarantine bool `json:"auto_quarantine" yaml:"auto_quarantine"`

	// AutoUnquarantine enables automatic unquarantining when tests stabilize.
	// Default: false (requires manual review)
	AutoUnquarantine bool `json:"auto_unquarantine" yaml:"auto_unquarantine"`

	// UnquarantineThreshold is the success rate required to auto-unquarantine.
	// Range: 0.0 to 1.0. Default: 0.9 (90% success over window)
	UnquarantineThreshold float64 `json:"unquarantine_threshold" yaml:"unquarantine_threshold"`

	// ConsecutiveFailuresThreshold is the number of consecutive failures before quarantine.
	// If set > 0, this overrides flake rate detection. Default: 0 (disabled)
	ConsecutiveFailuresThreshold int `json:"consecutive_failures_threshold" yaml:"consecutive_failures_threshold"`
}

// DefaultConfig returns the default flake detection configuration.
func DefaultConfig() Config {
	return Config{
		WindowSize:                   10,
		FlakeThreshold:               0.3,
		MinRuns:                      3,
		AutoQuarantine:               true,
		AutoUnquarantine:             false,
		UnquarantineThreshold:        0.9,
		ConsecutiveFailuresThreshold: 0,
	}
}

// RunOutcome represents the result of a single test run.
type RunOutcome string

const (
	OutcomePass  RunOutcome = "pass"
	OutcomeFail  RunOutcome = "fail"
	OutcomeError RunOutcome = "error" // Infrastructure error (retryable)
	OutcomeSkip  RunOutcome = "skip"
)

// RunRecord represents a single test run record.
type RunRecord struct {
	// Timestamp is when the run occurred.
	Timestamp time.Time `json:"timestamp"`

	// Outcome is the result of the run.
	Outcome RunOutcome `json:"outcome"`

	// RetryCount is how many retries were needed.
	RetryCount int `json:"retry_count"`

	// Duration is how long the run took.
	Duration time.Duration `json:"duration"`

	// BatchID links to the batch run (if applicable).
	BatchID string `json:"batch_id,omitempty"`

	// ErrorType categorizes the error (for error outcomes).
	ErrorType string `json:"error_type,omitempty"`

	// InfrastructureError indicates if this was an infra vs test failure.
	InfrastructureError bool `json:"infrastructure_error,omitempty"`
}

// ScenarioHistory tracks the run history for a single scenario.
type ScenarioHistory struct {
	// Scenario is the scenario name.
	Scenario string `json:"scenario"`

	// Runs is the list of run records (most recent first).
	Runs []RunRecord `json:"runs"`

	// FirstRun is when this scenario was first run.
	FirstRun time.Time `json:"first_run"`

	// LastRun is when this scenario was last run.
	LastRun time.Time `json:"last_run"`

	// TotalRuns is the total number of runs.
	TotalRuns int `json:"total_runs"`

	// TotalPasses is the total number of passes.
	TotalPasses int `json:"total_passes"`

	// TotalFailures is the total number of failures.
	TotalFailures int `json:"total_failures"`

	// TotalErrors is the total number of infrastructure errors.
	TotalErrors int `json:"total_errors"`

	// ConsecutiveFailures is the current consecutive failure count.
	ConsecutiveFailures int `json:"consecutive_failures"`

	// ConsecutivePasses is the current consecutive pass count.
	ConsecutivePasses int `json:"consecutive_passes"`
}

// FlakeMetrics contains calculated flake metrics for a scenario.
type FlakeMetrics struct {
	// Scenario is the scenario name.
	Scenario string `json:"scenario"`

	// FlakeRate is the failure rate over the window (0.0 to 1.0).
	FlakeRate float64 `json:"flake_rate"`

	// SuccessRate is the success rate over the window (0.0 to 1.0).
	SuccessRate float64 `json:"success_rate"`

	// WindowRuns is the number of runs in the window.
	WindowRuns int `json:"window_runs"`

	// WindowPasses is the number of passes in the window.
	WindowPasses int `json:"window_passes"`

	// WindowFailures is the number of failures in the window.
	WindowFailures int `json:"window_failures"`

	// WindowErrors is the number of errors in the window.
	WindowErrors int `json:"window_errors"`

	// IsFlaky indicates if the test is considered flaky.
	IsFlaky bool `json:"is_flaky"`

	// IsStable indicates if the test is considered stable.
	IsStable bool `json:"is_stable"`

	// ConsecutiveFailures is the current consecutive failure count.
	ConsecutiveFailures int `json:"consecutive_failures"`

	// ConsecutivePasses is the current consecutive pass count.
	ConsecutivePasses int `json:"consecutive_passes"`

	// LastOutcome is the most recent run outcome.
	LastOutcome RunOutcome `json:"last_outcome"`

	// AverageRetries is the average retry count over the window.
	AverageRetries float64 `json:"average_retries"`

	// AverageDuration is the average duration over the window.
	AverageDuration time.Duration `json:"average_duration"`
}

// QuarantineEntry represents a quarantined scenario.
type QuarantineEntry struct {
	// Scenario is the scenario name.
	Scenario string `json:"scenario"`

	// QuarantinedAt is when the scenario was quarantined.
	QuarantinedAt time.Time `json:"quarantined_at"`

	// Reason is why the scenario was quarantined.
	Reason string `json:"reason"`

	// FlakeRate is the flake rate when quarantined.
	FlakeRate float64 `json:"flake_rate"`

	// AutoQuarantined indicates if this was auto-quarantined.
	AutoQuarantined bool `json:"auto_quarantined"`

	// ReviewRequired indicates if manual review is needed.
	ReviewRequired bool `json:"review_required"`

	// LastRunAt is when the scenario was last run.
	LastRunAt *time.Time `json:"last_run_at,omitempty"`

	// Notes contains any manual notes about the quarantine.
	Notes string `json:"notes,omitempty"`
}

// QuarantineAction represents an action taken by the detector.
type QuarantineAction struct {
	// Action is the type of action (quarantine, unquarantine, flag).
	Action string `json:"action"`

	// Scenario is the affected scenario.
	Scenario string `json:"scenario"`

	// Reason is why the action was taken.
	Reason string `json:"reason"`

	// Metrics contains the metrics that triggered the action.
	Metrics *FlakeMetrics `json:"metrics,omitempty"`

	// Timestamp is when the action was taken.
	Timestamp time.Time `json:"timestamp"`
}

// Detector tracks test run history and detects flaky tests.
type Detector struct {
	config      Config
	storagePath string

	// history maps scenario name to its history.
	history map[string]*ScenarioHistory

	// quarantine maps scenario name to quarantine entry.
	quarantine map[string]*QuarantineEntry

	mu sync.RWMutex
}

// NewDetector creates a new flake detector.
func NewDetector(storagePath string, config Config) (*Detector, error) {
	d := &Detector{
		config:      config,
		storagePath: storagePath,
		history:     make(map[string]*ScenarioHistory),
		quarantine:  make(map[string]*QuarantineEntry),
	}

	// Apply defaults for zero values
	if d.config.WindowSize <= 0 {
		d.config.WindowSize = 10
	}
	if d.config.FlakeThreshold <= 0 {
		d.config.FlakeThreshold = 0.3
	}
	if d.config.MinRuns <= 0 {
		d.config.MinRuns = 3
	}
	if d.config.UnquarantineThreshold <= 0 {
		d.config.UnquarantineThreshold = 0.9
	}

	// Load existing data
	if err := d.load(); err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to load flake data: %w", err)
	}

	return d, nil
}

// RecordRun records a test run outcome and returns any quarantine actions.
func (d *Detector) RecordRun(scenario string, record RunRecord) ([]QuarantineAction, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	// Get or create history
	hist, ok := d.history[scenario]
	if !ok {
		hist = &ScenarioHistory{
			Scenario: scenario,
			FirstRun: record.Timestamp,
			Runs:     []RunRecord{},
		}
		d.history[scenario] = hist
	}

	// Prepend the new record (most recent first)
	hist.Runs = append([]RunRecord{record}, hist.Runs...)
	hist.LastRun = record.Timestamp
	hist.TotalRuns++

	// Update totals based on outcome
	switch record.Outcome {
	case OutcomePass:
		hist.TotalPasses++
		hist.ConsecutivePasses++
		hist.ConsecutiveFailures = 0
	case OutcomeFail:
		hist.TotalFailures++
		hist.ConsecutiveFailures++
		hist.ConsecutivePasses = 0
	case OutcomeError:
		if record.InfrastructureError {
			hist.TotalErrors++
		} else {
			hist.TotalFailures++
		}
		hist.ConsecutiveFailures++
		hist.ConsecutivePasses = 0
	}

	// Trim history to window size * 2 (keep some buffer)
	maxHistory := d.config.WindowSize * 2
	if len(hist.Runs) > maxHistory {
		hist.Runs = hist.Runs[:maxHistory]
	}

	// Calculate metrics and determine actions
	metrics := d.calculateMetrics(scenario)
	actions := d.determineActions(scenario, metrics)

	// Save updated state
	if err := d.save(); err != nil {
		return actions, fmt.Errorf("failed to save flake data: %w", err)
	}

	return actions, nil
}

// GetMetrics returns flake metrics for a scenario.
func (d *Detector) GetMetrics(scenario string) *FlakeMetrics {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.calculateMetrics(scenario)
}

// GetAllMetrics returns flake metrics for all tracked scenarios.
func (d *Detector) GetAllMetrics() []*FlakeMetrics {
	d.mu.RLock()
	defer d.mu.RUnlock()

	var metrics []*FlakeMetrics
	for scenario := range d.history {
		metrics = append(metrics, d.calculateMetrics(scenario))
	}

	// Sort by flake rate descending
	sort.Slice(metrics, func(i, j int) bool {
		return metrics[i].FlakeRate > metrics[j].FlakeRate
	})

	return metrics
}

// GetFlakyScenarios returns all scenarios currently considered flaky.
func (d *Detector) GetFlakyScenarios() []*FlakeMetrics {
	d.mu.RLock()
	defer d.mu.RUnlock()

	var flaky []*FlakeMetrics
	for scenario := range d.history {
		metrics := d.calculateMetrics(scenario)
		if metrics.IsFlaky {
			flaky = append(flaky, metrics)
		}
	}

	// Sort by flake rate descending
	sort.Slice(flaky, func(i, j int) bool {
		return flaky[i].FlakeRate > flaky[j].FlakeRate
	})

	return flaky
}

// IsQuarantined checks if a scenario is quarantined.
func (d *Detector) IsQuarantined(scenario string) bool {
	d.mu.RLock()
	defer d.mu.RUnlock()
	_, ok := d.quarantine[scenario]
	return ok
}

// GetQuarantineEntry returns the quarantine entry for a scenario.
func (d *Detector) GetQuarantineEntry(scenario string) *QuarantineEntry {
	d.mu.RLock()
	defer d.mu.RUnlock()
	if entry, ok := d.quarantine[scenario]; ok {
		copy := *entry
		return &copy
	}
	return nil
}

// ListQuarantined returns all quarantined scenarios.
func (d *Detector) ListQuarantined() []*QuarantineEntry {
	d.mu.RLock()
	defer d.mu.RUnlock()

	entries := make([]*QuarantineEntry, 0, len(d.quarantine))
	for _, entry := range d.quarantine {
		copy := *entry
		entries = append(entries, &copy)
	}

	// Sort by quarantine date
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].QuarantinedAt.After(entries[j].QuarantinedAt)
	})

	return entries
}

// Quarantine manually quarantines a scenario.
func (d *Detector) Quarantine(scenario, reason string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	var flakeRate float64
	if hist, ok := d.history[scenario]; ok {
		metrics := d.calculateMetricsUnlocked(scenario)
		flakeRate = metrics.FlakeRate
		now := time.Now()
		d.quarantine[scenario] = &QuarantineEntry{
			Scenario:        scenario,
			QuarantinedAt:   now,
			Reason:          reason,
			FlakeRate:       flakeRate,
			AutoQuarantined: false,
			ReviewRequired:  false,
			LastRunAt:       &hist.LastRun,
		}
	} else {
		d.quarantine[scenario] = &QuarantineEntry{
			Scenario:        scenario,
			QuarantinedAt:   time.Now(),
			Reason:          reason,
			FlakeRate:       0,
			AutoQuarantined: false,
			ReviewRequired:  false,
		}
	}

	return d.save()
}

// Unquarantine removes a scenario from quarantine.
func (d *Detector) Unquarantine(scenario string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	delete(d.quarantine, scenario)
	return d.save()
}

// GetHistory returns the run history for a scenario.
func (d *Detector) GetHistory(scenario string) *ScenarioHistory {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if hist, ok := d.history[scenario]; ok {
		// Return a copy
		copy := *hist
		copy.Runs = make([]RunRecord, len(hist.Runs))
		for i, r := range hist.Runs {
			copy.Runs[i] = r
		}
		return &copy
	}
	return nil
}

// ClearHistory clears the run history for a scenario.
func (d *Detector) ClearHistory(scenario string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	delete(d.history, scenario)
	return d.save()
}

// calculateMetrics calculates flake metrics for a scenario.
// Caller must hold at least a read lock.
func (d *Detector) calculateMetrics(scenario string) *FlakeMetrics {
	return d.calculateMetricsUnlocked(scenario)
}

// calculateMetricsUnlocked calculates metrics without locking.
func (d *Detector) calculateMetricsUnlocked(scenario string) *FlakeMetrics {
	metrics := &FlakeMetrics{
		Scenario: scenario,
	}

	hist, ok := d.history[scenario]
	if !ok || len(hist.Runs) == 0 {
		return metrics
	}

	// Calculate window metrics
	windowEnd := d.config.WindowSize
	if windowEnd > len(hist.Runs) {
		windowEnd = len(hist.Runs)
	}

	var totalDuration time.Duration
	var totalRetries int

	for i := 0; i < windowEnd; i++ {
		run := hist.Runs[i]
		metrics.WindowRuns++
		totalDuration += run.Duration
		totalRetries += run.RetryCount

		switch run.Outcome {
		case OutcomePass:
			metrics.WindowPasses++
		case OutcomeFail:
			metrics.WindowFailures++
		case OutcomeError:
			if run.InfrastructureError {
				metrics.WindowErrors++
			} else {
				metrics.WindowFailures++
			}
		}
	}

	// Calculate rates
	if metrics.WindowRuns > 0 {
		// Flake rate = (failures + errors) / total
		metrics.FlakeRate = float64(metrics.WindowFailures+metrics.WindowErrors) / float64(metrics.WindowRuns)
		metrics.SuccessRate = float64(metrics.WindowPasses) / float64(metrics.WindowRuns)
		metrics.AverageRetries = float64(totalRetries) / float64(metrics.WindowRuns)
		metrics.AverageDuration = totalDuration / time.Duration(metrics.WindowRuns)
	}

	// Set consecutive counts
	metrics.ConsecutiveFailures = hist.ConsecutiveFailures
	metrics.ConsecutivePasses = hist.ConsecutivePasses

	// Set last outcome
	if len(hist.Runs) > 0 {
		metrics.LastOutcome = hist.Runs[0].Outcome
	}

	// Determine flaky status
	if metrics.WindowRuns >= d.config.MinRuns {
		metrics.IsFlaky = metrics.FlakeRate >= d.config.FlakeThreshold

		// Also check consecutive failures threshold
		if d.config.ConsecutiveFailuresThreshold > 0 {
			if metrics.ConsecutiveFailures >= d.config.ConsecutiveFailuresThreshold {
				metrics.IsFlaky = true
			}
		}
	}

	// Determine stable status
	if metrics.WindowRuns >= d.config.MinRuns {
		metrics.IsStable = metrics.SuccessRate >= d.config.UnquarantineThreshold
	}

	return metrics
}

// determineActions determines quarantine actions based on metrics.
// Caller must hold the write lock.
func (d *Detector) determineActions(scenario string, metrics *FlakeMetrics) []QuarantineAction {
	var actions []QuarantineAction
	now := time.Now()

	_, isQuarantined := d.quarantine[scenario]

	// Check for auto-quarantine
	if !isQuarantined && d.config.AutoQuarantine && metrics.IsFlaky {
		reason := fmt.Sprintf("Auto-quarantined: %.0f%% failure rate over %d runs",
			metrics.FlakeRate*100, metrics.WindowRuns)

		if d.config.ConsecutiveFailuresThreshold > 0 && metrics.ConsecutiveFailures >= d.config.ConsecutiveFailuresThreshold {
			reason = fmt.Sprintf("Auto-quarantined: %d consecutive failures",
				metrics.ConsecutiveFailures)
		}

		d.quarantine[scenario] = &QuarantineEntry{
			Scenario:        scenario,
			QuarantinedAt:   now,
			Reason:          reason,
			FlakeRate:       metrics.FlakeRate,
			AutoQuarantined: true,
			ReviewRequired:  true,
		}

		actions = append(actions, QuarantineAction{
			Action:    "quarantine",
			Scenario:  scenario,
			Reason:    reason,
			Metrics:   metrics,
			Timestamp: now,
		})
	}

	// Check for auto-unquarantine
	if isQuarantined && d.config.AutoUnquarantine && metrics.IsStable {
		entry := d.quarantine[scenario]
		if entry.AutoQuarantined {
			reason := fmt.Sprintf("Auto-unquarantined: %.0f%% success rate over %d runs",
				metrics.SuccessRate*100, metrics.WindowRuns)

			delete(d.quarantine, scenario)

			actions = append(actions, QuarantineAction{
				Action:    "unquarantine",
				Scenario:  scenario,
				Reason:    reason,
				Metrics:   metrics,
				Timestamp: now,
			})
		}
	}

	// Flag for review if flaky but not auto-quarantining
	if !d.config.AutoQuarantine && metrics.IsFlaky && !isQuarantined {
		actions = append(actions, QuarantineAction{
			Action:    "flag",
			Scenario:  scenario,
			Reason:    fmt.Sprintf("Flagged as flaky: %.0f%% failure rate", metrics.FlakeRate*100),
			Metrics:   metrics,
			Timestamp: now,
		})
	}

	return actions
}

// storageData is the serialization format for the detector state.
type storageData struct {
	Version    int                         `json:"version"`
	Config     Config                      `json:"config"`
	History    map[string]*ScenarioHistory `json:"history"`
	Quarantine map[string]*QuarantineEntry `json:"quarantine"`
	UpdatedAt  time.Time                   `json:"updated_at"`
}

// load loads the detector state from disk.
func (d *Detector) load() error {
	data, err := os.ReadFile(d.storagePath)
	if err != nil {
		return err
	}

	var storage storageData
	if err := json.Unmarshal(data, &storage); err != nil {
		return fmt.Errorf("failed to parse flake data: %w", err)
	}

	d.history = storage.History
	d.quarantine = storage.Quarantine

	// Initialize maps if nil
	if d.history == nil {
		d.history = make(map[string]*ScenarioHistory)
	}
	if d.quarantine == nil {
		d.quarantine = make(map[string]*QuarantineEntry)
	}

	return nil
}

// save saves the detector state to disk.
func (d *Detector) save() error {
	storage := storageData{
		Version:    1,
		Config:     d.config,
		History:    d.history,
		Quarantine: d.quarantine,
		UpdatedAt:  time.Now(),
	}

	data, err := json.MarshalIndent(storage, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize flake data: %w", err)
	}

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(d.storagePath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	return os.WriteFile(d.storagePath, data, 0644)
}
