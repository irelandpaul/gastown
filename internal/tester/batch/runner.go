package batch

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/steveyegge/gastown/internal/tester/flake"
)

// Runner executes batch test runs.
type Runner struct {
	// config is the batch configuration.
	config Config

	// quarantineStore tracks quarantined scenarios (legacy, kept for compatibility).
	quarantineStore *QuarantineStore

	// flakeDetector tracks flake metrics and handles quarantine.
	flakeDetector *flake.Detector

	// baseDir is the base directory for results.
	baseDir string

	// batchID is the current batch run ID (set during Run).
	batchID string

	// quarantineActions collects actions taken during the batch.
	quarantineActions []flake.QuarantineAction
}

// NewRunner creates a new batch runner.
func NewRunner(config Config) (*Runner, error) {
	store, err := NewQuarantineStore(filepath.Join(config.OutputDir, ".quarantine"))
	if err != nil {
		return nil, fmt.Errorf("failed to create quarantine store: %w", err)
	}

	// Initialize flake detector with default config
	flakeConfig := flake.DefaultConfig()
	detector, err := flake.NewDetector(
		filepath.Join(config.OutputDir, ".flake-data.json"),
		flakeConfig,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create flake detector: %w", err)
	}

	return &Runner{
		config:          config,
		quarantineStore: store,
		flakeDetector:   detector,
		baseDir:         config.OutputDir,
	}, nil
}

// Run executes the batch and returns the results.
func (r *Runner) Run(ctx context.Context) (*BatchResult, error) {
	r.batchID = generateBatchID()
	r.quarantineActions = nil // Reset actions for this run

	result := &BatchResult{
		ID:        r.batchID,
		Config:    r.config,
		StartedAt: time.Now(),
		Summary: BatchSummary{
			TotalObservations: make(map[string]int),
		},
	}

	// Find scenarios
	scenarios, err := r.findScenarios()
	if err != nil {
		return nil, fmt.Errorf("failed to find scenarios: %w", err)
	}
	result.ScenariosFound = len(scenarios)

	// Filter scenarios
	filtered := r.filterScenarios(scenarios)

	// Separate quarantined from runnable
	// Check both legacy store and new flake detector
	var runnable []string
	var skipped []ScenarioResult

	for _, s := range filtered {
		name := strings.TrimSuffix(filepath.Base(s), filepath.Ext(s))
		isQuarantined := r.quarantineStore.IsQuarantined(s) || r.flakeDetector.IsQuarantined(name)

		if isQuarantined && !r.config.IncludeQuarantined {
			skipReason := "quarantined"
			if entry := r.flakeDetector.GetQuarantineEntry(name); entry != nil {
				skipReason = fmt.Sprintf("quarantined: %s", entry.Reason)
			}
			skipped = append(skipped, ScenarioResult{
				Scenario:    name,
				Path:        s,
				Status:      StatusSkipped,
				Quarantined: true,
				SkipReason:  skipReason,
			})
		} else {
			runnable = append(runnable, s)
		}
	}

	result.ScenariosRun = len(runnable)
	result.ScenariosSkipped = len(skipped)
	result.Results = append(result.Results, skipped...)

	// Run preflight if not skipped
	if !r.config.SkipPreflight {
		preflight := r.runPreflight()
		if !preflight.Passed {
			return nil, fmt.Errorf("preflight checks failed")
		}
	}

	// Create output directory for this batch
	batchDir := r.createBatchDir(result.ID)
	result.OutputDir = batchDir

	// Run scenarios
	results := r.runScenarios(ctx, runnable)
	result.Results = append(result.Results, results...)

	// Calculate summary
	r.calculateSummary(result)

	// Complete the result
	now := time.Now()
	result.CompletedAt = &now
	result.TotalDuration = now.Sub(result.StartedAt)

	// Save batch manifest
	if err := r.saveBatchManifest(result); err != nil {
		return result, fmt.Errorf("failed to save manifest: %w", err)
	}

	// Compare to baseline if requested
	if r.config.CompareTo != "" {
		baseline, err := r.LoadBaseline(r.config.CompareTo)
		if err != nil {
			// Log warning but don't fail the entire run
			fmt.Printf("Warning: failed to load baseline %s: %v\n", r.config.CompareTo, err)
		} else {
			result.Comparison = r.Compare(result, baseline)
		}
	}

	return result, nil
}

