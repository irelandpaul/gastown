package testdata

import (
	"strings"
	"testing"
)

func TestGenerateRunID(t *testing.T) {
	id := GenerateRunID()

	// Should be 8 hex characters
	if len(id) != 8 {
		t.Errorf("expected RunID length 8, got %d: %s", len(id), id)
	}

	// Should be unique
	id2 := GenerateRunID()
	if id == id2 {
		t.Errorf("expected unique RunIDs, got duplicates: %s", id)
	}
}

func TestGenerateUniqueSuffix(t *testing.T) {
	suffix := GenerateUniqueSuffix()

	// Should be 12 hex characters
	if len(suffix) != 12 {
		t.Errorf("expected suffix length 12, got %d: %s", len(suffix), suffix)
	}

	// Should be unique
	suffix2 := GenerateUniqueSuffix()
	if suffix == suffix2 {
		t.Errorf("expected unique suffixes, got duplicates: %s", suffix)
	}
}

func TestGenerateEmail(t *testing.T) {
	tests := []struct {
		name     string
		pattern  string
		scenario string
		runID    string
		suffix   string
		expected string
	}{
		{
			name:     "basic pattern",
			pattern:  "test+{scenario}+{run_id}@example.com",
			scenario: "login_flow",
			runID:    "abc123",
			suffix:   "",
			expected: "test+login_flow+abc123@example.com",
		},
		{
			name:     "with suffix",
			pattern:  "test+{scenario}+{run_id}+{suffix}@example.com",
			scenario: "register",
			runID:    "xyz789",
			suffix:   "unique123",
			expected: "test+register+xyz789+unique123@example.com",
		},
		{
			name:     "suffix placeholder without suffix",
			pattern:  "test+{scenario}+{suffix}@example.com",
			scenario: "checkout",
			runID:    "def456",
			suffix:   "",
			expected: "test+checkout@example.com",
		},
		{
			name:     "simple pattern",
			pattern:  "{scenario}@test.local",
			scenario: "simple_test",
			runID:    "123",
			suffix:   "",
			expected: "simple_test@test.local",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateEmail(tt.pattern, tt.scenario, tt.runID, tt.suffix)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestNewRunContext(t *testing.T) {
	config := DefaultConfig()
	ctx := NewRunContext("sarah_registers", config)

	// RunID should be set
	if ctx.RunID == "" {
		t.Error("expected RunID to be set")
	}

	// Scenario should match
	if ctx.Scenario != "sarah_registers" {
		t.Errorf("expected scenario 'sarah_registers', got %q", ctx.Scenario)
	}

	// Email should contain scenario and run ID
	if !strings.Contains(ctx.Email, "sarah_registers") {
		t.Errorf("expected email to contain scenario, got %q", ctx.Email)
	}
	if !strings.Contains(ctx.Email, ctx.RunID) {
		t.Errorf("expected email to contain runID, got %q", ctx.Email)
	}

	// With default config, unique suffix should be set
	if ctx.UniqueSuffix == "" {
		t.Error("expected UniqueSuffix to be set with default config")
	}
}

func TestRunContext_UniqueUsername(t *testing.T) {
	config := DefaultConfig()
	ctx := NewRunContext("test_scenario", config)

	username := ctx.UniqueUsername("sarah")

	// Should contain base name
	if !strings.HasPrefix(username, "sarah_") {
		t.Errorf("expected username to start with 'sarah_', got %q", username)
	}

	// Should be unique per context
	ctx2 := NewRunContext("test_scenario", config)
	username2 := ctx2.UniqueUsername("sarah")

	if username == username2 {
		t.Errorf("expected unique usernames, got duplicates: %s", username)
	}
}

func TestRunContext_UniqueChildName(t *testing.T) {
	config := DefaultConfig()
	ctx := NewRunContext("test_scenario", config)

	childName := ctx.UniqueChildName("Tommy")

	// Should contain base name
	if !strings.HasPrefix(childName, "Tommy_") {
		t.Errorf("expected child name to start with 'Tommy_', got %q", childName)
	}
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	// Check defaults match spec
	if config.EmailPattern != "test+{scenario}+{run_id}@screencoach.test" {
		t.Errorf("unexpected email pattern: %s", config.EmailPattern)
	}

	if config.EmailInbox != EmailSkipVerification {
		t.Errorf("expected EmailSkipVerification, got %s", config.EmailInbox)
	}

	if config.Cleanup.OnSuccess != CleanupDeleteAccount {
		t.Errorf("expected OnSuccess=delete_account, got %s", config.Cleanup.OnSuccess)
	}

	if config.Cleanup.OnFailure != CleanupMarkForReview {
		t.Errorf("expected OnFailure=mark_for_review, got %s", config.Cleanup.OnFailure)
	}

	if config.Cleanup.OnCrash != CleanupJob {
		t.Errorf("expected OnCrash=cleanup_job, got %s", config.Cleanup.OnCrash)
	}

	if !config.Isolation.UniqueSuffix {
		t.Error("expected Isolation.UniqueSuffix to be true")
	}
}

func TestNoIsolationSuffix(t *testing.T) {
	config := DefaultConfig()
	config.Isolation.UniqueSuffix = false

	ctx := NewRunContext("no_suffix_test", config)

	// UniqueSuffix should be empty when disabled
	if ctx.UniqueSuffix != "" {
		t.Errorf("expected empty UniqueSuffix, got %q", ctx.UniqueSuffix)
	}

	// Username should use RunID instead
	username := ctx.UniqueUsername("user")
	if !strings.Contains(username, ctx.RunID) {
		t.Errorf("expected username to contain RunID when suffix disabled: %s", username)
	}
}
