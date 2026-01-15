package testdata

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// CleanupManager handles test data cleanup across runs.
type CleanupManager struct {
	// dataDir is the directory where cleanup records are stored.
	dataDir string

	// mu protects concurrent access to cleanup records.
	mu sync.Mutex
}

// NewCleanupManager creates a new cleanup manager.
func NewCleanupManager(dataDir string) (*CleanupManager, error) {
	// Ensure the data directory exists
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create cleanup data directory: %w", err)
	}

	return &CleanupManager{
		dataDir: dataDir,
	}, nil
}

// RecordForCleanup registers test data that may need cleanup later.
func (m *CleanupManager) RecordForCleanup(ctx *RunContext, action CleanupAction) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	record := CleanupRecord{
		RunID:           ctx.RunID,
		Scenario:        ctx.Scenario,
		Email:           ctx.Email,
		CreatedAt:       ctx.StartedAt,
		Status:          "pending",
		Action:          action,
		MarkedForReview: action == CleanupMarkForReview,
	}

	return m.saveRecord(record)
}

// HandleTestOutcome determines and records the appropriate cleanup action.
func (m *CleanupManager) HandleTestOutcome(ctx *RunContext, outcome TestOutcome) error {
	var action CleanupAction

	switch outcome {
	case OutcomeSuccess:
		action = ctx.Config.Cleanup.OnSuccess
	case OutcomeFailure:
		action = ctx.Config.Cleanup.OnFailure
	case OutcomeCrash:
		action = ctx.Config.Cleanup.OnCrash
	default:
		action = CleanupMarkForReview
	}

	return m.RecordForCleanup(ctx, action)
}

// TestOutcome represents the outcome of a test run.
type TestOutcome string

const (
	// OutcomeSuccess means the test completed successfully.
	OutcomeSuccess TestOutcome = "success"

	// OutcomeFailure means the test failed (criteria not met).
	OutcomeFailure TestOutcome = "failure"

	// OutcomeCrash means the test crashed unexpectedly.
	OutcomeCrash TestOutcome = "crash"
)

// saveRecord writes a cleanup record to disk.
func (m *CleanupManager) saveRecord(record CleanupRecord) error {
	filename := fmt.Sprintf("%s_%s.json", record.Scenario, record.RunID)
	filepath := filepath.Join(m.dataDir, filename)

	data, err := json.MarshalIndent(record, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal cleanup record: %w", err)
	}

	if err := os.WriteFile(filepath, data, 0644); err != nil {
		return fmt.Errorf("failed to write cleanup record: %w", err)
	}

	return nil
}

// GetPendingCleanups returns all pending cleanup records.
func (m *CleanupManager) GetPendingCleanups() ([]CleanupRecord, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	var records []CleanupRecord

	files, err := filepath.Glob(filepath.Join(m.dataDir, "*.json"))
	if err != nil {
		return nil, fmt.Errorf("failed to list cleanup records: %w", err)
	}

	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			continue // Skip files that can't be read
		}

		var record CleanupRecord
		if err := json.Unmarshal(data, &record); err != nil {
			continue // Skip invalid JSON
		}

		if record.Status == "pending" {
			records = append(records, record)
		}
	}

	return records, nil
}

// MarkCompleted marks a cleanup record as completed.
func (m *CleanupManager) MarkCompleted(runID, scenario string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	filename := fmt.Sprintf("%s_%s.json", scenario, runID)
	filepath := filepath.Join(m.dataDir, filename)

	data, err := os.ReadFile(filepath)
	if err != nil {
		return fmt.Errorf("failed to read cleanup record: %w", err)
	}

	var record CleanupRecord
	if err := json.Unmarshal(data, &record); err != nil {
		return fmt.Errorf("failed to parse cleanup record: %w", err)
	}

	record.Status = "completed"
	return m.saveRecord(record)
}

// MarkFailed marks a cleanup record as failed.
func (m *CleanupManager) MarkFailed(runID, scenario, reason string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	filename := fmt.Sprintf("%s_%s.json", scenario, runID)
	filepath := filepath.Join(m.dataDir, filename)

	data, err := os.ReadFile(filepath)
	if err != nil {
		return fmt.Errorf("failed to read cleanup record: %w", err)
	}

	var record CleanupRecord
	if err := json.Unmarshal(data, &record); err != nil {
		return fmt.Errorf("failed to parse cleanup record: %w", err)
	}

	record.Status = "failed"
	record.ReviewNotes = reason
	return m.saveRecord(record)
}

// GetRecordsForReview returns all records marked for human review.
func (m *CleanupManager) GetRecordsForReview() ([]CleanupRecord, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	var records []CleanupRecord

	files, err := filepath.Glob(filepath.Join(m.dataDir, "*.json"))
	if err != nil {
		return nil, fmt.Errorf("failed to list cleanup records: %w", err)
	}

	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			continue
		}

		var record CleanupRecord
		if err := json.Unmarshal(data, &record); err != nil {
			continue
		}

		if record.MarkedForReview {
			records = append(records, record)
		}
	}

	return records, nil
}

// CleanupOlderThan removes cleanup records older than the specified duration.
func (m *CleanupManager) CleanupOlderThan(age time.Duration) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	cutoff := time.Now().Add(-age)
	count := 0

	files, err := filepath.Glob(filepath.Join(m.dataDir, "*.json"))
	if err != nil {
		return 0, fmt.Errorf("failed to list cleanup records: %w", err)
	}

	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			continue
		}

		var record CleanupRecord
		if err := json.Unmarshal(data, &record); err != nil {
			continue
		}

		// Only remove completed records that are old enough
		if record.Status == "completed" && record.CreatedAt.Before(cutoff) {
			if err := os.Remove(file); err == nil {
				count++
			}
		}
	}

	return count, nil
}
