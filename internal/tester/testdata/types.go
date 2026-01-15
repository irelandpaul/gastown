// Package testdata provides test data isolation for AI User Testing.
// It generates unique emails, manages cleanup strategies, and ensures
// test data doesn't pollute across runs.
package testdata

import (
	"time"
)

// CleanupAction represents what to do with test data after a run.
type CleanupAction string

const (
	// CleanupDeleteAccount removes the test account and all associated data.
	CleanupDeleteAccount CleanupAction = "delete_account"

	// CleanupKeep preserves the test data for debugging.
	CleanupKeep CleanupAction = "keep"

	// CleanupMarkForReview preserves data and flags it for human review.
	CleanupMarkForReview CleanupAction = "mark_for_review"

	// CleanupJob queues cleanup for a background job (for crashes).
	CleanupJob CleanupAction = "cleanup_job"
)

// EmailInbox represents the email verification strategy.
type EmailInbox string

const (
	// EmailSkipVerification bypasses email verification entirely.
	EmailSkipVerification EmailInbox = "skip_verification"

	// EmailMailhog uses a local mailhog instance for email capture.
	EmailMailhog EmailInbox = "mailhog"

	// EmailReal uses a real email provider (e.g., for integration tests).
	EmailReal EmailInbox = "real"
)

// CleanupStrategy defines cleanup behavior for different test outcomes.
type CleanupStrategy struct {
	// OnSuccess is the action to take when test completes successfully.
	OnSuccess CleanupAction `json:"on_success" yaml:"on_success"`

	// OnFailure is the action to take when test fails.
	OnFailure CleanupAction `json:"on_failure" yaml:"on_failure"`

	// OnCrash is the action to take when test crashes unexpectedly.
	OnCrash CleanupAction `json:"on_crash" yaml:"on_crash"`
}

// IsolationConfig defines test data isolation settings.
type IsolationConfig struct {
	// UniqueSuffix appends a UUID to all created data for isolation.
	UniqueSuffix bool `json:"unique_suffix" yaml:"unique_suffix"`
}

// Config holds the complete test data configuration for a scenario.
type Config struct {
	// EmailPattern is the template for generating unique email addresses.
	// Supports placeholders: {scenario}, {run_id}, {timestamp}
	// Example: "test+{scenario}+{run_id}@screencoach.test"
	EmailPattern string `json:"email_pattern" yaml:"email_pattern"`

	// EmailInbox specifies how email verification is handled.
	EmailInbox EmailInbox `json:"email_inbox" yaml:"email_inbox"`

	// SeedAccount is a pre-created test account to use (optional).
	SeedAccount string `json:"seed_account,omitempty" yaml:"seed_account,omitempty"`

	// SeedData is pre-populated data to use (optional).
	SeedData map[string]interface{} `json:"seed_data,omitempty" yaml:"seed_data,omitempty"`

	// Cleanup defines cleanup behavior for different outcomes.
	Cleanup CleanupStrategy `json:"cleanup_strategy" yaml:"cleanup_strategy"`

	// Isolation defines isolation settings.
	Isolation IsolationConfig `json:"isolation" yaml:"isolation"`
}

// DefaultConfig returns the default test data configuration.
func DefaultConfig() Config {
	return Config{
		EmailPattern: "test+{scenario}+{run_id}@screencoach.test",
		EmailInbox:   EmailSkipVerification,
		Cleanup: CleanupStrategy{
			OnSuccess: CleanupDeleteAccount,
			OnFailure: CleanupMarkForReview,
			OnCrash:   CleanupJob,
		},
		Isolation: IsolationConfig{
			UniqueSuffix: true,
		},
	}
}

// RunContext holds the context for a specific test run.
type RunContext struct {
	// RunID is the unique identifier for this test run.
	RunID string `json:"run_id"`

	// Scenario is the scenario name being run.
	Scenario string `json:"scenario"`

	// StartedAt is when the run started.
	StartedAt time.Time `json:"started_at"`

	// Email is the generated unique email for this run.
	Email string `json:"email"`

	// UniqueSuffix is the UUID suffix for data isolation.
	UniqueSuffix string `json:"unique_suffix,omitempty"`

	// Config is the test data configuration used.
	Config Config `json:"config"`
}

// CleanupRecord tracks test data that may need cleanup.
type CleanupRecord struct {
	// RunID identifies the test run.
	RunID string `json:"run_id"`

	// Scenario is the scenario that created this data.
	Scenario string `json:"scenario"`

	// Email is the email used for the test account.
	Email string `json:"email"`

	// CreatedAt is when the test data was created.
	CreatedAt time.Time `json:"created_at"`

	// Status is the cleanup status (pending, completed, failed).
	Status string `json:"status"`

	// Action is the cleanup action to perform.
	Action CleanupAction `json:"action"`

	// MarkedForReview indicates if human review is needed.
	MarkedForReview bool `json:"marked_for_review"`

	// ReviewNotes contains notes from human review.
	ReviewNotes string `json:"review_notes,omitempty"`
}
