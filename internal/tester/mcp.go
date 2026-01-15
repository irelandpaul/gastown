package tester

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// MCPConfig represents the structure of .mcp.json for Claude Code.
type MCPConfig struct {
	// MCPServers maps server names to their configurations.
	MCPServers map[string]MCPServerConfig `json:"mcpServers"`
}

// DefaultPlaywrightMCPConfig returns the default Playwright MCP server configuration.
// This uses the official Microsoft Playwright MCP server.
func DefaultPlaywrightMCPConfig() MCPServerConfig {
	return MCPServerConfig{
		Command: "npx",
		Args:    []string{"@playwright/mcp@latest"},
		Env:     map[string]string{},
	}
}

// PlaywrightMCPConfigWithRecording returns a Playwright MCP config with recording enabled.
// The outputDir specifies where to save video and trace files.
func PlaywrightMCPConfigWithRecording(outputDir string, cfg *PlaywrightConfig) MCPServerConfig {
	config := DefaultPlaywrightMCPConfig()

	// Add environment variables for recording configuration
	config.Env = make(map[string]string)

	if outputDir != "" {
		config.Env["PLAYWRIGHT_VIDEO_DIR"] = filepath.Join(outputDir, "video")
		config.Env["PLAYWRIGHT_TRACES_DIR"] = filepath.Join(outputDir, "trace")
	}

	if cfg != nil {
		if cfg.Headless {
			config.Env["PLAYWRIGHT_HEADLESS"] = "true"
		} else if cfg.Headed {
			config.Env["PLAYWRIGHT_HEADLESS"] = "false"
		}

		if cfg.Browser != "" {
			config.Env["PLAYWRIGHT_BROWSER"] = cfg.Browser
		}

		if cfg.SlowMo > 0 {
			config.Env["PLAYWRIGHT_SLOW_MO"] = fmt.Sprintf("%d", cfg.SlowMo)
		}

		if cfg.Timeout > 0 {
			config.Env["PLAYWRIGHT_TIMEOUT"] = fmt.Sprintf("%d", cfg.Timeout)
		}
	}

	return config
}

// GenerateMCPConfigFile creates a .mcp.json file for the tester agent.
// This configures the Playwright MCP server for browser automation.
func GenerateMCPConfigFile(outputDir string, playwrightCfg *PlaywrightConfig) (*MCPConfig, error) {
	config := &MCPConfig{
		MCPServers: map[string]MCPServerConfig{
			"playwright": PlaywrightMCPConfigWithRecording(outputDir, playwrightCfg),
		},
	}

	return config, nil
}

// WriteMCPConfig writes the MCP configuration to a .mcp.json file.
func WriteMCPConfig(dir string, config *MCPConfig) error {
	mcpPath := filepath.Join(dir, ".mcp.json")

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling MCP config: %w", err)
	}

	if err := os.WriteFile(mcpPath, data, 0644); err != nil {
		return fmt.Errorf("writing MCP config: %w", err)
	}

	return nil
}

// EnsureMCPConfig ensures the MCP configuration exists and is up to date.
// Returns the path to the .mcp.json file.
func EnsureMCPConfig(dir string, outputDir string, playwrightCfg *PlaywrightConfig) (string, error) {
	config, err := GenerateMCPConfigFile(outputDir, playwrightCfg)
	if err != nil {
		return "", err
	}

	if err := WriteMCPConfig(dir, config); err != nil {
		return "", err
	}

	return filepath.Join(dir, ".mcp.json"), nil
}

// PlaywrightTools lists the tools available from Playwright MCP.
// These match the tools documented in the Playwright MCP server.
var PlaywrightTools = []string{
	// Navigation
	"browser_navigate",
	"browser_navigate_back",
	"browser_navigate_forward",
	"browser_reload",

	// Actions
	"browser_click",
	"browser_type",
	"browser_fill",
	"browser_press_key",
	"browser_select_option",
	"browser_check",
	"browser_uncheck",
	"browser_hover",
	"browser_drag_and_drop",

	// Observation
	"browser_screenshot",
	"browser_get_text",
	"browser_get_html",
	"browser_get_attribute",
	"browser_get_url",
	"browser_get_title",

	// Waiting
	"browser_wait_for_selector",
	"browser_wait_for_navigation",
	"browser_wait_for_timeout",
	"browser_wait_for_load_state",

	// Frame handling
	"browser_switch_to_frame",
	"browser_switch_to_main",

	// Dialog handling
	"browser_handle_dialog",
}

// ObservationType defines valid observation types.
type ObservationType string

const (
	// ObservationConfusion indicates user would be confused.
	ObservationConfusion ObservationType = "confusion"

	// ObservationFriction indicates unnecessary steps or clicks.
	ObservationFriction ObservationType = "friction"

	// ObservationError indicates something went wrong.
	ObservationError ObservationType = "error"

	// ObservationSuccess indicates a goal was completed.
	ObservationSuccess ObservationType = "success"

	// ObservationSuggestion indicates a UX improvement idea.
	ObservationSuggestion ObservationType = "suggestion"
)

// Severity defines observation priority levels.
type Severity string

const (
	// SeverityP0 is blocking - user cannot complete goal.
	SeverityP0 Severity = "P0"

	// SeverityP1 is significant friction - user likely to abandon.
	SeverityP1 Severity = "P1"

	// SeverityP2 is minor friction - noticeable but not blocking.
	SeverityP2 Severity = "P2"

	// SeverityP3 is a nitpick - improvement opportunity.
	SeverityP3 Severity = "P3"
)

// Confidence defines agent self-assessment levels.
type Confidence string

const (
	// ConfidenceHigh means the agent is certain about the observation.
	ConfidenceHigh Confidence = "high"

	// ConfidenceMedium means the agent is fairly confident.
	ConfidenceMedium Confidence = "medium"

	// ConfidenceLow means the agent is uncertain.
	ConfidenceLow Confidence = "low"
)

// ValidateSeverity checks if a severity value is valid.
func ValidateSeverity(s string) bool {
	switch Severity(s) {
	case SeverityP0, SeverityP1, SeverityP2, SeverityP3:
		return true
	default:
		return false
	}
}

// ValidateConfidence checks if a confidence value is valid.
func ValidateConfidence(c string) bool {
	switch Confidence(c) {
	case ConfidenceHigh, ConfidenceMedium, ConfidenceLow:
		return true
	default:
		return false
	}
}

// ValidateObservationType checks if an observation type is valid.
func ValidateObservationType(t string) bool {
	switch ObservationType(t) {
	case ObservationConfusion, ObservationFriction, ObservationError, ObservationSuccess, ObservationSuggestion:
		return true
	default:
		return false
	}
}
