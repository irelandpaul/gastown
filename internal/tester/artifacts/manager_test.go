package artifacts

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewManager(t *testing.T) {
	tmpDir := t.TempDir()

	manager, err := NewManager(tmpDir)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	if manager.baseDir != tmpDir {
		t.Errorf("expected baseDir %q, got %q", tmpDir, manager.baseDir)
	}
}

func TestCreateRunDir(t *testing.T) {
	tmpDir := t.TempDir()
	manager, _ := NewManager(tmpDir)

	runDir, err := manager.CreateRunDir("sarah_registers", "abc123")
	if err != nil {
		t.Fatalf("failed to create run dir: %v", err)
	}

	// Should contain the scenario name
	if !strings.Contains(runDir, "sarah_registers") {
		t.Errorf("expected runDir to contain scenario name: %s", runDir)
	}

	// Should contain run-<id>
	if !strings.Contains(runDir, "run-abc123") {
		t.Errorf("expected runDir to contain run-abc123: %s", runDir)
	}

	// Directory should exist
	if _, err := os.Stat(runDir); os.IsNotExist(err) {
		t.Error("run directory was not created")
	}

	// Screenshots subdirectory should exist
	screenshotsDir := filepath.Join(runDir, "screenshots")
	if _, err := os.Stat(screenshotsDir); os.IsNotExist(err) {
		t.Error("screenshots directory was not created")
	}
}

func TestInitRun(t *testing.T) {
	tmpDir := t.TempDir()
	manager, _ := NewManager(tmpDir)
	config := DefaultRecordingConfig()

	artifacts, err := manager.InitRun("test_scenario", "xyz789", config)
	if err != nil {
		t.Fatalf("failed to init run: %v", err)
	}

	if artifacts.RunID != "xyz789" {
		t.Errorf("expected RunID xyz789, got %s", artifacts.RunID)
	}

	if artifacts.Scenario != "test_scenario" {
		t.Errorf("expected scenario test_scenario, got %s", artifacts.Scenario)
	}

	if artifacts.RunDir == "" {
		t.Error("expected RunDir to be set")
	}
}

func TestPaths(t *testing.T) {
	tmpDir := t.TempDir()
	manager, _ := NewManager(tmpDir)

	runDir := filepath.Join(tmpDir, "test-run")

	// Video path
	videoPath := manager.VideoPath(runDir)
	if !strings.HasSuffix(videoPath, "video.webm") {
		t.Errorf("unexpected video path: %s", videoPath)
	}

	// Trace path
	tracePath := manager.TracePath(runDir)
	if !strings.HasSuffix(tracePath, "trace.zip") {
		t.Errorf("unexpected trace path: %s", tracePath)
	}

	// Screenshot path
	screenshotPath := manager.ScreenshotPath(runDir, "confusion-button")
	if !strings.Contains(screenshotPath, "screenshots") {
		t.Errorf("expected screenshot path to contain 'screenshots': %s", screenshotPath)
	}
	if !strings.HasSuffix(screenshotPath, ".png") {
		t.Errorf("expected screenshot path to end with .png: %s", screenshotPath)
	}

	// Observations path
	obsPath := manager.ObservationsPath(runDir)
	if !strings.HasSuffix(obsPath, "observations.json") {
		t.Errorf("unexpected observations path: %s", obsPath)
	}

	// Summary path
	summaryPath := manager.SummaryPath(runDir)
	if !strings.HasSuffix(summaryPath, "summary.md") {
		t.Errorf("unexpected summary path: %s", summaryPath)
	}
}

func TestRecordScreenshot(t *testing.T) {
	tmpDir := t.TempDir()
	manager, _ := NewManager(tmpDir)
	config := DefaultRecordingConfig()

	artifacts, _ := manager.InitRun("test_scenario", "run123", config)

	// Create a test screenshot file
	screenshotPath := manager.ScreenshotPath(artifacts.RunDir, "test-screenshot")
	if err := os.WriteFile(screenshotPath, []byte("fake png data"), 0644); err != nil {
		t.Fatalf("failed to create test screenshot: %v", err)
	}

	// Record the screenshot
	screenshot, err := manager.RecordScreenshot(
		artifacts,
		"test-screenshot",
		TriggerConfusion,
		"00:23",
		"homepage",
		"Button is hard to find",
	)
	if err != nil {
		t.Fatalf("failed to record screenshot: %v", err)
	}

	if screenshot.Name != "test-screenshot" {
		t.Errorf("expected name test-screenshot, got %s", screenshot.Name)
	}

	if screenshot.Trigger != TriggerConfusion {
		t.Errorf("expected trigger confusion, got %s", screenshot.Trigger)
	}

	if screenshot.Timestamp != "00:23" {
		t.Errorf("expected timestamp 00:23, got %s", screenshot.Timestamp)
	}

	if !screenshot.RequiresReview {
		t.Error("expected RequiresReview to be true with default config")
	}

	// Should be added to artifacts
	if len(artifacts.Screenshots) != 1 {
		t.Errorf("expected 1 screenshot, got %d", len(artifacts.Screenshots))
	}
}

func TestFinalizeRun(t *testing.T) {
	tmpDir := t.TempDir()
	manager, _ := NewManager(tmpDir)
	config := DefaultRecordingConfig()

	artifacts, _ := manager.InitRun("test_scenario", "final123", config)

	// Create some test files
	videoPath := manager.VideoPath(artifacts.RunDir)
	if err := os.WriteFile(videoPath, []byte("fake video"), 0644); err != nil {
		t.Fatalf("failed to create test video: %v", err)
	}
	manager.RecordVideo(artifacts)

	// Finalize
	if err := manager.FinalizeRun(artifacts); err != nil {
		t.Fatalf("failed to finalize run: %v", err)
	}

	// Check that manifest was created
	manifestPath := filepath.Join(artifacts.RunDir, "manifest.json")
	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		t.Error("manifest.json was not created")
	}

	// Load and verify manifest
	manifest, err := manager.LoadManifest(artifacts.RunDir)
	if err != nil {
		t.Fatalf("failed to load manifest: %v", err)
	}

	if manifest.RunID != "final123" {
		t.Errorf("expected RunID final123, got %s", manifest.RunID)
	}

	if manifest.CompletedAt == nil {
		t.Error("expected CompletedAt to be set")
	}
}

func TestDefaultRecordingConfig(t *testing.T) {
	config := DefaultRecordingConfig()

	if !config.Video {
		t.Error("expected Video to be true by default")
	}

	if !config.Trace {
		t.Error("expected Trace to be true by default")
	}

	if !config.Screenshots.OnFailure {
		t.Error("expected Screenshots.OnFailure to be true by default")
	}

	if !config.Screenshots.OnConfusion {
		t.Error("expected Screenshots.OnConfusion to be true by default")
	}

	if !config.Screenshots.OnDemand {
		t.Error("expected Screenshots.OnDemand to be true by default")
	}

	if !config.Screenshots.RequireReview {
		t.Error("expected Screenshots.RequireReview to be true by default")
	}

	if config.Headed {
		t.Error("expected Headed to be false by default")
	}
}

func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"simple", "simple"},
		{"with space", "with_space"},
		{"path/with/slashes", "path_with_slashes"},
		{"has:colon", "has_colon"},
		{"has*star", "has_star"},
		{"has?question", "has_question"},
		{"normal-dashes", "normal-dashes"},
		{"under_scores", "under_scores"},
	}

	for _, tt := range tests {
		result := sanitizeFilename(tt.input)
		if result != tt.expected {
			t.Errorf("sanitizeFilename(%q) = %q, expected %q", tt.input, result, tt.expected)
		}
	}
}
