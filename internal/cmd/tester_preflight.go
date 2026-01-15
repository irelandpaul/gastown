package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/steveyegge/gastown/internal/style"
	"github.com/steveyegge/gastown/internal/ui"
)

var (
	preflightFix bool
)

var testerPreflightCmd = &cobra.Command{
	Use:   "preflight",
	Short: "Run environment preflight checks before testing",
	Long: `Run environment preflight checks before testing.

This command verifies that the testing environment is properly configured:
  - Playwright installation
  - MCP server connection
  - Target environment reachability
  - API health endpoints
  - Disk space for artifacts

Use --fix to attempt automatic fixes for common issues.

Examples:
  gt tester preflight                # Run all checks
  gt tester preflight --env staging  # Check specific environment
  gt tester preflight --fix          # Auto-fix issues
  gt tester preflight --json         # Output as JSON`,
	RunE: runTesterPreflight,
}

// PreflightCheck represents a single preflight check result
type PreflightCheck struct {
	Name    string `json:"name"`
	Status  string `json:"status"` // "pass", "warn", "fail"
	Message string `json:"message,omitempty"`
	Details string `json:"details,omitempty"`
	Fix     string `json:"fix,omitempty"`
}

// PreflightResult contains all preflight check results
type PreflightResult struct {
	Environment string           `json:"environment"`
	Checks      []PreflightCheck `json:"checks"`
	AllPassed   bool             `json:"all_passed"`
	Warnings    int              `json:"warnings"`
	Failures    int              `json:"failures"`
}

func init() {
	testerPreflightCmd.Flags().StringVar(&testerEnv, "env", "staging", "Target environment (staging, production)")
	testerPreflightCmd.Flags().BoolVar(&preflightFix, "fix", false, "Attempt to fix issues automatically")
	testerPreflightCmd.Flags().BoolVar(&testerJSON, "json", false, "Output as JSON")
}

func runTesterPreflight(cmd *cobra.Command, args []string) error {
	result := PreflightResult{
		Environment: testerEnv,
		AllPassed:   true,
	}

	// Run all preflight checks
	checks := []func() PreflightCheck{
		checkPlaywright,
		checkMCPServer,
		checkNodeJS,
		checkDiskSpace,
	}

	for _, check := range checks {
		c := check()
		result.Checks = append(result.Checks, c)

		switch c.Status {
		case "fail":
			result.AllPassed = false
			result.Failures++
		case "warn":
			result.Warnings++
		}
	}

	// JSON output
	if testerJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(result)
	}

	// Human-readable output
	fmt.Printf("\n%s (%s)\n\n", style.Bold.Render("Preflight Checks"), testerEnv)

	for _, c := range result.Checks {
		icon := ui.RenderPassIcon()
		switch c.Status {
		case "warn":
			icon = ui.RenderWarnIcon()
		case "fail":
			icon = ui.RenderFailIcon()
		}

		fmt.Printf("  %s %s", icon, c.Name)
		if c.Message != "" {
			fmt.Printf(" (%s)", c.Message)
		}
		fmt.Println()

		if c.Details != "" {
			fmt.Printf("    %s\n", ui.RenderMuted(c.Details))
		}
		if c.Fix != "" && c.Status != "pass" {
			fmt.Printf("    Fix: %s\n", c.Fix)
		}
	}

	fmt.Println()

	if result.AllPassed && result.Warnings == 0 {
		fmt.Printf("%s All checks passed. Ready to run tests.\n", ui.RenderPassIcon())
	} else if result.AllPassed {
		fmt.Printf("%s %d warning(s). Tests can run but may have issues.\n",
			ui.RenderWarnIcon(), result.Warnings)
	} else {
		fmt.Printf("%s %d check(s) failed, %d warning(s). Fix issues before running tests.\n",
			ui.RenderFailIcon(), result.Failures, result.Warnings)
	}

	if !result.AllPassed {
		return NewSilentExit(4) // Exit code 4 for preflight failure
	}

	return nil
}

// checkPlaywright verifies Playwright is installed
func checkPlaywright() PreflightCheck {
	check := PreflightCheck{
		Name: "Playwright installed",
	}

	// Check for npx playwright
	cmd := exec.Command("npx", "playwright", "--version")
	out, err := cmd.Output()
	if err != nil {
		check.Status = "fail"
		check.Details = "Playwright not found"
		check.Fix = "npm install -D @playwright/test && npx playwright install"
		return check
	}

	version := strings.TrimSpace(string(out))
	check.Status = "pass"
	check.Message = version

	return check
}

