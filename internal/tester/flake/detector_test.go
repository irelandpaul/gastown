package flake

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewDetector(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "flake.json")

	detector, err := NewDetector(storagePath, DefaultConfig())
	if err != nil {
		t.Fatalf("NewDetector failed: %v", err)
	}

	if detector == nil {
		t.Fatal("NewDetector returned nil")
	}
}

func TestRecordRun(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "flake.json")

	detector, err := NewDetector(storagePath, DefaultConfig())
	if err != nil {
		t.Fatalf("NewDetector failed: %v", err)
	}

	// Record a passing run
	actions, err := detector.RecordRun("test-scenario", RunRecord{
		Timestamp: time.Now(),
		Outcome:   OutcomePass,
		Duration:  5 * time.Second,
	})
	if err != nil {
		t.Fatalf("RecordRun failed: %v", err)
	}

	// Should not trigger any actions (not enough runs)
	if len(actions) != 0 {
		t.Errorf("Expected 0 actions, got %d", len(actions))
	}

	// Verify history was recorded
	history := detector.GetHistory("test-scenario")
	if history == nil {
		t.Fatal("Expected history to be recorded")
	}
	if history.TotalRuns != 1 {
		t.Errorf("Expected TotalRuns=1, got %d", history.TotalRuns)
	}
	if history.TotalPasses != 1 {
		t.Errorf("Expected TotalPasses=1, got %d", history.TotalPasses)
	}
}

func TestFlakeDetection(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "flake.json")

	config := DefaultConfig()
	config.WindowSize = 5
	config.FlakeThreshold = 0.3 // 30% failure rate
	config.MinRuns = 3
	config.AutoQuarantine = false // Disable auto-quarantine for this test

	detector, err := NewDetector(storagePath, config)
	if err != nil {
		t.Fatalf("NewDetector failed: %v", err)
	}

	// Record 2 passes
	for i := 0; i < 2; i++ {
		detector.RecordRun("flaky-test", RunRecord{
			Timestamp: time.Now(),
			Outcome:   OutcomePass,
		})
	}

	// Not flaky yet (only 2 runs)
	metrics := detector.GetMetrics("flaky-test")
	if metrics.IsFlaky {
		t.Error("Expected test not to be flaky with only 2 runs")
	}

	// Record 2 failures (2/4 = 50% failure rate)
	for i := 0; i < 2; i++ {
		detector.RecordRun("flaky-test", RunRecord{
			Timestamp: time.Now(),
			Outcome:   OutcomeFail,
		})
	}

	// Should be flaky now
	metrics = detector.GetMetrics("flaky-test")
	if !metrics.IsFlaky {
		t.Error("Expected test to be flaky with 50% failure rate")
	}
	if metrics.FlakeRate < 0.4 || metrics.FlakeRate > 0.6 {
		t.Errorf("Expected flake rate ~50%%, got %.0f%%", metrics.FlakeRate*100)
	}
}

func TestAutoQuarantine(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "flake.json")

	config := DefaultConfig()
	config.WindowSize = 5
	config.FlakeThreshold = 0.3
	config.MinRuns = 3
	config.AutoQuarantine = true

	detector, err := NewDetector(storagePath, config)
	if err != nil {
		t.Fatalf("NewDetector failed: %v", err)
	}

	// Record enough failures to trigger auto-quarantine
	// 3 failures out of 3 runs = 100% failure rate
	for i := 0; i < 3; i++ {
		actions, err := detector.RecordRun("auto-quarantine-test", RunRecord{
			Timestamp: time.Now(),
			Outcome:   OutcomeFail,
		})
		if err != nil {
			t.Fatalf("RecordRun failed: %v", err)
		}

		// On the 3rd run, should trigger quarantine
		if i == 2 {
			if len(actions) != 1 {
				t.Errorf("Expected 1 quarantine action, got %d", len(actions))
			} else if actions[0].Action != "quarantine" {
				t.Errorf("Expected 'quarantine' action, got %s", actions[0].Action)
			}
		}
	}

	// Verify scenario is quarantined
	if !detector.IsQuarantined("auto-quarantine-test") {
		t.Error("Expected scenario to be quarantined")
	}
}

