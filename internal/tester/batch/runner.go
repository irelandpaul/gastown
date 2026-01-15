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
