package usage

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Threshold != 80 {
		t.Errorf("expected threshold 80, got %f", cfg.Threshold)
	}

	if cfg.WeeklyLimit != 30_000_000 {
		t.Errorf("expected weekly limit 30M, got %d", cfg.WeeklyLimit)
	}

	if !cfg.Enabled {
		t.Error("expected enabled by default")
	}

	if cfg.FallbackAgent != "gemini" {
		t.Errorf("expected fallback agent 'gemini', got '%s'", cfg.FallbackAgent)
	}
}

func TestGetWeekStart(t *testing.T) {
	tests := []struct {
		input    time.Time
		expected time.Time
	}{
		{
			// Sunday Jan 11, 2026 - should return same day at midnight
			input:    time.Date(2026, 1, 11, 10, 30, 0, 0, time.UTC),
			expected: time.Date(2026, 1, 11, 0, 0, 0, 0, time.UTC),
		},
		{
			// Monday Jan 12, 2026 - should return previous Sunday (Jan 11)
			input:    time.Date(2026, 1, 12, 10, 30, 0, 0, time.UTC),
			expected: time.Date(2026, 1, 11, 0, 0, 0, 0, time.UTC),
		},
		{
			// Saturday Jan 17, 2026 - should return previous Sunday (Jan 11)
			input:    time.Date(2026, 1, 17, 23, 59, 59, 0, time.UTC),
			expected: time.Date(2026, 1, 11, 0, 0, 0, 0, time.UTC),
		},
	}

	for _, tc := range tests {
		result := getWeekStart(tc.input)
		if !result.Equal(tc.expected) {
			t.Errorf("getWeekStart(%v) = %v, expected %v", tc.input, result, tc.expected)
		}
	}
}

func TestCalculateWeeklyTokens(t *testing.T) {
	stats := ClaudeStatsCache{
		DailyModelTokens: []DailyTokens{
			{Date: "2026-01-12", TokensByModel: map[string]int{"claude": 100000}},
			{Date: "2026-01-13", TokensByModel: map[string]int{"claude": 200000}},
			{Date: "2026-01-14", TokensByModel: map[string]int{"claude": 300000}},
			// Older dates not in this week
			{Date: "2026-01-05", TokensByModel: map[string]int{"claude": 500000}},
		},
	}

	// Week starting 2026-01-12 (Sunday)
	weekStart := time.Date(2026, 1, 12, 0, 0, 0, 0, time.UTC)
	tokens := calculateWeeklyTokens(stats, weekStart)

	expected := int64(100000 + 200000 + 300000)
	if tokens != expected {
		t.Errorf("expected %d tokens, got %d", expected, tokens)
	}
}

func TestCheckUsage(t *testing.T) {
	// Create a temporary stats cache file
	tmpDir := t.TempDir()
	statsPath := filepath.Join(tmpDir, "stats-cache.json")

	stats := ClaudeStatsCache{
		Version:          1,
		LastComputedDate: time.Now().Format("2006-01-02"),
		DailyModelTokens: []DailyTokens{
			{Date: time.Now().Format("2006-01-02"), TokensByModel: map[string]int{"claude": 25000000}},
		},
	}

	data, err := json.Marshal(stats)
	if err != nil {
		t.Fatalf("failed to marshal stats: %v", err)
	}

	if err := os.WriteFile(statsPath, data, 0644); err != nil {
		t.Fatalf("failed to write stats file: %v", err)
	}

	// Set CLAUDE_CONFIG_DIR to use our temp directory
	oldConfigDir := os.Getenv("CLAUDE_CONFIG_DIR")
	os.Setenv("CLAUDE_CONFIG_DIR", tmpDir)
	defer os.Setenv("CLAUDE_CONFIG_DIR", oldConfigDir)

	cfg := &Config{
		Threshold:     80,
		WeeklyLimit:   30_000_000,
		Enabled:       true,
		FallbackAgent: "gemini",
	}

	result := CheckUsage(cfg)

	if result.Error != nil {
		t.Fatalf("unexpected error: %v", result.Error)
	}

	if result.WeeklyTokens != 25_000_000 {
		t.Errorf("expected 25M tokens, got %d", result.WeeklyTokens)
	}

	// 25M / 30M = ~83.3%
	if result.UsagePercent < 83 || result.UsagePercent > 84 {
		t.Errorf("expected usage percent ~83.3%%, got %.1f%%", result.UsagePercent)
	}

	if !result.ExceedsThreshold {
		t.Error("expected usage to exceed threshold")
	}
}

