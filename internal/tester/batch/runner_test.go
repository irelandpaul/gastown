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