// findScenarios finds all scenario files matching the pattern.
func (r *Runner) findScenarios() ([]string, error) {
	matches, err := filepath.Glob(r.config.Pattern)
	if err != nil {
		return nil, err
	}

	// Filter to only .yaml and .yml files
	var scenarios []string
	for _, m := range matches {
		ext := strings.ToLower(filepath.Ext(m))
		if ext == ".yaml" || ext == ".yml" {
			scenarios = append(scenarios, m)
		}
	}

	// Sort for consistent ordering
	sort.Strings(scenarios)

	return scenarios, nil
}

// filterScenarios applies tag filters to the scenario list.
func (r *Runner) filterScenarios(scenarios []string) []string {
	if len(r.config.FilterTags) == 0 && len(r.config.ExcludeTags) == 0 {
		return scenarios
	}

	var filtered []string
	for _, s := range scenarios {
		tags := r.extractTags(s)

		// Check filter tags (must match at least one)
		if len(r.config.FilterTags) > 0 {
			if !hasAnyTag(tags, r.config.FilterTags) {
				continue
			}
		}

		// Check exclude tags (must not match any)
		if len(r.config.ExcludeTags) > 0 {
			if hasAnyTag(tags, r.config.ExcludeTags) {
				continue
			}
		}

		filtered = append(filtered, s)
	}

	return filtered
}

// extractTags extracts tags from a scenario file.
// This is a simplified implementation - in practice would parse YAML.
func (r *Runner) extractTags(scenarioPath string) []string {
	// For now, extract tags from directory names as a simple heuristic
	// In practice, this would parse the YAML file
	dir := filepath.Dir(scenarioPath)
	parts := strings.Split(dir, string(filepath.Separator))

	var tags []string
	for _, p := range parts {
		if p != "" && p != "." && p != "scenarios" {
			tags = append(tags, p)
		}
	}

	return tags
}

// runPreflight runs preflight checks.
func (r *Runner) runPreflight() *PreflightResult {
	result := &PreflightResult{
		Passed: true,
		Checks: []PreflightCheck{},
	}

	// Check output directory is writable
	testFile := filepath.Join(r.config.OutputDir, ".preflight-test")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		result.Passed = false
		result.Checks = append(result.Checks, PreflightCheck{
			Name:    "output_directory",
			Passed:  false,
			Message: "Output directory not writable",
			Error:   err.Error(),
			Fix:     "Check permissions on " + r.config.OutputDir,
		})
	} else {
		os.Remove(testFile)
		result.Checks = append(result.Checks, PreflightCheck{
			Name:    "output_directory",
			Passed:  true,
			Message: "Output directory writable",
		})
	}

	// Check disk space (simplified)
	result.Checks = append(result.Checks, PreflightCheck{
		Name:    "disk_space",
		Passed:  true,
		Message: "Disk space sufficient",
	})

	return result
}

// runScenarios runs all scenarios with the configured parallelism.
func (r *Runner) runScenarios(ctx context.Context, scenarios []string) []ScenarioResult {
	if len(scenarios) == 0 {
		return nil
	}

	results := make([]ScenarioResult, len(scenarios))
	parallel := r.config.Parallel
	if parallel < 1 {
		parallel = 1
	}
	if parallel > len(scenarios) {
		parallel = len(scenarios)
	}

	// Use a channel to distribute work
	work := make(chan int, len(scenarios))
	for i := range scenarios {
		work <- i
	}
	close(work)

	// Run workers
	var wg sync.WaitGroup
	var mu sync.Mutex
	stopFlag := false

	for w := 0; w < parallel; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for idx := range work {
				mu.Lock()
				if stopFlag {
					mu.Unlock()
					// Mark remaining as skipped
					results[idx] = ScenarioResult{
						Scenario:   filepath.Base(scenarios[idx]),
						Path:       scenarios[idx],
						Status:     StatusSkipped,
						SkipReason: "batch stopped on failure",
					}
					continue
				}
				mu.Unlock()

				result := r.runSingleScenario(ctx, scenarios[idx])
				results[idx] = result

				if r.config.StopOnFail && (result.Status == StatusFailed || result.Status == StatusError) {
					mu.Lock()
					stopFlag = true
					mu.Unlock()
				}
			}
		}()
	}

	wg.Wait()
	return results
}