func TestShouldAutoSwitch(t *testing.T) {
	// Create a temporary stats cache file
	tmpDir := t.TempDir()
	statsPath := filepath.Join(tmpDir, "stats-cache.json")

	tests := []struct {
		name             string
		tokens           int
		threshold        float64
		enabled          bool
		expectSwitch     bool
		expectFallback   string
	}{
		{
			name:           "high usage triggers switch",
			tokens:         25_000_000,
			threshold:      80,
			enabled:        true,
			expectSwitch:   true,
			expectFallback: "gemini",
		},
		{
			name:           "low usage no switch",
			tokens:         10_000_000,
			threshold:      80,
			enabled:        true,
			expectSwitch:   false,
			expectFallback: "",
		},
		{
			name:           "disabled no switch",
			tokens:         25_000_000,
			threshold:      80,
			enabled:        false,
			expectSwitch:   false,
			expectFallback: "",
		},
		{
			name:           "exact threshold triggers switch",
			tokens:         24_000_000, // 80% of 30M
			threshold:      80,
			enabled:        true,
			expectSwitch:   true,
			expectFallback: "gemini",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			stats := ClaudeStatsCache{
				Version:          1,
				LastComputedDate: time.Now().Format("2006-01-02"),
				DailyModelTokens: []DailyTokens{
					{Date: time.Now().Format("2006-01-02"), TokensByModel: map[string]int{"claude": tc.tokens}},
				},
			}

			data, _ := json.Marshal(stats)
			os.WriteFile(statsPath, data, 0644)

			// Set CLAUDE_CONFIG_DIR to use our temp directory
			oldConfigDir := os.Getenv("CLAUDE_CONFIG_DIR")
			os.Setenv("CLAUDE_CONFIG_DIR", tmpDir)
			defer os.Setenv("CLAUDE_CONFIG_DIR", oldConfigDir)

			cfg := &Config{
				Threshold:     tc.threshold,
				WeeklyLimit:   30_000_000,
				Enabled:       tc.enabled,
				FallbackAgent: "gemini",
			}

			shouldSwitch, fallback, _ := ShouldAutoSwitch(cfg)

			if shouldSwitch != tc.expectSwitch {
				t.Errorf("expected shouldSwitch=%v, got %v", tc.expectSwitch, shouldSwitch)
			}

			if tc.expectSwitch && fallback != tc.expectFallback {
				t.Errorf("expected fallback='%s', got '%s'", tc.expectFallback, fallback)
			}
		})
	}
}

func TestFormatUsage(t *testing.T) {
	result := &UsageResult{
		WeeklyTokens: 25_000_000,
		WeeklyLimit:  30_000_000,
		UsagePercent: 83.33,
	}

	formatted := FormatUsage(result)

	if formatted == "" {
		t.Error("expected non-empty formatted string")
	}

	// Check it contains relevant info
	if !contains(formatted, "83.3%") {
		t.Errorf("expected formatted string to contain percentage, got: %s", formatted)
	}

	if !contains(formatted, "25.0M") {
		t.Errorf("expected formatted string to contain token count, got: %s", formatted)
	}
}

func TestFormatUsageWithError(t *testing.T) {
	result := &UsageResult{
		Error: os.ErrNotExist,
	}

	formatted := FormatUsage(result)

	if !contains(formatted, "unknown") {
		t.Errorf("expected 'unknown' in error case, got: %s", formatted)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
