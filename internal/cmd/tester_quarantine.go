package cmd

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/steveyegge/gastown/internal/tester/flake"
)

var (
	quarantineReason     string
	quarantineOutputDir  string
	quarantineShowAll    bool
	quarantineClearHist  bool
)

var testerQuarantineCmd = &cobra.Command{
	Use:   "quarantine",
	Short: "Manage test quarantine",
	Long: `Manage quarantined tests and view flake metrics.

Quarantined tests are automatically skipped during batch runs unless
--include-quarantined is specified. Tests can be quarantined manually
or automatically based on flake detection.

SUBCOMMANDS:
  list     List all quarantined tests
  add      Add a test to quarantine
  remove   Remove a test from quarantine
  status   Show flake metrics for a test
  flaky    List all flaky tests (not yet quarantined)

Examples:
  gt tester quarantine list
  gt tester quarantine add registration-flow --reason "Flaky login button"
  gt tester quarantine remove registration-flow
  gt tester quarantine status registration-flow
  gt tester quarantine flaky`,
	RunE: requireSubcommand,
}

var quarantineListCmd = &cobra.Command{
	Use:   "list",
	Short: "List quarantined tests",
	Long: `List all tests currently in quarantine.

Shows the scenario name, quarantine date, reason, and flake rate.`,
	Args: cobra.NoArgs,
	RunE: runQuarantineList,
}

var quarantineAddCmd = &cobra.Command{
	Use:   "add <scenario>",
	Short: "Add a test to quarantine",
	Long: `Manually quarantine a test scenario.

Quarantined tests are skipped during batch runs. Use this to temporarily
disable flaky tests while investigating the root cause.

Examples:
  gt tester quarantine add registration-flow --reason "Investigating timeout"
  gt tester quarantine add checkout --reason "Depends on external service"`,
	Args: cobra.ExactArgs(1),
	RunE: runQuarantineAdd,
}

var quarantineRemoveCmd = &cobra.Command{
	Use:   "remove <scenario>",
	Short: "Remove a test from quarantine",
	Long: `Remove a test from quarantine, allowing it to run in batches again.

Use this after fixing the underlying issue causing the flakiness.

Examples:
  gt tester quarantine remove registration-flow`,
	Args: cobra.ExactArgs(1),
	RunE: runQuarantineRemove,
}

var quarantineStatusCmd = &cobra.Command{
	Use:   "status [scenario]",
	Short: "Show flake metrics for tests",
	Long: `Show flake metrics for a specific test or all tracked tests.

Displays:
- Flake rate (failure rate over the window)
- Success rate
- Consecutive failures/passes
- Run history summary
- Quarantine status

Examples:
  gt tester quarantine status                    # All tests
  gt tester quarantine status registration-flow  # Single test`,
	Args: cobra.MaximumNArgs(1),
	RunE: runQuarantineStatus,
}

var quarantineFlakyCmd = &cobra.Command{
	Use:   "flaky",
	Short: "List flaky tests",
	Long: `List all tests currently considered flaky but not yet quarantined.

These are candidates for quarantine or investigation. A test is considered
flaky when its failure rate exceeds the threshold (default 30%) over the
recent window (default 10 runs).

Examples:
  gt tester quarantine flaky`,
	Args: cobra.NoArgs,
	RunE: runQuarantineFlaky,
}

var quarantineClearCmd = &cobra.Command{
	Use:   "clear <scenario>",
	Short: "Clear history for a test",
	Long: `Clear the run history for a test scenario.

This resets all flake metrics and removes the scenario from tracking.
Use with caution - this loses all historical data for the test.

Examples:
  gt tester quarantine clear registration-flow`,
	Args: cobra.ExactArgs(1),
	RunE: runQuarantineClear,
}

