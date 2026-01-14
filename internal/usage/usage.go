// Package usage provides Claude usage tracking and auto-switching logic.
// When Claude usage exceeds a configured threshold, agents can auto-switch to Gemini.
package usage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// ClaudeStatsCache represents the stats-cache.json file from Claude Code.
type ClaudeStatsCache struct {
	Version          int          `json:"version"`
	LastComputedDate string       `json:"lastComputedDate"`
	DailyActivity    []DailyStats `json:"dailyActivity"`
	DailyModelTokens []DailyTokens `json:"dailyModelTokens"`
	ModelUsage       map[string]ModelUsage `json:"modelUsage"`
	TotalSessions    int          `json:"totalSessions"`
	TotalMessages    int          `json:"totalMessages"`
}

// DailyStats represents daily activity counts.
type DailyStats struct {
	Date          string `json:"date"`
	MessageCount  int    `json:"messageCount"`
	SessionCount  int    `json:"sessionCount"`
	ToolCallCount int    `json:"toolCallCount"`
}

// DailyTokens represents daily token usage per model.
type DailyTokens struct {
	Date          string         `json:"date"`
	TokensByModel map[string]int `json:"tokensByModel"`
}

// ModelUsage represents token usage for a specific model.
type ModelUsage struct {
	InputTokens            int64   `json:"inputTokens"`
	OutputTokens           int64   `json:"outputTokens"`
	CacheReadInputTokens   int64   `json:"cacheReadInputTokens"`
	CacheCreationInputTokens int64 `json:"cacheCreationInputTokens"`
	WebSearchRequests      int     `json:"webSearchRequests"`
	CostUSD                float64 `json:"costUSD"`
}

// UsageResult contains the result of checking Claude usage.
type UsageResult struct {
	// WeeklyTokens is the total tokens used this week (input + output)
	WeeklyTokens int64

	// WeeklyLimit is the configured weekly token limit
	WeeklyLimit int64

	// UsagePercent is the percentage of the weekly limit used (0-100)
	UsagePercent float64

	// ExceedsThreshold is true if usage exceeds the configured threshold
	ExceedsThreshold bool

	// Threshold is the configured threshold percentage
	Threshold float64

	// Error is set if usage couldn't be determined
	Error error
}

// Config contains usage checking configuration.
type Config struct {
	// Threshold is the percentage (0-100) at which to trigger auto-switch.
	// Default: 80
	Threshold float64

	// WeeklyLimit is the weekly token limit.
	// Default: 30_000_000 (30M tokens, typical Pro limit)
	WeeklyLimit int64

	// Enabled controls whether usage checking is active.
	// Default: true
	Enabled bool

	// FallbackAgent is the agent to switch to when threshold is exceeded.
	// Default: "gemini"
	FallbackAgent string
}

// DefaultConfig returns the default usage configuration.
func DefaultConfig() *Config {
	return &Config{
		Threshold:     80,
		WeeklyLimit:   30_000_000, // 30M tokens (typical Claude Pro weekly limit)
		Enabled:       true,
		FallbackAgent: "gemini",
	}
}

// CheckUsage checks Claude's current weekly usage against configured limits.
func CheckUsage(cfg *Config) *UsageResult {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	result := &UsageResult{
		WeeklyLimit: cfg.WeeklyLimit,
		Threshold:   cfg.Threshold,
	}

	// Find Claude config directory
	configDir := getClaudeConfigDir()
	statsPath := filepath.Join(configDir, "stats-cache.json")

	// Read stats cache
	data, err := os.ReadFile(statsPath)
	if err != nil {
		result.Error = fmt.Errorf("reading Claude stats: %w", err)
		return result
	}

	var stats ClaudeStatsCache
	if err := json.Unmarshal(data, &stats); err != nil {
		result.Error = fmt.Errorf("parsing Claude stats: %w", err)
		return result
	}

	// Calculate this week's tokens
	weekStart := getWeekStart(time.Now())
	weeklyTokens := calculateWeeklyTokens(stats, weekStart)

	result.WeeklyTokens = weeklyTokens
	result.UsagePercent = float64(weeklyTokens) / float64(cfg.WeeklyLimit) * 100
	result.ExceedsThreshold = result.UsagePercent >= cfg.Threshold

	return result
}

// ShouldAutoSwitch returns true if the agent should auto-switch due to high usage,
// along with the recommended fallback agent.
func ShouldAutoSwitch(cfg *Config) (bool, string, *UsageResult) {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	if !cfg.Enabled {
		return false, "", nil
	}

	result := CheckUsage(cfg)

	// If there's an error reading usage, don't auto-switch (fail open)
	if result.Error != nil {
		return false, "", result
	}

	if result.ExceedsThreshold {
		return true, cfg.FallbackAgent, result
	}

	return false, "", result
}

// getClaudeConfigDir returns the Claude config directory.
// Uses CLAUDE_CONFIG_DIR env var if set, otherwise ~/.claude.
func getClaudeConfigDir() string {
	if dir := os.Getenv("CLAUDE_CONFIG_DIR"); dir != "" {
		return dir
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	return filepath.Join(home, ".claude")
}

// getWeekStart returns the start of the week (Sunday) for the given time.
func getWeekStart(t time.Time) time.Time {
	// Reset to start of day
	year, month, day := t.Date()
	t = time.Date(year, month, day, 0, 0, 0, 0, t.Location())

	// Go back to Sunday
	weekday := int(t.Weekday())
	return t.AddDate(0, 0, -weekday)
}

// calculateWeeklyTokens calculates total tokens used since weekStart.
func calculateWeeklyTokens(stats ClaudeStatsCache, weekStart time.Time) int64 {
	var total int64

	// Sum up tokens from daily token counts
	for _, daily := range stats.DailyModelTokens {
		date, err := time.Parse("2006-01-02", daily.Date)
		if err != nil {
			continue
		}

		// Only count tokens from this week
		if !date.Before(weekStart) {
			for _, tokens := range daily.TokensByModel {
				total += int64(tokens)
			}
		}
	}

	// If no daily data, fall back to model usage totals
	// (this is less accurate as it's cumulative, not weekly)
	if total == 0 && len(stats.ModelUsage) > 0 {
		for _, usage := range stats.ModelUsage {
			total += usage.InputTokens + usage.OutputTokens
		}
	}

	return total
}

// FormatUsage returns a human-readable usage string.
func FormatUsage(result *UsageResult) string {
	if result.Error != nil {
		return fmt.Sprintf("Usage unknown: %v", result.Error)
	}

	return fmt.Sprintf("%.1f%% of weekly limit (%.1fM / %.1fM tokens)",
		result.UsagePercent,
		float64(result.WeeklyTokens)/1_000_000,
		float64(result.WeeklyLimit)/1_000_000)
}