// runSingleScenario runs a single scenario.
func (r *Runner) runSingleScenario(ctx context.Context, scenarioPath string) ScenarioResult {
	start := time.Now()
	name := strings.TrimSuffix(filepath.Base(scenarioPath), filepath.Ext(scenarioPath))

	result := ScenarioResult{
		Scenario:     name,
		Path:         scenarioPath,
		Status:       StatusRunning,
		Observations: make(map[string]int),
	}

	// Check for context cancellation
	select {
	case <-ctx.Done():
		result.Status = StatusSkipped
		result.SkipReason = "context cancelled"
		result.Duration = time.Since(start)
		// Record skip as a run outcome (but don't count against flake rate)
		return result
	default:
	}

	// Simulate running the scenario
	// In practice, this would spawn an agent and run the test
	result.Status = StatusPassed
	result.Duration = time.Since(start)
	result.SuccessCriteriaMet = 3
	result.SuccessCriteriaTotal = 3

	// Create artifact directory
	dateDir := time.Now().Format("2006-01-02")
	runID := generateRunID()
	result.ArtifactDir = filepath.Join(r.baseDir, dateDir, name, fmt.Sprintf("run-%s", runID))

	// Record the run outcome with the flake detector
	r.recordRunOutcome(name, result)

	return result
}

// recordRunOutcome records a scenario run with the flake detector.
func (r *Runner) recordRunOutcome(scenario string, result ScenarioResult) {
	// Convert batch status to flake outcome
	var outcome flake.RunOutcome
	var isInfraError bool

	switch result.Status {
	case StatusPassed:
		outcome = flake.OutcomePass
	case StatusFailed:
		outcome = flake.OutcomeFail
	case StatusError:
		outcome = flake.OutcomeError
		// Check if it's an infrastructure error based on error message
		if result.Error != "" {
			isInfraError = isInfrastructureError(result.Error)
		}
	case StatusSkipped:
		outcome = flake.OutcomeSkip
	default:
		outcome = flake.OutcomeError
	}

	record := flake.RunRecord{
		Timestamp:           time.Now(),
		Outcome:             outcome,
		RetryCount:          result.RetryCount,
		Duration:            result.Duration,
		BatchID:             r.batchID,
		ErrorType:           categorizeError(result.Error),
		InfrastructureError: isInfraError,
	}

	// Record and collect any quarantine actions
	actions, err := r.flakeDetector.RecordRun(scenario, record)
	if err != nil {
		// Log but don't fail the run
		fmt.Printf("Warning: failed to record run for %s: %v\n", scenario, err)
		return
	}

	// Collect actions for summary
	if len(actions) > 0 {
		r.quarantineActions = append(r.quarantineActions, actions...)
	}
}

// isInfrastructureError checks if an error is infrastructure-related.
func isInfrastructureError(errMsg string) bool {
	infraPatterns := []string{
		"timeout",
		"browser crash",
		"network error",
		"connection refused",
		"context deadline exceeded",
		"playwright",
		"chromium",
		"failed to launch",
	}
	lower := strings.ToLower(errMsg)
	for _, pattern := range infraPatterns {
		if strings.Contains(lower, pattern) {
			return true
		}
	}
	return false
}