func init() {
	// Quarantine add flags
	quarantineAddCmd.Flags().StringVarP(&quarantineReason, "reason", "r", "", "Reason for quarantining (required)")
	quarantineAddCmd.MarkFlagRequired("reason")

	// Quarantine status flags
	quarantineStatusCmd.Flags().BoolVar(&quarantineShowAll, "all", false, "Show all tracked scenarios (including stable)")

	// Global flags
	testerQuarantineCmd.PersistentFlags().StringVar(&quarantineOutputDir, "output", "test-results", "Output directory for flake data")

	// Add subcommands
	testerQuarantineCmd.AddCommand(quarantineListCmd)
	testerQuarantineCmd.AddCommand(quarantineAddCmd)
	testerQuarantineCmd.AddCommand(quarantineRemoveCmd)
	testerQuarantineCmd.AddCommand(quarantineStatusCmd)
	testerQuarantineCmd.AddCommand(quarantineFlakyCmd)
	testerQuarantineCmd.AddCommand(quarantineClearCmd)

	testerCmd.AddCommand(testerQuarantineCmd)
}

func getDetector() (*flake.Detector, error) {
	storagePath := filepath.Join(quarantineOutputDir, ".flake-data.json")
	return flake.NewDetector(storagePath, flake.DefaultConfig())
}

func runQuarantineList(cmd *cobra.Command, args []string) error {
	detector, err := getDetector()
	if err != nil {
		return fmt.Errorf("failed to initialize flake detector: %w", err)
	}

	entries := detector.ListQuarantined()

	if testerJSON {
		data, _ := json.MarshalIndent(entries, "", "  ")
		fmt.Println(string(data))
		return nil
	}

	if len(entries) == 0 {
		fmt.Println("No quarantined tests")
		return nil
	}

	fmt.Printf("Quarantined Tests (%d)\n", len(entries))
	fmt.Println(strings.Repeat("─", 60))

	for _, entry := range entries {
		autoTag := ""
		if entry.AutoQuarantined {
			autoTag = " [auto]"
		}

		reviewTag := ""
		if entry.ReviewRequired {
			reviewTag = " (needs review)"
		}

		fmt.Printf("  %s%s%s\n", entry.Scenario, autoTag, reviewTag)
		fmt.Printf("    Quarantined: %s\n", entry.QuarantinedAt.Format("2006-01-02 15:04"))
		fmt.Printf("    Reason: %s\n", entry.Reason)
		if entry.FlakeRate > 0 {
			fmt.Printf("    Flake rate: %.0f%%\n", entry.FlakeRate*100)
		}
		if entry.Notes != "" {
			fmt.Printf("    Notes: %s\n", entry.Notes)
		}
		fmt.Println()
	}

	return nil
}

func runQuarantineAdd(cmd *cobra.Command, args []string) error {
	scenario := args[0]

	detector, err := getDetector()
	if err != nil {
		return fmt.Errorf("failed to initialize flake detector: %w", err)
	}

	if detector.IsQuarantined(scenario) {
		return fmt.Errorf("scenario %q is already quarantined", scenario)
	}

	if err := detector.Quarantine(scenario, quarantineReason); err != nil {
		return fmt.Errorf("failed to quarantine scenario: %w", err)
	}

	fmt.Printf("Quarantined: %s\n", scenario)
	fmt.Printf("  Reason: %s\n", quarantineReason)
	fmt.Println("\nThis test will be skipped in batch runs. Use 'gt tester quarantine remove' to unquarantine.")

	return nil
}

func runQuarantineRemove(cmd *cobra.Command, args []string) error {
	scenario := args[0]

	detector, err := getDetector()
	if err != nil {
		return fmt.Errorf("failed to initialize flake detector: %w", err)
	}

	if !detector.IsQuarantined(scenario) {
		return fmt.Errorf("scenario %q is not quarantined", scenario)
	}

	if err := detector.Unquarantine(scenario); err != nil {
		return fmt.Errorf("failed to unquarantine scenario: %w", err)
	}

	fmt.Printf("Unquarantined: %s\n", scenario)
	fmt.Println("\nThis test will now run in batch executions.")

	return nil
}