func TestManualQuarantine(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "flake.json")

	detector, err := NewDetector(storagePath, DefaultConfig())
	if err != nil {
		t.Fatalf("NewDetector failed: %v", err)
	}

	// Quarantine a scenario
	err = detector.Quarantine("manual-test", "Manual quarantine for testing")
	if err != nil {
		t.Fatalf("Quarantine failed: %v", err)
	}

	// Verify it's quarantined
	if !detector.IsQuarantined("manual-test") {
		t.Error("Expected scenario to be quarantined")
	}

	// Get entry and verify
	entry := detector.GetQuarantineEntry("manual-test")
	if entry == nil {
		t.Fatal("Expected quarantine entry")
	}
	if entry.Reason != "Manual quarantine for testing" {
		t.Errorf("Unexpected reason: %s", entry.Reason)
	}
	if entry.AutoQuarantined {
		t.Error("Expected AutoQuarantined to be false")
	}

	// Unquarantine
	err = detector.Unquarantine("manual-test")
	if err != nil {
		t.Fatalf("Unquarantine failed: %v", err)
	}

	if detector.IsQuarantined("manual-test") {
		t.Error("Expected scenario to be unquarantined")
	}
}

func TestAutoUnquarantine(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "flake.json")

	config := DefaultConfig()
	config.WindowSize = 5
	config.FlakeThreshold = 0.3
	config.MinRuns = 3
	config.AutoQuarantine = true
	config.AutoUnquarantine = true
	config.UnquarantineThreshold = 0.8 // 80% success required

	detector, err := NewDetector(storagePath, config)
	if err != nil {
		t.Fatalf("NewDetector failed: %v", err)
	}

	// First, get it quarantined
	for i := 0; i < 3; i++ {
		detector.RecordRun("recovery-test", RunRecord{
			Timestamp: time.Now(),
			Outcome:   OutcomeFail,
		})
	}

	if !detector.IsQuarantined("recovery-test") {
		t.Fatal("Expected scenario to be quarantined")
	}

	// Now record passes to recover
	// Need 4 passes in a row to get to 80% (4/5 = 80%)
	var lastActions []QuarantineAction
	for i := 0; i < 4; i++ {
		actions, err := detector.RecordRun("recovery-test", RunRecord{
			Timestamp: time.Now(),
			Outcome:   OutcomePass,
		})
		if err != nil {
			t.Fatalf("RecordRun failed: %v", err)
		}
		lastActions = actions
	}

	// Should have been unquarantined
	if detector.IsQuarantined("recovery-test") {
		t.Error("Expected scenario to be unquarantined after recovery")
	}

	// Check that we got an unquarantine action
	foundUnquarantine := false
	for _, a := range lastActions {
		if a.Action == "unquarantine" {
			foundUnquarantine = true
			break
		}
	}
	if !foundUnquarantine {
		t.Error("Expected unquarantine action")
	}
}

func TestConsecutiveFailuresThreshold(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "flake.json")

	config := DefaultConfig()
	config.ConsecutiveFailuresThreshold = 3
	config.AutoQuarantine = true
	config.MinRuns = 3
	config.FlakeThreshold = 0.9 // Set high so flake rate won't trigger

	detector, err := NewDetector(storagePath, config)
	if err != nil {
		t.Fatalf("NewDetector failed: %v", err)
	}

	// Mix of passes and fails won't trigger consecutive threshold
	// Pattern: P, P, P, F, P, F (2 failures, but never 3 consecutive)
	detector.RecordRun("consecutive-test", RunRecord{Timestamp: time.Now(), Outcome: OutcomePass})
	detector.RecordRun("consecutive-test", RunRecord{Timestamp: time.Now(), Outcome: OutcomePass})
	detector.RecordRun("consecutive-test", RunRecord{Timestamp: time.Now(), Outcome: OutcomePass})
	detector.RecordRun("consecutive-test", RunRecord{Timestamp: time.Now(), Outcome: OutcomeFail})
	detector.RecordRun("consecutive-test", RunRecord{Timestamp: time.Now(), Outcome: OutcomePass})
	detector.RecordRun("consecutive-test", RunRecord{Timestamp: time.Now(), Outcome: OutcomeFail})

	if detector.IsQuarantined("consecutive-test") {
		t.Error("Should not be quarantined with non-consecutive failures")
	}

	// Verify metrics show non-consecutive state
	metrics := detector.GetMetrics("consecutive-test")
	if metrics.ConsecutiveFailures >= 3 {
		t.Errorf("Expected < 3 consecutive failures, got %d", metrics.ConsecutiveFailures)
	}

	// Clear and try 3 consecutive failures
	detector.ClearHistory("consecutive-test")

	for i := 0; i < 3; i++ {
		detector.RecordRun("consecutive-test", RunRecord{
			Timestamp: time.Now(),
			Outcome:   OutcomeFail,
		})
	}

	if !detector.IsQuarantined("consecutive-test") {
		t.Error("Expected quarantine after 3 consecutive failures")
	}
}

