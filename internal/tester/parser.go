package tester

import (
	"fmt"
	"net/url"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// ParseScenarioFile reads and parses a scenario YAML file.
func ParseScenarioFile(path string) (*ScenarioConfig, error) {
	data, err := os.ReadFile(path) //nolint:gosec // G304: path is from trusted scenario directory
	if err != nil {
		return nil, fmt.Errorf("reading scenario file: %w", err)
	}
	return ParseScenario(data)
}

// ParseScenario parses scenario YAML content from bytes.
func ParseScenario(data []byte) (*ScenarioConfig, error) {
	var s ScenarioConfig
	if err := yaml.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("parsing YAML: %w", err)
	}

	// Apply defaults
	s.applyDefaults()

	if err := s.Validate(); err != nil {
		return nil, err
	}

	return &s, nil
}

// applyDefaults sets default values for optional fields.
func (s *ScenarioConfig) applyDefaults() {
	// Default timeout
	if s.Timeout == 0 {
		s.Timeout = DefaultScenarioTimeout()
	}

	// Default viewport
	if s.Environment.Viewport == nil && s.Environment.Device == "" {
		v := DefaultScenarioViewport()
		s.Environment.Viewport = &v
	}

	// Default recording settings
	if s.Recording == nil {
		s.Recording = DefaultScenarioRecording()
	} else {
		// Fill in nil booleans with defaults
		t := true
		if s.Recording.Video == nil {
			s.Recording.Video = &t
		}
		if s.Recording.Trace == nil {
			s.Recording.Trace = &t
		}
		if s.Recording.Screenshots == nil {
			s.Recording.Screenshots = &t
		}
	}

	// Default retry settings
	if s.Retry == nil {
		s.Retry = DefaultScenarioRetry()
	} else {
		if s.Retry.MaxAttempts == 0 {
			s.Retry.MaxAttempts = 3
		}
		if s.Retry.Backoff == "" {
			s.Retry.Backoff = "exponential"
		}
		if len(s.Retry.OnErrors) == 0 {
			s.Retry.OnErrors = []string{"browser_crash", "timeout", "network_error"}
		}
		if len(s.Retry.NotOn) == 0 {
			s.Retry.NotOn = []string{"test_failure", "blocked"}
		}
	}

	// Default cleanup strategy
	if s.TestData != nil && s.TestData.CleanupStrategy == nil {
		s.TestData.CleanupStrategy = &ScenarioCleanupStrategy{
			OnSuccess: "delete_account",
			OnFailure: "mark_for_review",
			OnCrash:   "cleanup_job",
		}
	}
}