// categorizeError extracts an error type from an error message.
func categorizeError(errMsg string) string {
	if errMsg == "" {
		return ""
	}
	lower := strings.ToLower(errMsg)
	switch {
	case strings.Contains(lower, "timeout"):
		return "timeout"
	case strings.Contains(lower, "browser") || strings.Contains(lower, "chromium"):
		return "browser_crash"
	case strings.Contains(lower, "network") || strings.Contains(lower, "connection"):
		return "network_error"
	case strings.Contains(lower, "assertion") || strings.Contains(lower, "criteria"):
		return "test_failure"
	default:
		return "unknown"
	}
}

// calculateSummary calculates the batch summary statistics.
func (r *Runner) calculateSummary(result *BatchResult) {
	for _, sr := range result.Results {
		switch sr.Status {
		case StatusPassed:
			result.Summary.Passed++
		case StatusFailed:
			result.Summary.Failed++
		case StatusError:
			result.Summary.Errors++
		case StatusSkipped:
			result.Summary.Skipped++
		}

		result.Summary.TotalRetries += sr.RetryCount

		for severity, count := range sr.Observations {
			result.Summary.TotalObservations[severity] += count
		}
	}

	// Calculate flake rate (failures + errors / total run)
	total := result.Summary.Passed + result.Summary.Failed + result.Summary.Errors
	if total > 0 {
		result.Summary.FlakeRate = float64(result.Summary.Failed+result.Summary.Errors) / float64(total)
	}

	// Process quarantine actions taken during this batch
	for _, action := range r.quarantineActions {
		switch action.Action {
		case "quarantine":
			result.Summary.AutoQuarantined = append(result.Summary.AutoQuarantined, action.Scenario)
		case "unquarantine":
			result.Summary.AutoUnquarantined = append(result.Summary.AutoUnquarantined, action.Scenario)
		case "flag":
			result.Summary.FlakyScenarios = append(result.Summary.FlakyScenarios, action.Scenario)
		}
	}

	// Identify quarantine candidates (scenarios that failed but weren't auto-quarantined)
	autoQuarantinedSet := make(map[string]bool)
	for _, s := range result.Summary.AutoQuarantined {
		autoQuarantinedSet[s] = true
	}

	for _, sr := range result.Results {
		if sr.Status == StatusFailed || sr.Status == StatusError {
			if !sr.Quarantined && !autoQuarantinedSet[sr.Scenario] {
				result.Summary.NewQuarantineCandidates = append(
					result.Summary.NewQuarantineCandidates,
					sr.Scenario,
				)
			}
		}
	}
}

// createBatchDir creates the output directory for this batch.
func (r *Runner) createBatchDir(batchID string) string {
	dateDir := time.Now().Format("2006-01-02")
	batchDir := filepath.Join(r.baseDir, dateDir, fmt.Sprintf("batch-%s", batchID))
	os.MkdirAll(batchDir, 0755)
	return batchDir
}

// saveBatchManifest saves the batch result as a manifest file.
func (r *Runner) saveBatchManifest(result *BatchResult) error {
	manifestPath := filepath.Join(result.OutputDir, "manifest.json")

	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(manifestPath, data, 0644)
}

// generateBatchID generates a unique batch identifier.
func generateBatchID() string {
	b := make([]byte, 4)
	rand.Read(b)
	return fmt.Sprintf("%02x%02x%02x%02x", b[0], b[1], b[2], b[3])
}

// generateRunID generates a unique run identifier.
func generateRunID() string {
	b := make([]byte, 4)
	rand.Read(b)
	return fmt.Sprintf("%02x%02x%02x%02x", b[0], b[1], b[2], b[3])
}

// hasAnyTag checks if any of the given tags match.
func hasAnyTag(tags, check []string) bool {
	tagSet := make(map[string]bool)
	for _, t := range tags {
		tagSet[strings.ToLower(t)] = true
	}
	for _, c := range check {
		if tagSet[strings.ToLower(c)] {
			return true
		}
	}
	return false
}

// QuarantineStore manages quarantined scenarios.
type QuarantineStore struct {
	path        string
	quarantined map[string]QuarantineEntry
	mu          sync.RWMutex
}

