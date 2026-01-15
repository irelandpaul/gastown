package tester

import (
	"strings"
	"testing"
	"time"
)

func TestParseScenario_MinimalScenario(t *testing.T) {
	yaml := `
scenario: register_new_parent
persona: sarah
goal: |
  As a tech-comfortable parent, I want to register for ScreenCoach
  so that I can start monitoring my child's screen time.
success_criteria:
  - Account created successfully
  - Welcome email received
environment:
  url: https://staging.example.com
`
	s, err := ParseScenario([]byte(yaml))
	if err != nil {
		t.Fatalf("ParseScenario failed: %v", err)
	}

	if s.Scenario != "register_new_parent" {
		t.Errorf("Scenario = %q, want %q", s.Scenario, "register_new_parent")
	}
	if s.Persona != "sarah" {
		t.Errorf("Persona = %q, want %q", s.Persona, "sarah")
	}
	if !strings.Contains(s.Goal, "tech-comfortable parent") {
		t.Errorf("Goal should contain 'tech-comfortable parent'")
	}
	if len(s.SuccessCriteria) != 2 {
		t.Errorf("SuccessCriteria length = %d, want 2", len(s.SuccessCriteria))
	}
	if s.Environment.URL != "https://staging.example.com" {
		t.Errorf("Environment.URL = %q, want %q", s.Environment.URL, "https://staging.example.com")
	}
}

func TestParseScenario_AppliesDefaults(t *testing.T) {
	yaml := `
scenario: test
persona: sarah
goal: Test goal
success_criteria:
  - Test criterion
environment:
  url: https://example.com
`
	s, err := ParseScenario([]byte(yaml))
	if err != nil {
		t.Fatalf("ParseScenario failed: %v", err)
	}

	// Check timeout default
	if s.Timeout.Duration() != 5*time.Minute {
		t.Errorf("Timeout = %v, want 5m", s.Timeout.Duration())
	}

	// Check viewport default
	if s.Environment.Viewport == nil {
		t.Fatal("Expected default viewport to be set")
	}
	if s.Environment.Viewport.Width != 1280 || s.Environment.Viewport.Height != 720 {
		t.Errorf("Viewport = %dx%d, want 1280x720",
			s.Environment.Viewport.Width, s.Environment.Viewport.Height)
	}

	// Check recording defaults
	if s.Recording == nil {
		t.Fatal("Expected default recording to be set")
	}
	if s.Recording.Headed {
		t.Error("Expected Headed to be false by default")
	}
	if s.Recording.Video == nil || !*s.Recording.Video {
		t.Error("Expected Video to be true by default")
	}
	if s.Recording.Trace == nil || !*s.Recording.Trace {
		t.Error("Expected Trace to be true by default")
	}

	// Check retry defaults
	if s.Retry == nil {
		t.Fatal("Expected default retry to be set")
	}
	if s.Retry.MaxAttempts != 3 {
		t.Errorf("Retry.MaxAttempts = %d, want 3", s.Retry.MaxAttempts)
	}
	if s.Retry.Backoff != "exponential" {
		t.Errorf("Retry.Backoff = %q, want %q", s.Retry.Backoff, "exponential")
	}
}

func TestParseScenario_FullScenario(t *testing.T) {
	yaml := `
scenario: registration_with_errors
version: 2
persona: miguel
goal: |
  Register for ScreenCoach, but make typical mistakes:
  - Enter invalid email first
  - Use weak password
  Observe error message clarity and recovery UX.
success_criteria:
  - Account created after correcting errors
  - Error messages are clear
  - Recovery path is obvious

environment:
  url: https://staging.example.com/signup
  viewport:
    width: 1920
    height: 1080

test_data:
  email_pattern: "test+{scenario}+{run_id}@example.test"
  email_inbox: mailhog
  cleanup_strategy:
    on_success: delete_account
    on_failure: mark_for_review
    on_crash: cleanup_job
  isolation:
    unique_suffix: true

wait_strategies:
  network_idle: true
  animation_complete: true
  custom_selectors:
    - "#app-loaded"
    - "[data-ready=true]"
  min_load_time: 2000

recording:
  headed: false
  headed_verification: weekly
  video: true
  trace: true
  screenshots: true

retry:
  max_attempts: 5
  on_errors:
    - browser_crash
    - timeout
  not_on:
    - test_failure
  backoff: linear

timeout: 10m
tags:
  - registration
  - error-recovery
`
	s, err := ParseScenario([]byte(yaml))
	if err != nil {
		t.Fatalf("ParseScenario failed: %v", err)
	}

	// Verify all fields parsed correctly
	if s.Version != 2 {
		t.Errorf("Version = %d, want 2", s.Version)
	}
	if s.Persona != "miguel" {
		t.Errorf("Persona = %q, want %q", s.Persona, "miguel")
	}
	if s.Environment.Viewport.Width != 1920 {
		t.Errorf("Viewport.Width = %d, want 1920", s.Environment.Viewport.Width)
	}
	if s.TestData.EmailPattern != "test+{scenario}+{run_id}@example.test" {
		t.Errorf("EmailPattern = %q, want test+{scenario}+{run_id}@example.test", s.TestData.EmailPattern)
	}
	if s.TestData.EmailInbox != "mailhog" {
		t.Errorf("EmailInbox = %q, want mailhog", s.TestData.EmailInbox)
	}
	if !s.WaitStrategies.NetworkIdle {
		t.Error("Expected NetworkIdle to be true")
	}
	if len(s.WaitStrategies.CustomSelectors) != 2 {
		t.Errorf("CustomSelectors length = %d, want 2", len(s.WaitStrategies.CustomSelectors))
	}
	if s.Retry.MaxAttempts != 5 {
		t.Errorf("Retry.MaxAttempts = %d, want 5", s.Retry.MaxAttempts)
	}
	if s.Timeout.Duration() != 10*time.Minute {
		t.Errorf("Timeout = %v, want 10m", s.Timeout.Duration())
	}
	if len(s.Tags) != 2 {
		t.Errorf("Tags length = %d, want 2", len(s.Tags))
	}
}