// Validate checks that the scenario has all required fields and valid structure.
func (s *ScenarioConfig) Validate() error {
	var errs []string

	// Required fields
	if s.Scenario == "" {
		errs = append(errs, "scenario field is required")
	}

	if s.Persona == "" {
		errs = append(errs, "persona field is required")
	}

	if s.Goal == "" {
		errs = append(errs, "goal field is required")
	}

	if len(s.SuccessCriteria) == 0 {
		errs = append(errs, "success_criteria field is required (at least one criterion)")
	}

	// Environment validation
	if err := s.validateEnvironment(); err != nil {
		errs = append(errs, err.Error())
	}

	// Test data validation
	if s.TestData != nil {
		if err := s.validateTestData(); err != nil {
			errs = append(errs, err.Error())
		}
	}

	// Wait strategies validation
	if s.WaitStrategies != nil {
		if err := s.validateWaitStrategies(); err != nil {
			errs = append(errs, err.Error())
		}
	}

	// Retry validation
	if s.Retry != nil {
		if err := s.validateRetry(); err != nil {
			errs = append(errs, err.Error())
		}
	}

	// Recording validation
	if s.Recording != nil {
		if err := s.validateRecording(); err != nil {
			errs = append(errs, err.Error())
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("scenario validation failed:\n  - %s", strings.Join(errs, "\n  - "))
	}

	return nil
}

func (s *ScenarioConfig) validateEnvironment() error {
	if s.Environment.URL == "" {
		return fmt.Errorf("environment.url is required")
	}

	// Validate URL format
	u, err := url.Parse(s.Environment.URL)
	if err != nil {
		return fmt.Errorf("environment.url is invalid: %w", err)
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("environment.url must use http or https scheme")
	}

	if u.Host == "" {
		return fmt.Errorf("environment.url must have a host")
	}

	// Validate viewport if specified
	if s.Environment.Viewport != nil {
		if s.Environment.Viewport.Width <= 0 {
			return fmt.Errorf("environment.viewport.width must be positive")
		}
		if s.Environment.Viewport.Height <= 0 {
			return fmt.Errorf("environment.viewport.height must be positive")
		}
	}

	// Cannot specify both viewport and device
	if s.Environment.Viewport != nil && s.Environment.Device != "" {
		return fmt.Errorf("cannot specify both environment.viewport and environment.device")
	}

	return nil
}

func (s *ScenarioConfig) validateTestData() error {
	td := s.TestData

	// Validate email inbox options
	if td.EmailInbox != "" {
		validInboxes := map[string]bool{
			"mailhog":           true,
			"mailinator":        true,
			"skip_verification": true,
		}
		if !validInboxes[td.EmailInbox] {
			return fmt.Errorf("test_data.email_inbox must be one of: mailhog, mailinator, skip_verification")
		}
	}

	// Validate cleanup strategy options
	if td.CleanupStrategy != nil {
		cs := td.CleanupStrategy

		validOnSuccess := map[string]bool{
			"delete_account":  true,
			"keep":            true,
			"mark_for_review": true,
		}
		if cs.OnSuccess != "" && !validOnSuccess[cs.OnSuccess] {
			return fmt.Errorf("test_data.cleanup_strategy.on_success must be one of: delete_account, keep, mark_for_review")
		}

		validOnFailure := map[string]bool{
			"keep":            true,
			"mark_for_review": true,
			"delete_account":  true,
		}
		if cs.OnFailure != "" && !validOnFailure[cs.OnFailure] {
			return fmt.Errorf("test_data.cleanup_strategy.on_failure must be one of: keep, mark_for_review, delete_account")
		}

		validOnCrash := map[string]bool{
			"cleanup_job":     true,
			"mark_for_review": true,
			"keep":            true,
		}
		if cs.OnCrash != "" && !validOnCrash[cs.OnCrash] {
			return fmt.Errorf("test_data.cleanup_strategy.on_crash must be one of: cleanup_job, mark_for_review, keep")
		}
	}

	return nil
}

func (s *ScenarioConfig) validateWaitStrategies() error {
	ws := s.WaitStrategies

	// Validate min_load_time is reasonable
	if ws.MinLoadTime < 0 {
		return fmt.Errorf("wait_strategies.min_load_time cannot be negative")
	}
	if ws.MinLoadTime > 60000 {
		return fmt.Errorf("wait_strategies.min_load_time cannot exceed 60000ms (1 minute)")
	}

	return nil
}

func (s *ScenarioConfig) validateRetry() error {
	r := s.Retry

	// Validate max_attempts
	if r.MaxAttempts < 1 {
		return fmt.Errorf("retry.max_attempts must be at least 1")
	}
	if r.MaxAttempts > 10 {
		return fmt.Errorf("retry.max_attempts cannot exceed 10")
	}

	// Validate backoff strategy
	validBackoff := map[string]bool{
		"exponential": true,
		"linear":      true,
		"fixed":       true,
	}
	if r.Backoff != "" && !validBackoff[r.Backoff] {
		return fmt.Errorf("retry.backoff must be one of: exponential, linear, fixed")
	}

	// Validate error types
	validErrors := map[string]bool{
		"browser_crash": true,
		"timeout":       true,
		"network_error": true,
		"test_failure":  true,
		"blocked":       true,
	}

	for _, e := range r.OnErrors {
		if !validErrors[e] {
			return fmt.Errorf("retry.on_errors contains invalid error type: %s", e)
		}
	}

	for _, e := range r.NotOn {
		if !validErrors[e] {
			return fmt.Errorf("retry.not_on contains invalid error type: %s", e)
		}
	}

	return nil
}

func (s *ScenarioConfig) validateRecording() error {
	r := s.Recording

	// Validate headed_verification schedule
	if r.HeadedVerification != "" {
		validSchedules := map[string]bool{
			"daily":   true,
			"weekly":  true,
			"monthly": true,
			"never":   true,
			"always":  true,
		}
		if !validSchedules[r.HeadedVerification] {
			return fmt.Errorf("recording.headed_verification must be one of: daily, weekly, monthly, never, always")
		}
	}

	return nil
}

// IsRetryable returns true if the given error type should trigger a retry.
func (s *ScenarioConfig) IsRetryable(errorType string) bool {
	if s.Retry == nil {
		return false
	}

	// Check not_on list first
	for _, e := range s.Retry.NotOn {
		if e == errorType {
			return false
		}
	}

	// Check on_errors list
	for _, e := range s.Retry.OnErrors {
		if e == errorType {
			return true
		}
	}

	return false
}

// ShouldRunHeaded returns true if this run should use headed mode.
// It checks the headed flag and headed_verification schedule.
func (s *ScenarioConfig) ShouldRunHeaded() bool {
	if s.Recording == nil {
		return false
	}

	// Always headed
	if s.Recording.Headed {
		return true
	}

	// Check verification schedule (implementation would check last headed run date)
	// For now, just return the headed flag
	return false
}

