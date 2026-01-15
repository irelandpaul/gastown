package batch

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.Parallel != 1 {
		t.Errorf("expected Parallel=1, got %d", config.Parallel)
	}

	if config.StopOnFail {
		t.Error("expected StopOnFail=false")
	}

	if config.Environment != "staging" {
		t.Errorf("expected Environment=staging, got %s", config.Environment)
	}

	if config.IncludeQuarantined {
		t.Error("expected IncludeQuarantined=false")
	}
}

func TestNewRunner(t *testing.T) {
	tmpDir := t.TempDir()

	config := DefaultConfig()
	config.OutputDir = tmpDir
	config.Pattern = "*.yaml"

	runner, err := NewRunner(config)
	if err != nil {
		t.Fatalf("failed to create runner: %v", err)
	}

	if runner == nil {
		t.Fatal("expected runner to be non-nil")
	}
}

func TestFindScenarios(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test scenario files
	scenarios := []string{
		"test1.yaml",
		"test2.yml",
		"subdir/test3.yaml",
		"not-yaml.txt",
	}

	for _, s := range scenarios {
		path := filepath.Join(tmpDir, s)
		os.MkdirAll(filepath.Dir(path), 0755)
		os.WriteFile(path, []byte("test: true"), 0644)
	}

	config := DefaultConfig()
	config.OutputDir = tmpDir
	config.Pattern = filepath.Join(tmpDir, "**/*.yaml")

	runner, _ := NewRunner(config)

	// Test glob pattern matching
	config.Pattern = filepath.Join(tmpDir, "*.yaml")
	runner.config = config

	found, err := runner.findScenarios()
	if err != nil {
		t.Fatalf("findScenarios failed: %v", err)
	}

	// Should find test1.yaml (but not subdir or non-yaml)
	if len(found) != 1 {
		t.Errorf("expected 1 scenario, got %d", len(found))
	}
}