func TestParseScenario_MissingRequiredFields(t *testing.T) {
	tests := []struct {
		name    string
		yaml    string
		wantErr string
	}{
		{
			name:    "missing scenario",
			yaml:    `persona: sarah\ngoal: test\nsuccess_criteria: [ok]\nenvironment:\n  url: https://example.com`,
			wantErr: "scenario field is required",
		},
		{
			name:    "missing persona",
			yaml:    `scenario: test\ngoal: test\nsuccess_criteria: [ok]\nenvironment:\n  url: https://example.com`,
			wantErr: "persona field is required",
		},
		{
			name:    "missing goal",
			yaml:    `scenario: test\npersona: sarah\nsuccess_criteria: [ok]\nenvironment:\n  url: https://example.com`,
			wantErr: "goal field is required",
		},
		{
			name:    "missing success_criteria",
			yaml:    `scenario: test\npersona: sarah\ngoal: test\nenvironment:\n  url: https://example.com`,
			wantErr: "success_criteria field is required",
		},
		{
			name:    "missing environment url",
			yaml:    `scenario: test\npersona: sarah\ngoal: test\nsuccess_criteria: [ok]\nenvironment:\n  device: iPhone`,
			wantErr: "environment.url is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Convert \n to actual newlines
			yaml := strings.ReplaceAll(tt.yaml, `\n`, "\n")
			_, err := ParseScenario([]byte(yaml))
			if err == nil {
				t.Fatal("Expected error, got nil")
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("Error = %q, want to contain %q", err.Error(), tt.wantErr)
			}
		})
	}
}

func TestParseScenario_InvalidURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr string
	}{
		{
			name:    "invalid scheme",
			url:     "ftp://example.com",
			wantErr: "must use http or https scheme",
		},
		{
			name:    "missing host",
			url:     "https:///path",
			wantErr: "must have a host",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			yaml := `
scenario: test
persona: sarah
goal: test
success_criteria:
  - ok
environment:
  url: ` + tt.url

			_, err := ParseScenario([]byte(yaml))
			if err == nil {
				t.Fatal("Expected error, got nil")
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("Error = %q, want to contain %q", err.Error(), tt.wantErr)
			}
		})
	}
}