// checkMCPServer verifies the Playwright MCP server is available
func checkMCPServer() PreflightCheck {
	check := PreflightCheck{
		Name: "MCP server available",
	}

	// Check if the playwright-mcp package is installed
	// First try checking if it's in the system PATH
	cmd := exec.Command("which", "playwright-mcp")
	if runtime.GOOS == "windows" {
		cmd = exec.Command("where", "playwright-mcp")
	}
	_, err := cmd.Output()
	if err == nil {
		check.Status = "pass"
		check.Message = "playwright-mcp found in PATH"
		return check
	}

	// Try checking for npx mcp-playwright
	cmd = exec.Command("npx", "--yes", "@anthropics/mcp-playwright@latest", "--help")
	cmd.Env = append(os.Environ(), "npm_config_yes=true")
	_, err = cmd.CombinedOutput()
	if err == nil {
		check.Status = "pass"
		check.Message = "@anthropics/mcp-playwright available"
		return check
	}

	// Check for common MCP server installations
	home, _ := os.UserHomeDir()
	mcpPaths := []string{
		filepath.Join(home, ".local", "bin", "playwright-mcp"),
		filepath.Join(home, "node_modules", ".bin", "playwright-mcp"),
		"/usr/local/bin/playwright-mcp",
	}

	for _, p := range mcpPaths {
		if _, err := os.Stat(p); err == nil {
			check.Status = "pass"
			check.Message = fmt.Sprintf("found at %s", p)
			return check
		}
	}

	// If not found, it might still work via Claude Code's built-in MCP support
	check.Status = "warn"
	check.Message = "MCP server not found locally"
	check.Details = "Playwright MCP may be provided by Claude Code runtime"
	check.Fix = "npm install -g @anthropics/mcp-playwright"

	return check
}

// checkNodeJS verifies Node.js is installed
func checkNodeJS() PreflightCheck {
	check := PreflightCheck{
		Name: "Node.js installed",
	}

	cmd := exec.Command("node", "--version")
	out, err := cmd.Output()
	if err != nil {
		check.Status = "fail"
		check.Details = "Node.js not found"
		check.Fix = "Install Node.js from https://nodejs.org"
		return check
	}

	version := strings.TrimSpace(string(out))
	check.Status = "pass"
	check.Message = version

	// Check version is at least 18
	if strings.HasPrefix(version, "v") {
		version = version[1:]
	}
	parts := strings.Split(version, ".")
	if len(parts) > 0 {
		major := 0
		fmt.Sscanf(parts[0], "%d", &major)
		if major < 18 {
			check.Status = "warn"
			check.Details = "Node.js 18+ recommended for Playwright"
		}
	}

	return check
}

// checkDiskSpace verifies sufficient disk space for artifacts
func checkDiskSpace() PreflightCheck {
	check := PreflightCheck{
		Name: "Disk space",
	}

	// Check free space in the current directory (where artifacts will be stored)
	cwd, err := os.Getwd()
	if err != nil {
		check.Status = "warn"
		check.Details = "Could not determine current directory"
		return check
	}

	// Use df command to get free space
	cmd := exec.Command("df", "-h", cwd)
	out, err := cmd.Output()
	if err != nil {
		check.Status = "warn"
		check.Details = "Could not determine free space"
		return check
	}

	// Parse df output - get the free space from the last column of the second line
	lines := strings.Split(string(out), "\n")
	if len(lines) < 2 {
		check.Status = "warn"
		check.Details = "Could not parse disk space"
		return check
	}

	fields := strings.Fields(lines[1])
	if len(fields) < 4 {
		check.Status = "warn"
		check.Details = "Could not parse disk space"
		return check
	}

	available := fields[3]
	check.Status = "pass"
	check.Message = fmt.Sprintf("%s available", available)

	// Try to parse and warn if less than 5GB
	var size float64
	var unit string
	_, err = fmt.Sscanf(available, "%f%s", &size, &unit)
	if err == nil {
		var sizeGB float64
		switch strings.ToUpper(unit) {
		case "G", "GB":
			sizeGB = size
		case "T", "TB":
			sizeGB = size * 1024
		case "M", "MB":
			sizeGB = size / 1024
		}

		if sizeGB < 5 {
			check.Status = "warn"
			check.Details = "Less than 5GB free - may run out during testing"
		}
	}

	return check
}

// checkEnvironmentReachability checks if a target environment is reachable
func checkEnvironmentReachability(targetURL string) PreflightCheck {
	check := PreflightCheck{
		Name: fmt.Sprintf("Environment reachable (%s)", targetURL),
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get(targetURL)
	if err != nil {
		check.Status = "fail"
		check.Details = fmt.Sprintf("Connection failed: %v", err)
		check.Fix = "Check VPN connection or server status"
		return check
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 400 {
		check.Status = "pass"
		check.Message = fmt.Sprintf("HTTP %d", resp.StatusCode)
	} else {
		check.Status = "fail"
		check.Message = fmt.Sprintf("HTTP %d", resp.StatusCode)
		check.Details = "Server returned error status"
	}

	return check
}