func runQuarantineStatus(cmd *cobra.Command, args []string) error {
	detector, err := getDetector()
	if err != nil {
		return fmt.Errorf("failed to initialize flake detector: %w", err)
	}

	if len(args) == 1 {
		// Show status for specific scenario
		scenario := args[0]
		return showScenarioStatus(detector, scenario)
	}

	// Show status for all scenarios
	metrics := detector.GetAllMetrics()

	if testerJSON {
		data, _ := json.MarshalIndent(metrics, "", "  ")
		fmt.Println(string(data))
		return nil
	}

	if len(metrics) == 0 {
		fmt.Println("No test history tracked yet")
		return nil
	}

	// Filter to show only relevant scenarios unless --all
	var filtered []*flake.FlakeMetrics
	for _, m := range metrics {
		if quarantineShowAll || m.IsFlaky || detector.IsQuarantined(m.Scenario) {
			filtered = append(filtered, m)
		}
	}

	if len(filtered) == 0 {
		fmt.Println("All tracked tests are stable. Use --all to see all metrics.")
		return nil
	}

	fmt.Printf("Flake Status (%d scenarios)\n", len(filtered))
	fmt.Println(strings.Repeat("─", 60))

	for _, m := range filtered {
		printMetricsSummary(m, detector.IsQuarantined(m.Scenario))
	}

	return nil
}

func showScenarioStatus(detector *flake.Detector, scenario string) error {
	metrics := detector.GetMetrics(scenario)
	history := detector.GetHistory(scenario)
	entry := detector.GetQuarantineEntry(scenario)

	if testerJSON {
		data := map[string]interface{}{
			"metrics":    metrics,
			"quarantine": entry,
		}
		if history != nil {
			data["history"] = map[string]interface{}{
				"total_runs":    history.TotalRuns,
				"total_passes":  history.TotalPasses,
				"total_failures": history.TotalFailures,
				"first_run":     history.FirstRun,
				"last_run":      history.LastRun,
			}
		}
		output, _ := json.MarshalIndent(data, "", "  ")
		fmt.Println(string(output))
		return nil
	}

	fmt.Printf("Scenario: %s\n", scenario)
	fmt.Println(strings.Repeat("─", 40))

	if history == nil || history.TotalRuns == 0 {
		fmt.Println("No run history")
		return nil
	}

	// Quarantine status
	if entry != nil {
		fmt.Println("Status: QUARANTINED")
		fmt.Printf("  Quarantined: %s\n", entry.QuarantinedAt.Format("2006-01-02 15:04"))
		fmt.Printf("  Reason: %s\n", entry.Reason)
		if entry.AutoQuarantined {
			fmt.Println("  Type: Auto-quarantined")
		} else {
			fmt.Println("  Type: Manually quarantined")
		}
		if entry.ReviewRequired {
			fmt.Println("  Review: Required")
		}
		fmt.Println()
	} else if metrics.IsFlaky {
		fmt.Println("Status: FLAKY (not quarantined)")
		fmt.Println()
	} else {
		fmt.Println("Status: Stable")
		fmt.Println()
	}

	// Metrics
	fmt.Println("Window Metrics:")
	fmt.Printf("  Flake rate: %.0f%% (%d/%d failed)\n",
		metrics.FlakeRate*100, metrics.WindowFailures+metrics.WindowErrors, metrics.WindowRuns)
	fmt.Printf("  Success rate: %.0f%% (%d/%d passed)\n",
		metrics.SuccessRate*100, metrics.WindowPasses, metrics.WindowRuns)
	fmt.Printf("  Average retries: %.1f\n", metrics.AverageRetries)
	if metrics.AverageDuration > 0 {
		fmt.Printf("  Average duration: %s\n", formatDuration(metrics.AverageDuration))
	}
	fmt.Println()

	// Streak info
	fmt.Println("Current Streak:")
	if metrics.ConsecutiveFailures > 0 {
		fmt.Printf("  %d consecutive failures\n", metrics.ConsecutiveFailures)
	} else if metrics.ConsecutivePasses > 0 {
		fmt.Printf("  %d consecutive passes\n", metrics.ConsecutivePasses)
	}
	fmt.Printf("  Last outcome: %s\n", metrics.LastOutcome)
	fmt.Println()

	// History summary
	fmt.Println("All-Time Stats:")
	fmt.Printf("  Total runs: %d\n", history.TotalRuns)
	fmt.Printf("  Passes: %d (%.0f%%)\n", history.TotalPasses,
		float64(history.TotalPasses)/float64(history.TotalRuns)*100)
	fmt.Printf("  Failures: %d (%.0f%%)\n", history.TotalFailures,
		float64(history.TotalFailures)/float64(history.TotalRuns)*100)
	fmt.Printf("  Errors: %d (%.0f%%)\n", history.TotalErrors,
		float64(history.TotalErrors)/float64(history.TotalRuns)*100)
	fmt.Printf("  First run: %s\n", history.FirstRun.Format("2006-01-02 15:04"))
	fmt.Printf("  Last run: %s\n", history.LastRun.Format("2006-01-02 15:04"))
	fmt.Println()

	// Recent runs
	fmt.Println("Recent Runs:")
	maxShow := 5
	if len(history.Runs) < maxShow {
		maxShow = len(history.Runs)
	}
	for i := 0; i < maxShow; i++ {
		run := history.Runs[i]
		status := "✓"
		if run.Outcome == flake.OutcomeFail || run.Outcome == flake.OutcomeError {
			status = "✗"
		} else if run.Outcome == flake.OutcomeSkip {
			status = "○"
		}
		line := fmt.Sprintf("  %s %s %s",
			status,
			run.Timestamp.Format("01-02 15:04"),
			run.Outcome)
		if run.RetryCount > 0 {
			line += fmt.Sprintf(" (retry %d)", run.RetryCount)
		}
		if run.Duration > 0 {
			line += fmt.Sprintf(" [%s]", formatDuration(run.Duration))
		}
		fmt.Println(line)
	}

	return nil
}