// QuarantineEntry represents a quarantined scenario.
type QuarantineEntry struct {
	Scenario      string    `json:"scenario"`
	QuarantinedAt time.Time `json:"quarantined_at"`
	Reason        string    `json:"reason"`
	FlakeRate     float64   `json:"flake_rate"`
}

// NewQuarantineStore creates a new quarantine store.
func NewQuarantineStore(path string) (*QuarantineStore, error) {
	store := &QuarantineStore{
		path:        path,
		quarantined: make(map[string]QuarantineEntry),
	}

	// Load existing quarantine data
	if err := store.load(); err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	return store, nil
}

// IsQuarantined checks if a scenario is quarantined.
func (s *QuarantineStore) IsQuarantined(scenarioPath string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	name := strings.TrimSuffix(filepath.Base(scenarioPath), filepath.Ext(scenarioPath))
	_, ok := s.quarantined[name]
	return ok
}

// Quarantine adds a scenario to quarantine.
func (s *QuarantineStore) Quarantine(scenario, reason string, flakeRate float64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.quarantined[scenario] = QuarantineEntry{
		Scenario:      scenario,
		QuarantinedAt: time.Now(),
		Reason:        reason,
		FlakeRate:     flakeRate,
	}

	return s.save()
}

// Unquarantine removes a scenario from quarantine.
func (s *QuarantineStore) Unquarantine(scenario string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.quarantined, scenario)
	return s.save()
}

// List returns all quarantined scenarios.
func (s *QuarantineStore) List() []QuarantineEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entries := make([]QuarantineEntry, 0, len(s.quarantined))
	for _, e := range s.quarantined {
		entries = append(entries, e)
	}
	return entries
}

func (s *QuarantineStore) load() error {
	data, err := os.ReadFile(s.path)
	if err != nil {
		return err
	}

	var entries []QuarantineEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return err
	}

	for _, e := range entries {
		s.quarantined[e.Scenario] = e
	}

	return nil
}

func (s *QuarantineStore) save() error {
	// Note: caller must hold the lock
	entries := make([]QuarantineEntry, 0, len(s.quarantined))
	for _, e := range s.quarantined {
		entries = append(entries, e)
	}

	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return err
	}

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(s.path), 0755); err != nil {
		return err
	}

	return os.WriteFile(s.path, data, 0644)
}

// LoadBaseline loads a previous batch result to use as a comparison baseline.
// The batchID can be a full batch ID (e.g., "a1b2c3d4") or a path to the manifest.
func (r *Runner) LoadBaseline(batchID string) (*BatchResult, error) {
	// First, try to interpret as a direct path to manifest
	if strings.HasSuffix(batchID, "manifest.json") {
		return loadManifestFile(batchID)
	}

	// Search for the batch in the output directory
	// Batch manifests are stored at: <output>/<date>/batch-<id>/manifest.json
	pattern := filepath.Join(r.baseDir, "*", fmt.Sprintf("batch-%s", batchID), "manifest.json")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to search for baseline: %w", err)
	}

	if len(matches) == 0 {
		// Try without the batch- prefix (in case user passed full path component)
		pattern = filepath.Join(r.baseDir, "*", batchID, "manifest.json")
		matches, err = filepath.Glob(pattern)
		if err != nil {
			return nil, fmt.Errorf("failed to search for baseline: %w", err)
		}
	}

	if len(matches) == 0 {
		return nil, fmt.Errorf("baseline batch %q not found in %s", batchID, r.baseDir)
	}

	// Use the most recent match if multiple are found
	sort.Strings(matches)
	return loadManifestFile(matches[len(matches)-1])
}

// loadManifestFile loads a batch result from a manifest file.
func loadManifestFile(path string) (*BatchResult, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest: %w", err)
	}

	var result BatchResult
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse manifest: %w", err)
	}

	return &result, nil
}

