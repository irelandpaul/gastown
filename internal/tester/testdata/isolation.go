package testdata

import (
	"crypto/rand"
	"fmt"
	"strings"
	"time"
)

// GenerateRunID creates a unique run identifier.
// Format: 8-character hex string (e.g., "a1b2c3d4").
func GenerateRunID() string {
	b := make([]byte, 4)
	if _, err := rand.Read(b); err != nil {
		// Fallback to timestamp-based ID if crypto/rand fails
		return fmt.Sprintf("%08x", time.Now().UnixNano()&0xFFFFFFFF)
	}
	return fmt.Sprintf("%02x%02x%02x%02x", b[0], b[1], b[2], b[3])
}

// GenerateUniqueSuffix creates a UUID-like suffix for data isolation.
// Format: 12-character hex string (e.g., "a1b2c3d4e5f6").
func GenerateUniqueSuffix() string {
	b := make([]byte, 6)
	if _, err := rand.Read(b); err != nil {
		// Fallback to timestamp-based suffix
		return fmt.Sprintf("%012x", time.Now().UnixNano()&0xFFFFFFFFFFFF)
	}
	return fmt.Sprintf("%02x%02x%02x%02x%02x%02x", b[0], b[1], b[2], b[3], b[4], b[5])
}

// GenerateEmail generates a unique email address from a pattern.
//
// Supported placeholders:
//   - {scenario}: The scenario name
//   - {run_id}: The unique run identifier
//   - {timestamp}: Unix timestamp
//   - {suffix}: A unique suffix (if isolation enabled)
//
// Example:
//
//	pattern: "test+{scenario}+{run_id}@screencoach.test"
//	scenario: "sarah_registers"
//	runID: "a1b2c3d4"
//	result: "test+sarah_registers+a1b2c3d4@screencoach.test"
func GenerateEmail(pattern, scenario, runID string, suffix string) string {
	email := pattern

	// Replace placeholders
	email = strings.ReplaceAll(email, "{scenario}", scenario)
	email = strings.ReplaceAll(email, "{run_id}", runID)
	email = strings.ReplaceAll(email, "{timestamp}", fmt.Sprintf("%d", time.Now().Unix()))

	if suffix != "" {
		email = strings.ReplaceAll(email, "{suffix}", suffix)
	} else {
		// Remove {suffix} placeholder if no suffix provided
		email = strings.ReplaceAll(email, "+{suffix}", "")
		email = strings.ReplaceAll(email, "{suffix}", "")
	}

	return email
}

// NewRunContext creates a new run context with generated identifiers.
func NewRunContext(scenario string, config Config) *RunContext {
	runID := GenerateRunID()

	var suffix string
	if config.Isolation.UniqueSuffix {
		suffix = GenerateUniqueSuffix()
	}

	email := GenerateEmail(config.EmailPattern, scenario, runID, suffix)

	return &RunContext{
		RunID:        runID,
		Scenario:     scenario,
		StartedAt:    time.Now(),
		Email:        email,
		UniqueSuffix: suffix,
		Config:       config,
	}
}

// UniqueUsername generates a unique username from a base name.
// Appends the run ID to ensure uniqueness.
func (ctx *RunContext) UniqueUsername(baseName string) string {
	if ctx.UniqueSuffix != "" {
		return fmt.Sprintf("%s_%s", baseName, ctx.UniqueSuffix[:8])
	}
	return fmt.Sprintf("%s_%s", baseName, ctx.RunID)
}

// UniqueChildName generates a unique child profile name.
func (ctx *RunContext) UniqueChildName(baseName string) string {
	if ctx.UniqueSuffix != "" {
		return fmt.Sprintf("%s_%s", baseName, ctx.UniqueSuffix[:6])
	}
	return fmt.Sprintf("%s_%s", baseName, ctx.RunID[:6])
}

// UniqueData wraps any data value with the unique suffix for isolation.
// This can be used for any string data that needs to be unique per run.
func (ctx *RunContext) UniqueData(key, value string) string {
	if ctx.UniqueSuffix != "" {
		return fmt.Sprintf("%s_%s", value, ctx.UniqueSuffix[:8])
	}
	return fmt.Sprintf("%s_%s", value, ctx.RunID)
}
