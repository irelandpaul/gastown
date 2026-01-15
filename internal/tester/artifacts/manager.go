package artifacts

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Manager handles artifact recording and storage for test runs.
type Manager struct {
	// baseDir is the root directory for all test artifacts.
	baseDir string

	// mu protects concurrent access.
	mu sync.Mutex
}

// NewManager creates a new artifact manager.
func NewManager(baseDir string) (*Manager, error) {
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create artifacts base directory: %w", err)
	}

	return &Manager{
		baseDir: baseDir,
	}, nil
}

// CreateRunDir creates a new directory for a test run's artifacts.
// Returns the path to the run directory.
//
// Directory structure:
//
//	baseDir/
//	└── 2026-01-14/
//	    └── scenario_name/
//	        └── run-001/
func (m *Manager) CreateRunDir(scenario, runID string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Create date-based directory
	dateDir := time.Now().Format("2006-01-02")

	// Build the full path
	runDir := filepath.Join(m.baseDir, dateDir, scenario, fmt.Sprintf("run-%s", runID))

	// Create the directory and screenshots subdirectory
	screenshotsDir := filepath.Join(runDir, "screenshots")
	if err := os.MkdirAll(screenshotsDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create run directory: %w", err)
	}

	return runDir, nil
}

// InitRun initializes artifact tracking for a new test run.
func (m *Manager) InitRun(scenario, runID string, config RecordingConfig) (*RunArtifacts, error) {
	runDir, err := m.CreateRunDir(scenario, runID)
	if err != nil {
		return nil, err
	}

	return &RunArtifacts{
		RunID:       runID,
		Scenario:    scenario,
		RunDir:      runDir,
		StartedAt:   time.Now(),
		Screenshots: []Screenshot{},
		Config:      config,
	}, nil
}

// VideoPath returns the path where the video should be saved.
func (m *Manager) VideoPath(runDir string) string {
	return filepath.Join(runDir, "video.webm")
}

// TracePath returns the path where the trace should be saved.
func (m *Manager) TracePath(runDir string) string {
	return filepath.Join(runDir, "trace.zip")
}

// ScreenshotPath generates a path for a new screenshot.
func (m *Manager) ScreenshotPath(runDir, name string) string {
	// Sanitize the name for use as a filename
	safeName := sanitizeFilename(name)
	return filepath.Join(runDir, "screenshots", fmt.Sprintf("%s.png", safeName))
}

// ObservationsPath returns the path for the observations.json file.
func (m *Manager) ObservationsPath(runDir string) string {
	return filepath.Join(runDir, "observations.json")
}

// SummaryPath returns the path for the summary.md file.
func (m *Manager) SummaryPath(runDir string) string {
	return filepath.Join(runDir, "summary.md")
}

// RecordVideo registers that a video was recorded.
func (m *Manager) RecordVideo(artifacts *RunArtifacts) error {
	path := m.VideoPath(artifacts.RunDir)

	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("video file not found: %w", err)
	}

	artifacts.Video = &Artifact{
		Type:         ArtifactVideo,
		Path:         "video.webm",
		AbsolutePath: path,
		CreatedAt:    time.Now(),
		SizeBytes:    info.Size(),
	}

	return nil
}

// RecordTrace registers that a trace was recorded.
func (m *Manager) RecordTrace(artifacts *RunArtifacts) error {
	path := m.TracePath(artifacts.RunDir)

	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("trace file not found: %w", err)
	}

	artifacts.Trace = &Artifact{
		Type:         ArtifactTrace,
		Path:         "trace.zip",
		AbsolutePath: path,
		CreatedAt:    time.Now(),
		SizeBytes:    info.Size(),
	}

	return nil
}

// RecordScreenshot registers a screenshot taken during the test.
func (m *Manager) RecordScreenshot(artifacts *RunArtifacts, name string, trigger ScreenshotTrigger, timestamp, location, description string) (*Screenshot, error) {
	path := m.ScreenshotPath(artifacts.RunDir, name)
	relPath := filepath.Join("screenshots", fmt.Sprintf("%s.png", sanitizeFilename(name)))

	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("screenshot file not found: %w", err)
	}

	screenshot := Screenshot{
		Artifact: Artifact{
			Type:         ArtifactScreenshot,
			Path:         relPath,
			AbsolutePath: path,
			CreatedAt:    time.Now(),
			SizeBytes:    info.Size(),
		},
		Name:           name,
		Trigger:        trigger,
		Timestamp:      timestamp,
		Location:       location,
		Description:    description,
		RequiresReview: artifacts.Config.Screenshots.RequireReview,
		Reviewed:       false,
	}

	artifacts.Screenshots = append(artifacts.Screenshots, screenshot)
	return &screenshot, nil
}