func TestPersistence(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "flake.json")

	// Create detector and add some data
	detector1, err := NewDetector(storagePath, DefaultConfig())
	if err != nil {
		t.Fatalf("NewDetector failed: %v", err)
	}

	detector1.RecordRun("persist-test", RunRecord{
		Timestamp: time.Now(),
		Outcome:   OutcomePass,
		Duration:  10 * time.Second,
	})
	detector1.Quarantine("quarantine-persist", "Test persistence")

	// Create a new detector and verify data was persisted
	detector2, err := NewDetector(storagePath, DefaultConfig())
	if err != nil {
		t.Fatalf("NewDetector failed: %v", err)
	}

	history := detector2.GetHistory("persist-test")
	if history == nil {
		t.Fatal("Expected history to be persisted")
	}
	if history.TotalRuns != 1 {
		t.Errorf("Expected TotalRuns=1, got %d", history.TotalRuns)
	}

	if !detector2.IsQuarantined("quarantine-persist") {
		t.Error("Expected quarantine to be persisted")
	}
}

func TestGetAllMetrics(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "flake.json")

	detector, err := NewDetector(storagePath, DefaultConfig())
	if err != nil {
		t.Fatalf("NewDetector failed: %v", err)
	}

	// Add runs for multiple scenarios
	scenarios := []string{"scenario-a", "scenario-b", "scenario-c"}
	for _, s := range scenarios {
		detector.RecordRun(s, RunRecord{
			Timestamp: time.Now(),
			Outcome:   OutcomePass,
		})
	}

	allMetrics := detector.GetAllMetrics()
	if len(allMetrics) != 3 {
		t.Errorf("Expected 3 metrics, got %d", len(allMetrics))
	}
}

func TestGetFlakyScenarios(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "flake.json")

	config := DefaultConfig()
	config.MinRuns = 2
	config.FlakeThreshold = 0.3
	config.AutoQuarantine = false

	detector, err := NewDetector(storagePath, config)
	if err != nil {
		t.Fatalf("NewDetector failed: %v", err)
	}

	// Add a stable scenario
	for i := 0; i < 3; i++ {
		detector.RecordRun("stable", RunRecord{Timestamp: time.Now(), Outcome: OutcomePass})
	}

	// Add a flaky scenario
	detector.RecordRun("flaky", RunRecord{Timestamp: time.Now(), Outcome: OutcomeFail})
	detector.RecordRun("flaky", RunRecord{Timestamp: time.Now(), Outcome: OutcomeFail})

	flaky := detector.GetFlakyScenarios()
	if len(flaky) != 1 {
		t.Errorf("Expected 1 flaky scenario, got %d", len(flaky))
	}
	if len(flaky) > 0 && flaky[0].Scenario != "flaky" {
		t.Errorf("Expected 'flaky' scenario, got %s", flaky[0].Scenario)
	}
}

func TestListQuarantined(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "flake.json")

	detector, err := NewDetector(storagePath, DefaultConfig())
	if err != nil {
		t.Fatalf("NewDetector failed: %v", err)
	}

	// Quarantine multiple scenarios
	detector.Quarantine("q1", "reason 1")
	time.Sleep(10 * time.Millisecond) // Ensure different timestamps
	detector.Quarantine("q2", "reason 2")
	time.Sleep(10 * time.Millisecond)
	detector.Quarantine("q3", "reason 3")

	entries := detector.ListQuarantined()
	if len(entries) != 3 {
		t.Errorf("Expected 3 entries, got %d", len(entries))
	}

	// Should be sorted by date (most recent first)
	if len(entries) >= 2 {
		if entries[0].QuarantinedAt.Before(entries[1].QuarantinedAt) {
			t.Error("Expected entries to be sorted by date descending")
		}
	}
}

