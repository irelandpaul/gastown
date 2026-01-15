// Package artifacts provides artifact recording for AI User Testing.
// It manages video recordings, Playwright traces, and screenshots
// captured during test runs.
package artifacts

import (
	"time"
)

// ArtifactType represents the type of recorded artifact.
type ArtifactType string

const (
	// ArtifactVideo is a full session video recording (.webm).
	ArtifactVideo ArtifactType = "video"

	// ArtifactTrace is an interactive Playwright trace (.zip).
	ArtifactTrace ArtifactType = "trace"

	// ArtifactScreenshot is a point-in-time screenshot (.png).
	ArtifactScreenshot ArtifactType = "screenshot"

	// ArtifactObservations is the structured observations JSON.
	ArtifactObservations ArtifactType = "observations"

	// ArtifactSummary is the human-readable summary markdown.
	ArtifactSummary ArtifactType = "summary"
)

// ScreenshotTrigger represents what triggered a screenshot.
type ScreenshotTrigger string

const (
	// TriggerFailure means the screenshot was taken due to a test failure.
	TriggerFailure ScreenshotTrigger = "failure"

	// TriggerConfusion means the screenshot was taken when the agent felt confused.
	TriggerConfusion ScreenshotTrigger = "confusion"

	// TriggerOnDemand means the screenshot was taken on-demand by the agent.
	TriggerOnDemand ScreenshotTrigger = "on_demand"

	// TriggerStep means the screenshot was taken at a test step.
	TriggerStep ScreenshotTrigger = "step"
)

// HeadedVerification represents when to run headed verification.
type HeadedVerification string

const (
	HeadedDaily   HeadedVerification = "daily"
	HeadedWeekly  HeadedVerification = "weekly"
	HeadedMonthly HeadedVerification = "monthly"
)

// ScreenshotConfig defines when screenshots should be taken.
type ScreenshotConfig struct {
	// OnFailure takes a screenshot when the test fails.
	OnFailure bool `json:"on_failure" yaml:"on_failure"`

	// OnConfusion takes a screenshot when the agent feels confused.
	OnConfusion bool `json:"on_confusion" yaml:"on_confusion"`

	// OnDemand allows the agent to take screenshots anytime.
	OnDemand bool `json:"on_demand" yaml:"on_demand"`

	// RequireReview marks screenshots for human validation.
	RequireReview bool `json:"require_review" yaml:"require_review"`
}

// RecordingConfig defines what artifacts to record during a test.
type RecordingConfig struct {
	// Video enables full session video recording.
	Video bool `json:"video" yaml:"video"`

	// Trace enables Playwright trace recording.
	Trace bool `json:"trace" yaml:"trace"`

	// Screenshots configures screenshot behavior.
	Screenshots ScreenshotConfig `json:"screenshots" yaml:"screenshots"`

	// Headed runs the browser in visible mode (not headless).
	Headed bool `json:"headed" yaml:"headed"`

	// HeadedVerification schedules periodic headed runs for verification.
	HeadedVerification HeadedVerification `json:"headed_verification,omitempty" yaml:"headed_verification,omitempty"`
}

// DefaultRecordingConfig returns the default recording configuration.
func DefaultRecordingConfig() RecordingConfig {
	return RecordingConfig{
		Video: true,
		Trace: true,
		Screenshots: ScreenshotConfig{
			OnFailure:     true,
			OnConfusion:   true,
			OnDemand:      true,
			RequireReview: true,
		},
		Headed:             false,
		HeadedVerification: "",
	}
}

// Artifact represents a single recorded artifact.
type Artifact struct {
	// Type is the artifact type.
	Type ArtifactType `json:"type"`

	// Path is the file path relative to the run directory.
	Path string `json:"path"`

	// AbsolutePath is the full absolute path to the artifact.
	AbsolutePath string `json:"absolute_path"`

	// CreatedAt is when the artifact was created.
	CreatedAt time.Time `json:"created_at"`

	// SizeBytes is the file size in bytes.
	SizeBytes int64 `json:"size_bytes"`

	// Metadata contains artifact-specific metadata.
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// Screenshot represents a screenshot artifact with additional context.
type Screenshot struct {
	Artifact

	// Name is the descriptive name of the screenshot.
	Name string `json:"name"`

	// Trigger is what caused this screenshot to be taken.
	Trigger ScreenshotTrigger `json:"trigger"`

	// Timestamp is the time offset from test start (e.g., "00:23").
	Timestamp string `json:"timestamp"`

	// Location is where in the app the screenshot was taken.
	Location string `json:"location,omitempty"`

	// Description explains what the screenshot shows.
	Description string `json:"description,omitempty"`

	// RequiresReview indicates if human review is needed.
	RequiresReview bool `json:"requires_review"`

	// Reviewed indicates if the screenshot has been reviewed.
	Reviewed bool `json:"reviewed"`

	// ReviewNotes contains notes from human review.
	ReviewNotes string `json:"review_notes,omitempty"`
}

// RunArtifacts holds all artifacts for a single test run.
type RunArtifacts struct {
	// RunID identifies the test run.
	RunID string `json:"run_id"`

	// Scenario is the scenario that was run.
	Scenario string `json:"scenario"`

	// RunDir is the directory containing all artifacts.
	RunDir string `json:"run_dir"`

	// StartedAt is when the test run started.
	StartedAt time.Time `json:"started_at"`

	// CompletedAt is when the test run completed.
	CompletedAt *time.Time `json:"completed_at,omitempty"`

	// Video is the video recording artifact (if enabled).
	Video *Artifact `json:"video,omitempty"`

	// Trace is the Playwright trace artifact (if enabled).
	Trace *Artifact `json:"trace,omitempty"`

	// Screenshots are all screenshots taken during the run.
	Screenshots []Screenshot `json:"screenshots,omitempty"`

	// Observations is the observations.json artifact.
	Observations *Artifact `json:"observations,omitempty"`

	// Summary is the summary.md artifact.
	Summary *Artifact `json:"summary,omitempty"`

	// Config is the recording configuration used.
	Config RecordingConfig `json:"config"`
}

// ArtifactManifest is saved to disk to track all artifacts for a run.
type ArtifactManifest struct {
	// Version is the manifest schema version.
	Version int `json:"version"`

	// RunArtifacts contains all artifact information.
	RunArtifacts

	// TotalSizeBytes is the total size of all artifacts.
	TotalSizeBytes int64 `json:"total_size_bytes"`
}