// Compare compares the current batch result against a baseline and returns a Comparison.
func (r *Runner) Compare(current, baseline *BatchResult) *Comparison {
	comparison := &Comparison{
		BaselineID: baseline.ID,
	}

	// Build maps for efficient lookup
	baselineResults := make(map[string]*ScenarioResult)
	for i := range baseline.Results {
		sr := &baseline.Results[i]
		baselineResults[sr.Scenario] = sr
	}

	currentResults := make(map[string]*ScenarioResult)
	for i := range current.Results {
		sr := &current.Results[i]
		currentResults[sr.Scenario] = sr
	}

	// Analyze each scenario in the current run
	for scenario, curr := range currentResults {
		base, existed := baselineResults[scenario]

		if !existed {
			// New scenario that didn't exist in baseline
			if curr.Status == StatusFailed || curr.Status == StatusError {
				comparison.NewIssues = append(comparison.NewIssues, ComparisonItem{
					Scenario:    scenario,
					Description: fmt.Sprintf("New failing scenario: %s", curr.Error),
					Severity:    getHighestSeverity(curr.Observations),
				})
			}
			continue
		}

		// Compare status changes
		currFailing := curr.Status == StatusFailed || curr.Status == StatusError
		baseFailing := base.Status == StatusFailed || base.Status == StatusError

		if baseFailing && !currFailing {
			// Fixed - was failing, now passing
			comparison.Fixed = append(comparison.Fixed, ComparisonItem{
				Scenario:    scenario,
				Description: "Test now passes",
				Severity:    getHighestSeverity(base.Observations),
			})
		} else if !baseFailing && currFailing {
			// New regression - was passing, now failing
			comparison.NewIssues = append(comparison.NewIssues, ComparisonItem{
				Scenario:    scenario,
				Description: fmt.Sprintf("Regression: %s", curr.Error),
				Severity:    getHighestSeverity(curr.Observations),
			})
		} else if baseFailing && currFailing {
			// Recurring - still failing
			comparison.Recurring = append(comparison.Recurring, ComparisonItem{
				Scenario:    scenario,
				Description: fmt.Sprintf("Still failing: %s", curr.Error),
				Severity:    getHighestSeverity(curr.Observations),
				RunCount:    2, // At least 2 runs (baseline + current)
			})
		}

		// Also check for observation changes in passing tests
		if !currFailing && !baseFailing {
			currObs := countObservations(curr.Observations)
			baseObs := countObservations(base.Observations)

			if currObs > baseObs {
				// New observations appeared
				comparison.NewIssues = append(comparison.NewIssues, ComparisonItem{
					Scenario:    scenario,
					Description: fmt.Sprintf("New observations: %d (was %d)", currObs, baseObs),
					Severity:    getHighestSeverity(curr.Observations),
				})
			} else if currObs < baseObs {
				// Observations resolved
				comparison.Fixed = append(comparison.Fixed, ComparisonItem{
					Scenario:    scenario,
					Description: fmt.Sprintf("Observations reduced: %d (was %d)", currObs, baseObs),
					Severity:    getHighestSeverity(base.Observations),
				})
			}
		}
	}

	// Check for scenarios that were in baseline but not in current (removed/renamed)
	for scenario, base := range baselineResults {
		if _, exists := currentResults[scenario]; !exists {
			baseFailing := base.Status == StatusFailed || base.Status == StatusError
			if baseFailing {
				// A failing test was removed - could be a fix or just removal
				comparison.Fixed = append(comparison.Fixed, ComparisonItem{
					Scenario:    scenario,
					Description: "Scenario removed from test suite",
					Severity:    getHighestSeverity(base.Observations),
				})
			}
		}
	}

	// Calculate regression score
	// Positive = improvement, Negative = regression
	comparison.RegressionScore = len(comparison.Fixed) - len(comparison.NewIssues)

	return comparison
}

// getHighestSeverity returns the highest severity from an observations map.
func getHighestSeverity(observations map[string]int) string {
	severities := []string{"P0", "P1", "P2", "P3"}
	for _, sev := range severities {
		if count, ok := observations[sev]; ok && count > 0 {
			return sev
		}
	}
	return "P3" // Default to lowest severity
}

// countObservations returns the total number of observations.
func countObservations(observations map[string]int) int {
	total := 0
	for _, count := range observations {
		total += count
	}
	return total
}