func TestParseScenario_InvalidRetry(t *testing.T) {
	tests := []struct {
		name    string
		retry   string
		wantErr string
	}{
		// Note: max_attempts: 0 defaults to 3, so there's no "too low" error case
		{
			name:    "max_attempts too high",
			retry:   "max_attempts: 100",
			wantErr: "cannot exceed 10",
		},
		{
			name:    "invalid backoff",
			retry:   "backoff: geometric",
			wantErr: "must be one of: exponential, linear, fixed",
		},
		{
			name:    "invalid error type",
			retry:   "on_errors: [invalid_error]",
			wantErr: "invalid error type: invalid_error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			yaml := `
scenario: test
persona: sarah
goal: test
success_criteria:
  - ok
environment:
  url: https://example.com
retry:
  ` + tt.retry

			_, err := ParseScenario([]byte(yaml))
			if err == nil {
				t.Fatal("Expected error, got nil")
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("Error = %q, want to contain %q", err.Error(), tt.wantErr)
			}
		})
	}
}

func TestParseScenario_InvalidEmailInbox(t *testing.T) {
	yaml := `
scenario: test
persona: sarah
goal: test
success_criteria:
  - ok
environment:
  url: https://example.com
test_data:
  email_inbox: gmail
`
	_, err := ParseScenario([]byte(yaml))
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if !strings.Contains(err.Error(), "must be one of: mailhog, mailinator, skip_verification") {
		t.Errorf("Error = %q, want to contain email inbox options", err.Error())
	}
}

func TestParseScenario_InvalidWaitStrategies(t *testing.T) {
	yaml := `
scenario: test
persona: sarah
goal: test
success_criteria:
  - ok
environment:
  url: https://example.com
wait_strategies:
  min_load_time: -100
`
	_, err := ParseScenario([]byte(yaml))
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if !strings.Contains(err.Error(), "cannot be negative") {
		t.Errorf("Error = %q, want to contain 'cannot be negative'", err.Error())
	}
}

func TestParseScenario_ViewportAndDeviceMutuallyExclusive(t *testing.T) {
	yaml := `
scenario: test
persona: sarah
goal: test
success_criteria:
  - ok
environment:
  url: https://example.com
  viewport:
    width: 1920
    height: 1080
  device: iPhone 12
`
	_, err := ParseScenario([]byte(yaml))
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if !strings.Contains(err.Error(), "cannot specify both") {
		t.Errorf("Error = %q, want to contain 'cannot specify both'", err.Error())
	}
}

func TestScenarioConfig_IsRetryable(t *testing.T) {
	s := &ScenarioConfig{
		Retry: &ScenarioRetry{
			OnErrors: []string{"browser_crash", "timeout"},
			NotOn:    []string{"test_failure"},
		},
	}

	tests := []struct {
		errorType string
		want      bool
	}{
		{"browser_crash", true},
		{"timeout", true},
		{"test_failure", false},
		{"network_error", false}, // Not in on_errors
	}

	for _, tt := range tests {
		t.Run(tt.errorType, func(t *testing.T) {
			got := s.IsRetryable(tt.errorType)
			if got != tt.want {
				t.Errorf("IsRetryable(%q) = %v, want %v", tt.errorType, got, tt.want)
			}
		})
	}
}

func TestScenarioConfig_ShouldRunHeaded(t *testing.T) {
	tests := []struct {
		name      string
		recording *ScenarioRecording
		want      bool
	}{
		{
			name:      "nil recording",
			recording: nil,
			want:      false,
		},
		{
			name:      "headed false",
			recording: &ScenarioRecording{Headed: false},
			want:      false,
		},
		{
			name:      "headed true",
			recording: &ScenarioRecording{Headed: true},
			want:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &ScenarioConfig{Recording: tt.recording}
			got := s.ShouldRunHeaded()
			if got != tt.want {
				t.Errorf("ShouldRunHeaded() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestYAMLDuration_UnmarshalYAML(t *testing.T) {
	tests := []struct {
		input string
		want  time.Duration
	}{
		{"5m", 5 * time.Minute},
		{"10s", 10 * time.Second},
		{"1h30m", 90 * time.Minute},
		{"500ms", 500 * time.Millisecond},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			yaml := `
scenario: test
persona: sarah
goal: test
success_criteria:
  - ok
environment:
  url: https://example.com
timeout: ` + tt.input

			s, err := ParseScenario([]byte(yaml))
			if err != nil {
				t.Fatalf("ParseScenario failed: %v", err)
			}
			if s.Timeout.Duration() != tt.want {
				t.Errorf("Timeout = %v, want %v", s.Timeout.Duration(), tt.want)
			}
		})
	}
}

func TestGenerateRunID(t *testing.T) {
	id1 := GenerateRunID()
	id2 := GenerateRunID()

	if !strings.HasPrefix(id1, "run-") {
		t.Errorf("RunID should start with 'run-': %s", id1)
	}
	if id1 == id2 {
		t.Error("Generated RunIDs should be unique")
	}
}

func TestParseScenario_DeviceWithoutViewport(t *testing.T) {
	yaml := `
scenario: mobile_test
persona: sarah
goal: Test on mobile device
success_criteria:
  - Works on mobile
environment:
  url: https://example.com
  device: iPhone 12
`
	s, err := ParseScenario([]byte(yaml))
	if err != nil {
		t.Fatalf("ParseScenario failed: %v", err)
	}

	if s.Environment.Device != "iPhone 12" {
		t.Errorf("Device = %q, want %q", s.Environment.Device, "iPhone 12")
	}
	// Viewport should not be set when device is specified
	if s.Environment.Viewport != nil {
		t.Error("Viewport should be nil when device is specified")
	}
}

func TestParseScenario_TestDataCleanupDefaults(t *testing.T) {
	yaml := `
scenario: test
persona: sarah
goal: test
success_criteria:
  - ok
environment:
  url: https://example.com
test_data:
  email_pattern: "test@example.com"
`
	s, err := ParseScenario([]byte(yaml))
	if err != nil {
		t.Fatalf("ParseScenario failed: %v", err)
	}

	// Should have default cleanup strategy
	if s.TestData.CleanupStrategy == nil {
		t.Fatal("Expected cleanup_strategy to be set")
	}
	if s.TestData.CleanupStrategy.OnSuccess != "delete_account" {
		t.Errorf("OnSuccess = %q, want delete_account", s.TestData.CleanupStrategy.OnSuccess)
	}
	if s.TestData.CleanupStrategy.OnFailure != "mark_for_review" {
		t.Errorf("OnFailure = %q, want mark_for_review", s.TestData.CleanupStrategy.OnFailure)
	}
	if s.TestData.CleanupStrategy.OnCrash != "cleanup_job" {
		t.Errorf("OnCrash = %q, want cleanup_job", s.TestData.CleanupStrategy.OnCrash)
	}
}
