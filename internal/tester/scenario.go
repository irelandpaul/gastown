// Scenario YAML parsing types for AI User Testing.
//
// These types define the structure of scenario YAML files that configure
// AI persona tests. They are separate from the runtime execution types
// in types.go.
package tester

import "time"

// ScenarioConfig represents a parsed scenario YAML file.
// Scenarios define what an AI persona should accomplish and how to verify success.
type ScenarioConfig struct {
	// Scenario is the unique identifier/name for this test scenario.
	Scenario string `yaml:"scenario"`

	// Version enables breaking change detection and migration tooling.
	Version int `yaml:"version"`

	// Persona is the AI persona that will execute this scenario.
	// Examples: "sarah" (tech-comfortable parent), "miguel" (overwhelmed dad)
	Persona string `yaml:"persona"`

	// Goal describes what the persona should accomplish in user story format.
	// This is the primary instruction given to the AI agent.
	Goal string `yaml:"goal"`

	// SuccessCriteria defines specific, testable conditions for pass/fail.
	// Each criterion should be verifiable by the AI agent.
	SuccessCriteria []string `yaml:"success_criteria"`

	// Environment configures the target application.
	Environment ScenarioEnvironment `yaml:"environment"`

	// TestData configures test data isolation and cleanup.
	TestData *ScenarioTestData `yaml:"test_data,omitempty"`

	// WaitStrategies configures timing and synchronization.
	WaitStrategies *ScenarioWaitStrategies `yaml:"wait_strategies,omitempty"`

	// Recording configures artifact capture settings.
	Recording *ScenarioRecording `yaml:"recording,omitempty"`

	// Retry configures retry logic for infrastructure failures.
	Retry *ScenarioRetry `yaml:"retry,omitempty"`

	// Timeout sets the maximum duration for the scenario.
	// Defaults to 5 minutes if not specified.
	Timeout YAMLDuration `yaml:"timeout,omitempty"`

	// Tags allow categorizing and filtering scenarios.
	Tags []string `yaml:"tags,omitempty"`
}

// ScenarioEnvironment configures the target application for testing.
type ScenarioEnvironment struct {
	// URL is the starting URL for the test.
	URL string `yaml:"url"`

	// Viewport configures the browser viewport size.
	// Defaults to desktop (1280x720) if not specified.
	Viewport *ScenarioViewport `yaml:"viewport,omitempty"`

	// Device simulates a specific device (overrides viewport).
	// Examples: "iPhone 12", "Pixel 5", "iPad Pro"
	Device string `yaml:"device,omitempty"`
}

// ScenarioViewport defines browser viewport dimensions for YAML parsing.
type ScenarioViewport struct {
	Width  int `yaml:"width"`
	Height int `yaml:"height"`
}

// ScenarioTestData configures test data isolation and cleanup.
type ScenarioTestData struct {
	// SeedAccount provides pre-existing test account credentials.
	SeedAccount *ScenarioSeedAccount `yaml:"seed_account,omitempty"`

	// EmailPattern defines how unique test emails are generated.
	// Use {scenario} and {run_id} placeholders.
	// Example: "test+{scenario}+{run_id}@example.test"
	EmailPattern string `yaml:"email_pattern,omitempty"`

	// EmailInbox specifies the test email service.
	// Options: "mailhog", "mailinator", "skip_verification"
	EmailInbox string `yaml:"email_inbox,omitempty"`

	// CleanupStrategy defines how test data is handled after runs.
	CleanupStrategy *ScenarioCleanupStrategy `yaml:"cleanup_strategy,omitempty"`

	// Isolation configures data isolation for parallel runs.
	Isolation *ScenarioIsolation `yaml:"isolation,omitempty"`
}

// ScenarioSeedAccount provides pre-existing test account credentials.
type ScenarioSeedAccount struct {
	Email    string `yaml:"email"`
	Password string `yaml:"password"`
}

// ScenarioCleanupStrategy defines how test data is cleaned up.
type ScenarioCleanupStrategy struct {
	// OnSuccess action when test passes.
	// Options: "delete_account", "keep", "mark_for_review"
	OnSuccess string `yaml:"on_success,omitempty"`

	// OnFailure action when test fails.
	// Options: "keep", "mark_for_review", "delete_account"
	OnFailure string `yaml:"on_failure,omitempty"`

	// OnCrash action when test crashes unexpectedly.
	// Options: "cleanup_job", "mark_for_review", "keep"
	OnCrash string `yaml:"on_crash,omitempty"`
}