func runQuarantineFlaky(cmd *cobra.Command, args []string) error {
	detector, err := getDetector()
	if err != nil {
		return fmt.Errorf("failed to initialize flake detector: %w", err)
	}

	flaky := detector.GetFlakyScenarios()

	// Filter out already quarantined
	var candidates []*flake.FlakeMetrics
	for _, m := range flaky {
		if !detector.IsQuarantined(m.Scenario) {
			candidates = append(candidates, m)
		}
	}

	if testerJSON {
		data, _ := json.MarshalIndent(candidates, "", "  ")
		fmt.Println(string(data))
		return nil
	}

	if len(candidates) == 0 {
		fmt.Println("No flaky tests detected (that aren't already quarantined)")
		return nil
	}

	fmt.Printf("Flaky Tests - Quarantine Candidates (%d)\n", len(candidates))
	fmt.Println(strings.Repeat("─", 60))
	fmt.Println("These tests have a failure rate above threshold but aren't quarantined yet.")
	fmt.Println()

	for _, m := range candidates {
		printMetricsSummary(m, false)
	}

	fmt.Println()
	fmt.Println("To quarantine a test:")
	fmt.Println("  gt tester quarantine add <scenario> --reason \"description\"")

	return nil
}

func runQuarantineClear(cmd *cobra.Command, args []string) error {
	scenario := args[0]

	detector, err := getDetector()
	if err != nil {
		return fmt.Errorf("failed to initialize flake detector: %w", err)
	}

	history := detector.GetHistory(scenario)
	if history == nil {
		return fmt.Errorf("no history found for scenario %q", scenario)
	}

	if err := detector.ClearHistory(scenario); err != nil {
		return fmt.Errorf("failed to clear history: %w", err)
	}

	fmt.Printf("Cleared history for: %s\n", scenario)
	fmt.Printf("  Deleted %d run records\n", history.TotalRuns)

	return nil
}

func printMetricsSummary(m *flake.FlakeMetrics, isQuarantined bool) {
	status := ""
	if isQuarantined {
		status = " [QUARANTINED]"
	} else if m.IsFlaky {
		status = " [FLAKY]"
	}

	fmt.Printf("  %s%s\n", m.Scenario, status)
	fmt.Printf("    Flake rate: %.0f%% | Success: %.0f%% | Runs: %d\n",
		m.FlakeRate*100, m.SuccessRate*100, m.WindowRuns)

	if m.ConsecutiveFailures > 0 {
		fmt.Printf("    Consecutive failures: %d\n", m.ConsecutiveFailures)
	}

	fmt.Println()
}