func TestClearHistory(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "flake.json")

	detector, err := NewDetector(storagePath, DefaultConfig())
	if err != nil {
		t.Fatalf("NewDetector failed: %v", err)
	}

	// Add some history
	for i := 0; i < 5; i++ {
		detector.RecordRun("clear-test", RunRecord{
			Timestamp: time.Now(),
			Outcome:   OutcomePass,
		})
	}

	history := detector.GetHistory("clear-test")
	if history == nil || history.TotalRuns != 5 {
		t.Fatal("Expected 5 runs in history")
	}

	// Clear history
	err = detector.ClearHistory("clear-test")
	if err != nil {
		t.Fatalf("ClearHistory failed: %v", err)
	}

	history = detector.GetHistory("clear-test")
	if history != nil {
		t.Error("Expected history to be cleared")
	}
}

func TestMetricsCalculation(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "flake.json")

	config := DefaultConfig()
	config.WindowSize = 10

	detector, err := NewDetector(storagePath, config)
	if err != nil {
		t.Fatalf("NewDetector failed: %v", err)
	}

	// Record mixed outcomes with retries
	runs := []struct {
		outcome    RunOutcome
		retryCount int
		duration   time.Duration
	}{
		{OutcomePass, 0, 5 * time.Second},
		{OutcomePass, 1, 10 * time.Second},
		{OutcomeFail, 2, 15 * time.Second},
		{OutcomePass, 0, 5 * time.Second},
		{OutcomeError, 3, 20 * time.Second},
	}

	for _, r := range runs {
		detector.RecordRun("metrics-test", RunRecord{
			Timestamp:           time.Now(),
			Outcome:             r.outcome,
			RetryCount:          r.retryCount,
			Duration:            r.duration,
			InfrastructureError: r.outcome == OutcomeError,
		})
	}

	metrics := detector.GetMetrics("metrics-test")

	// Verify counts
	if metrics.WindowRuns != 5 {
		t.Errorf("Expected WindowRuns=5, got %d", metrics.WindowRuns)
	}
	if metrics.WindowPasses != 3 {
		t.Errorf("Expected WindowPasses=3, got %d", metrics.WindowPasses)
	}
	if metrics.WindowFailures != 1 {
		t.Errorf("Expected WindowFailures=1, got %d", metrics.WindowFailures)
	}
	if metrics.WindowErrors != 1 {
		t.Errorf("Expected WindowErrors=1, got %d", metrics.WindowErrors)
	}

	// Verify flake rate (2/5 = 40%)
	expectedFlakeRate := 0.4
	if metrics.FlakeRate < expectedFlakeRate-0.01 || metrics.FlakeRate > expectedFlakeRate+0.01 {
		t.Errorf("Expected FlakeRate=%.1f, got %.1f", expectedFlakeRate, metrics.FlakeRate)
	}

	// Verify average retries (0+1+2+0+3)/5 = 1.2
	expectedRetries := 1.2
	if metrics.AverageRetries < expectedRetries-0.01 || metrics.AverageRetries > expectedRetries+0.01 {
		t.Errorf("Expected AverageRetries=%.1f, got %.1f", expectedRetries, metrics.AverageRetries)
	}

	// Verify average duration (5+10+15+5+20)/5 = 11s
	expectedDuration := 11 * time.Second
	if metrics.AverageDuration != expectedDuration {
		t.Errorf("Expected AverageDuration=%v, got %v", expectedDuration, metrics.AverageDuration)
	}
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.WindowSize != 10 {
		t.Errorf("Expected WindowSize=10, got %d", config.WindowSize)
	}
	if config.FlakeThreshold != 0.3 {
		t.Errorf("Expected FlakeThreshold=0.3, got %f", config.FlakeThreshold)
	}
	if config.MinRuns != 3 {
		t.Errorf("Expected MinRuns=3, got %d", config.MinRuns)
	}
	if !config.AutoQuarantine {
		t.Error("Expected AutoQuarantine=true")
	}
	if config.AutoUnquarantine {
		t.Error("Expected AutoUnquarantine=false")
	}
}

func TestStorageCreatesDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	storagePath := filepath.Join(tmpDir, "nested", "dir", "flake.json")

	detector, err := NewDetector(storagePath, DefaultConfig())
	if err != nil {
		t.Fatalf("NewDetector failed: %v", err)
	}

	// Record something to trigger save
	detector.RecordRun("test", RunRecord{Timestamp: time.Now(), Outcome: OutcomePass})

	// Verify file was created
	if _, err := os.Stat(storagePath); os.IsNotExist(err) {
		t.Error("Expected storage file to be created")
	}
}