// ScenarioIsolation configures data isolation for parallel test runs.
type ScenarioIsolation struct {
	// UniqueSuffix appends UUID to all created data.
	UniqueSuffix bool `yaml:"unique_suffix,omitempty"`
}

// ScenarioWaitStrategies configures timing and synchronization for reliable tests.
type ScenarioWaitStrategies struct {
	// NetworkIdle waits for no pending network requests.
	NetworkIdle bool `yaml:"network_idle,omitempty"`

	// AnimationComplete waits for CSS transitions to finish.
	AnimationComplete bool `yaml:"animation_complete,omitempty"`

	// CustomSelectors are app-specific ready indicators.
	// Example: ["#app-loaded", "[data-ready=true]"]
	CustomSelectors []string `yaml:"custom_selectors,omitempty"`

	// MinLoadTime is minimum wait (ms) after navigation.
	MinLoadTime int `yaml:"min_load_time,omitempty"`
}

// ScenarioRecording configures artifact capture settings.
type ScenarioRecording struct {
	// Headed runs browser in headed mode (visible UI).
	// Defaults to false (headless).
	Headed bool `yaml:"headed,omitempty"`

	// HeadedVerification runs headed periodically (e.g., "weekly").
	HeadedVerification string `yaml:"headed_verification,omitempty"`

	// Video enables video recording of the test.
	// Defaults to true.
	Video *bool `yaml:"video,omitempty"`

	// Trace enables Playwright trace recording.
	// Defaults to true.
	Trace *bool `yaml:"trace,omitempty"`

	// Screenshots captures screenshots at key moments.
	// Defaults to true.
	Screenshots *bool `yaml:"screenshots,omitempty"`
}

// ScenarioRetry configures retry logic for infrastructure failures.
type ScenarioRetry struct {
	// MaxAttempts is the maximum number of retry attempts.
	// Defaults to 3.
	MaxAttempts int `yaml:"max_attempts,omitempty"`

	// OnErrors lists error types that trigger retry.
	// Options: "browser_crash", "timeout", "network_error"
	OnErrors []string `yaml:"on_errors,omitempty"`

	// NotOn lists error types that should not be retried.
	// Options: "test_failure", "blocked"
	NotOn []string `yaml:"not_on,omitempty"`

	// Backoff strategy for retry delays.
	// Options: "exponential", "linear", "fixed"
	Backoff string `yaml:"backoff,omitempty"`
}

// YAMLDuration is a wrapper for time.Duration that supports YAML unmarshaling.
type YAMLDuration time.Duration

// UnmarshalYAML implements yaml.Unmarshaler for YAMLDuration.
func (d *YAMLDuration) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var s string
	if err := unmarshal(&s); err != nil {
		return err
	}
	dur, err := time.ParseDuration(s)
	if err != nil {
		return err
	}
	*d = YAMLDuration(dur)
	return nil
}

// MarshalYAML implements yaml.Marshaler for YAMLDuration.
func (d YAMLDuration) MarshalYAML() (interface{}, error) {
	return time.Duration(d).String(), nil
}

// Duration returns the underlying time.Duration.
func (d YAMLDuration) Duration() time.Duration {
	return time.Duration(d)
}

// DefaultScenarioViewport returns the default desktop viewport.
func DefaultScenarioViewport() ScenarioViewport {
	return ScenarioViewport{Width: 1280, Height: 720}
}

// DefaultScenarioTimeout returns the default scenario timeout (5 minutes).
func DefaultScenarioTimeout() YAMLDuration {
	return YAMLDuration(5 * time.Minute)
}

// DefaultScenarioRetry returns the default retry configuration.
func DefaultScenarioRetry() *ScenarioRetry {
	return &ScenarioRetry{
		MaxAttempts: 3,
		OnErrors:    []string{"browser_crash", "timeout", "network_error"},
		NotOn:       []string{"test_failure", "blocked"},
		Backoff:     "exponential",
	}
}

// DefaultScenarioRecording returns the default recording configuration.
func DefaultScenarioRecording() *ScenarioRecording {
	t := true
	return &ScenarioRecording{
		Headed:      false,
		Video:       &t,
		Trace:       &t,
		Screenshots: &t,
	}
}

// ToViewport converts ScenarioViewport to the runtime Viewport type.
func (v *ScenarioViewport) ToViewport() *Viewport {
	if v == nil {
		return nil
	}
	return &Viewport{
		Width:  v.Width,
		Height: v.Height,
	}
}