// RecordObservations registers the observations.json artifact.
func (m *Manager) RecordObservations(artifacts *RunArtifacts) error {
	path := m.ObservationsPath(artifacts.RunDir)

	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("observations file not found: %w", err)
	}

	artifacts.Observations = &Artifact{
		Type:         ArtifactObservations,
		Path:         "observations.json",
		AbsolutePath: path,
		CreatedAt:    time.Now(),
		SizeBytes:    info.Size(),
	}

	return nil
}

// RecordSummary registers the summary.md artifact.
func (m *Manager) RecordSummary(artifacts *RunArtifacts) error {
	path := m.SummaryPath(artifacts.RunDir)

	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("summary file not found: %w", err)
	}

	artifacts.Summary = &Artifact{
		Type:         ArtifactSummary,
		Path:         "summary.md",
		AbsolutePath: path,
		CreatedAt:    time.Now(),
		SizeBytes:    info.Size(),
	}

	return nil
}

// FinalizeRun completes the artifact tracking and saves the manifest.
func (m *Manager) FinalizeRun(artifacts *RunArtifacts) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	artifacts.CompletedAt = &now

	// Calculate total size
	var totalSize int64
	if artifacts.Video != nil {
		totalSize += artifacts.Video.SizeBytes
	}
	if artifacts.Trace != nil {
		totalSize += artifacts.Trace.SizeBytes
	}
	for _, s := range artifacts.Screenshots {
		totalSize += s.SizeBytes
	}
	if artifacts.Observations != nil {
		totalSize += artifacts.Observations.SizeBytes
	}
	if artifacts.Summary != nil {
		totalSize += artifacts.Summary.SizeBytes
	}

	manifest := ArtifactManifest{
		Version:        1,
		RunArtifacts:   *artifacts,
		TotalSizeBytes: totalSize,
	}

	// Save manifest
	manifestPath := filepath.Join(artifacts.RunDir, "manifest.json")
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal manifest: %w", err)
	}

	if err := os.WriteFile(manifestPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write manifest: %w", err)
	}

	return nil
}

// LoadManifest loads the artifact manifest for a run.
func (m *Manager) LoadManifest(runDir string) (*ArtifactManifest, error) {
	manifestPath := filepath.Join(runDir, "manifest.json")

	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest: %w", err)
	}

	var manifest ArtifactManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("failed to parse manifest: %w", err)
	}

	return &manifest, nil
}

// ListRuns returns all run directories for a scenario on a given date.
func (m *Manager) ListRuns(date, scenario string) ([]string, error) {
	pattern := filepath.Join(m.baseDir, date, scenario, "run-*")
	return filepath.Glob(pattern)
}

// ListAllRuns returns all run directories across all dates and scenarios.
func (m *Manager) ListAllRuns() ([]string, error) {
	pattern := filepath.Join(m.baseDir, "*", "*", "run-*")
	return filepath.Glob(pattern)
}

// CleanupOlderThan removes artifacts older than the specified duration.
func (m *Manager) CleanupOlderThan(age time.Duration) (int, int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	cutoff := time.Now().Add(-age)
	var removed int
	var freedBytes int64

	runs, err := m.ListAllRuns()
	if err != nil {
		return 0, 0, err
	}

	for _, runDir := range runs {
		manifest, err := m.LoadManifest(runDir)
		if err != nil {
			continue // Skip runs without manifest
		}

		if manifest.StartedAt.Before(cutoff) {
			freedBytes += manifest.TotalSizeBytes
			if err := os.RemoveAll(runDir); err == nil {
				removed++
			}
		}
	}

	return removed, freedBytes, nil
}

// GetScreenshotsForReview returns all screenshots that need human review.
func (m *Manager) GetScreenshotsForReview() ([]Screenshot, error) {
	var needReview []Screenshot

	runs, err := m.ListAllRuns()
	if err != nil {
		return nil, err
	}

	for _, runDir := range runs {
		manifest, err := m.LoadManifest(runDir)
		if err != nil {
			continue
		}

		for _, s := range manifest.Screenshots {
			if s.RequiresReview && !s.Reviewed {
				needReview = append(needReview, s)
			}
		}
	}

	return needReview, nil
}

// sanitizeFilename replaces unsafe characters in a filename.
func sanitizeFilename(name string) string {
	// Replace problematic characters with underscores
	safe := make([]byte, len(name))
	for i, c := range []byte(name) {
		switch c {
		case '/', '\\', ':', '*', '?', '"', '<', '>', '|', ' ':
			safe[i] = '_'
		default:
			safe[i] = c
		}
	}
	return string(safe)
}