func TestRunBasicBatch(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a test scenario
	scenarioPath := filepath.Join(tmpDir, "test.yaml")
	os.WriteFile(scenarioPath, []byte("scenario: test\n"), 0644)

	config := DefaultConfig()
	config.OutputDir = tmpDir
	config.Pattern = filepath.Join(tmpDir, "*.yaml")
	config.SkipPreflight = true

	runner, err := NewRunner(config)
	if err != nil {
		t.Fatalf("failed to create runner: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := runner.Run(ctx)
	if err != nil {
		t.Fatalf("batch run failed: %v", err)
	}

	if result.ScenariosFound != 1 {
		t.Errorf("expected 1 scenario found, got %d", result.ScenariosFound)
	}

	if result.ScenariosRun != 1 {
		t.Errorf("expected 1 scenario run, got %d", result.ScenariosRun)
	}

	if result.Summary.Passed != 1 {
		t.Errorf("expected 1 passed, got %d", result.Summary.Passed)
	}
}

func TestQuarantineStore(t *testing.T) {
	tmpDir := t.TempDir()
	storePath := filepath.Join(tmpDir, ".quarantine")

	store, err := NewQuarantineStore(storePath)
	if err != nil {
		t.Fatalf("failed to create quarantine store: %v", err)
	}

	// Initially empty
	if store.IsQuarantined("test_scenario") {
		t.Error("expected test_scenario to not be quarantined")
	}

	// Add to quarantine
	err = store.Quarantine("test_scenario", "flaky", 0.15)
	if err != nil {
		t.Fatalf("failed to quarantine: %v", err)
	}

	if !store.IsQuarantined("test_scenario") {
		t.Error("expected test_scenario to be quarantined")
	}

	// List should include it
	list := store.List()
	if len(list) != 1 {
		t.Errorf("expected 1 quarantined, got %d", len(list))
	}

	// Unquarantine
	err = store.Unquarantine("test_scenario")
	if err != nil {
		t.Fatalf("failed to unquarantine: %v", err)
	}

	if store.IsQuarantined("test_scenario") {
		t.Error("expected test_scenario to not be quarantined after removal")
	}
}

func TestQuarantineFiltering(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test scenarios
	os.WriteFile(filepath.Join(tmpDir, "good.yaml"), []byte("scenario: good\n"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "flaky.yaml"), []byte("scenario: flaky\n"), 0644)

	config := DefaultConfig()
	config.OutputDir = tmpDir
	config.Pattern = filepath.Join(tmpDir, "*.yaml")
	config.SkipPreflight = true

	runner, _ := NewRunner(config)

	// Quarantine flaky scenario
	runner.quarantineStore.Quarantine("flaky", "too flaky", 0.25)

	ctx := context.Background()
	result, err := runner.Run(ctx)
	if err != nil {
		t.Fatalf("batch run failed: %v", err)
	}

	// Should find 2 but only run 1
	if result.ScenariosFound != 2 {
		t.Errorf("expected 2 scenarios found, got %d", result.ScenariosFound)
	}

	if result.ScenariosRun != 1 {
		t.Errorf("expected 1 scenario run, got %d", result.ScenariosRun)
	}

	if result.ScenariosSkipped != 1 {
		t.Errorf("expected 1 skipped, got %d", result.ScenariosSkipped)
	}
}

func TestIncludeQuarantined(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test scenarios
	os.WriteFile(filepath.Join(tmpDir, "good.yaml"), []byte("scenario: good\n"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "flaky.yaml"), []byte("scenario: flaky\n"), 0644)

	config := DefaultConfig()
	config.OutputDir = tmpDir
	config.Pattern = filepath.Join(tmpDir, "*.yaml")
	config.SkipPreflight = true
	config.IncludeQuarantined = true // Include quarantined

	runner, _ := NewRunner(config)

	// Quarantine flaky scenario
	runner.quarantineStore.Quarantine("flaky", "too flaky", 0.25)

	ctx := context.Background()
	result, err := runner.Run(ctx)
	if err != nil {
		t.Fatalf("batch run failed: %v", err)
	}

	// Should run both since IncludeQuarantined is true
	if result.ScenariosRun != 2 {
		t.Errorf("expected 2 scenarios run with IncludeQuarantined, got %d", result.ScenariosRun)
	}
}

func TestBatchSummaryCalculation(t *testing.T) {
	result := &BatchResult{
		Results: []ScenarioResult{
			{Status: StatusPassed, Observations: map[string]int{"P2": 2}},
			{Status: StatusPassed, Observations: map[string]int{"P3": 1}},
			{Status: StatusFailed},
			{Status: StatusSkipped},
		},
		Summary: BatchSummary{
			TotalObservations: make(map[string]int),
		},
	}

	tmpDir := t.TempDir()
	config := DefaultConfig()
	config.OutputDir = tmpDir

	runner, _ := NewRunner(config)
	runner.calculateSummary(result)

	if result.Summary.Passed != 2 {
		t.Errorf("expected 2 passed, got %d", result.Summary.Passed)
	}

	if result.Summary.Failed != 1 {
		t.Errorf("expected 1 failed, got %d", result.Summary.Failed)
	}

	if result.Summary.Skipped != 1 {
		t.Errorf("expected 1 skipped, got %d", result.Summary.Skipped)
	}

	if result.Summary.TotalObservations["P2"] != 2 {
		t.Errorf("expected 2 P2 observations, got %d", result.Summary.TotalObservations["P2"])
	}

	if result.Summary.TotalObservations["P3"] != 1 {
		t.Errorf("expected 1 P3 observation, got %d", result.Summary.TotalObservations["P3"])
	}
}

func TestHasAnyTag(t *testing.T) {
	tests := []struct {
		tags     []string
		check    []string
		expected bool
	}{
		{[]string{"a", "b", "c"}, []string{"b"}, true},
		{[]string{"a", "b", "c"}, []string{"d"}, false},
		{[]string{"Critical-Path"}, []string{"critical-path"}, true}, // case insensitive
		{[]string{}, []string{"a"}, false},
		{[]string{"a"}, []string{}, false},
	}

	for _, tt := range tests {
		result := hasAnyTag(tt.tags, tt.check)
		if result != tt.expected {
			t.Errorf("hasAnyTag(%v, %v) = %v, expected %v",
				tt.tags, tt.check, result, tt.expected)
		}
	}
}

func TestCompare(t *testing.T) {
	tmpDir := t.TempDir()
	config := DefaultConfig()
	config.OutputDir = tmpDir

	runner, _ := NewRunner(config)

	// Create baseline result with mixed outcomes
	baseline := &BatchResult{
		ID: "baseline123",
		Results: []ScenarioResult{
			{Scenario: "login", Status: StatusPassed, Observations: map[string]int{"P2": 1}},
			{Scenario: "checkout", Status: StatusFailed, Error: "timeout"},
			{Scenario: "profile", Status: StatusFailed, Error: "element not found"},
			{Scenario: "search", Status: StatusPassed},
		},
	}

	// Create current result - checkout fixed, search regressed, profile still failing
	current := &BatchResult{
		ID: "current456",
		Results: []ScenarioResult{
			{Scenario: "login", Status: StatusPassed, Observations: map[string]int{"P2": 1}},
			{Scenario: "checkout", Status: StatusPassed}, // Fixed!
			{Scenario: "profile", Status: StatusFailed, Error: "element not found"}, // Still failing
			{Scenario: "search", Status: StatusFailed, Error: "regression"}, // Regressed!
		},
	}

	comparison := runner.Compare(current, baseline)

	// Verify baseline ID
	if comparison.BaselineID != "baseline123" {
		t.Errorf("expected BaselineID=baseline123, got %s", comparison.BaselineID)
	}

	// Verify fixed issues (checkout was fixed)
	if len(comparison.Fixed) != 1 {
		t.Errorf("expected 1 fixed, got %d", len(comparison.Fixed))
	} else if comparison.Fixed[0].Scenario != "checkout" {
		t.Errorf("expected checkout to be fixed, got %s", comparison.Fixed[0].Scenario)
	}

	// Verify new issues (search regressed)
	if len(comparison.NewIssues) != 1 {
		t.Errorf("expected 1 new issue, got %d", len(comparison.NewIssues))
	} else if comparison.NewIssues[0].Scenario != "search" {
		t.Errorf("expected search to be a new issue, got %s", comparison.NewIssues[0].Scenario)
	}

	// Verify recurring issues (profile still failing)
	if len(comparison.Recurring) != 1 {
		t.Errorf("expected 1 recurring, got %d", len(comparison.Recurring))
	} else if comparison.Recurring[0].Scenario != "profile" {
		t.Errorf("expected profile to be recurring, got %s", comparison.Recurring[0].Scenario)
	}

	// Regression score: 1 fixed - 1 new = 0
	if comparison.RegressionScore != 0 {
		t.Errorf("expected RegressionScore=0, got %d", comparison.RegressionScore)
	}
}

func TestCompareAllFixed(t *testing.T) {
	tmpDir := t.TempDir()
	config := DefaultConfig()
	config.OutputDir = tmpDir

	runner, _ := NewRunner(config)

	baseline := &BatchResult{
		ID: "baseline",
		Results: []ScenarioResult{
			{Scenario: "test1", Status: StatusFailed, Error: "bug"},
			{Scenario: "test2", Status: StatusFailed, Error: "bug"},
		},
	}

	current := &BatchResult{
		ID: "current",
		Results: []ScenarioResult{
			{Scenario: "test1", Status: StatusPassed},
			{Scenario: "test2", Status: StatusPassed},
		},
	}

	comparison := runner.Compare(current, baseline)

	if len(comparison.Fixed) != 2 {
		t.Errorf("expected 2 fixed, got %d", len(comparison.Fixed))
	}

	if len(comparison.NewIssues) != 0 {
		t.Errorf("expected 0 new issues, got %d", len(comparison.NewIssues))
	}

	if comparison.RegressionScore != 2 {
		t.Errorf("expected RegressionScore=2 (improvement), got %d", comparison.RegressionScore)
	}
}

func TestCompareObservationChanges(t *testing.T) {
	tmpDir := t.TempDir()
	config := DefaultConfig()
	config.OutputDir = tmpDir

	runner, _ := NewRunner(config)

	baseline := &BatchResult{
		ID: "baseline",
		Results: []ScenarioResult{
			{Scenario: "test1", Status: StatusPassed, Observations: map[string]int{"P2": 3}},
		},
	}

	// More observations = regression
	current := &BatchResult{
		ID: "current",
		Results: []ScenarioResult{
			{Scenario: "test1", Status: StatusPassed, Observations: map[string]int{"P2": 5}},
		},
	}

	comparison := runner.Compare(current, baseline)

	if len(comparison.NewIssues) != 1 {
		t.Errorf("expected 1 new issue for increased observations, got %d", len(comparison.NewIssues))
	}

	// Fewer observations = improvement
	currentBetter := &BatchResult{
		ID: "current",
		Results: []ScenarioResult{
			{Scenario: "test1", Status: StatusPassed, Observations: map[string]int{"P2": 1}},
		},
	}

	comparison2 := runner.Compare(currentBetter, baseline)

	if len(comparison2.Fixed) != 1 {
		t.Errorf("expected 1 fixed for reduced observations, got %d", len(comparison2.Fixed))
	}
}

func TestGetHighestSeverity(t *testing.T) {
	tests := []struct {
		observations map[string]int
		expected     string
	}{
		{map[string]int{"P0": 1, "P2": 2}, "P0"},
		{map[string]int{"P1": 1, "P3": 5}, "P1"},
		{map[string]int{"P3": 3}, "P3"},
		{map[string]int{}, "P3"}, // Default to lowest
		{nil, "P3"},
	}

	for _, tt := range tests {
		result := getHighestSeverity(tt.observations)
		if result != tt.expected {
			t.Errorf("getHighestSeverity(%v) = %s, expected %s",
				tt.observations, result, tt.expected)
		}
	}
}

func TestCountObservations(t *testing.T) {
	tests := []struct {
		observations map[string]int
		expected     int
	}{
		{map[string]int{"P0": 1, "P2": 2, "P3": 3}, 6},
		{map[string]int{"P1": 5}, 5},
		{map[string]int{}, 0},
		{nil, 0},
	}

	for _, tt := range tests {
		result := countObservations(tt.observations)
		if result != tt.expected {
			t.Errorf("countObservations(%v) = %d, expected %d",
				tt.observations, result, tt.expected)
		}
	}
}
