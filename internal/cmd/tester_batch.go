package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/steveyegge/gastown/internal/tester/batch"
)

var (
	batchParallel           int
	batchStopOnFail         bool
	batchConvoy             string
	batchModel              string
	batchFilter             []string
	batchExclude            []string
	batchIncludeQuarantined bool
	batchCompareTo          string
	batchOutputDir          string
)

var testerBatchCmd = &cobra.Command{
	Use:   "batch <pattern>",
	Short: "Run multiple test scenarios",
	Long: `Run multiple test scenarios matching a glob pattern.

By default, quarantined tests are skipped. Use --include-quarantined to run them.

The batch runner:
1. Runs preflight checks (once for the batch)
2. Finds all matching scenario files
3. Filters out quarantined tests
4. Creates a convoy bead for tracking (if --convoy)
5. Runs scenarios (parallel if --parallel)
6. Aggregates results and prints summary

Examples:
  gt tester batch "scenarios/**/*.yaml"
  gt tester batch "scenarios/registration/*.yaml" --parallel 3
  gt tester batch "**/*.yaml" --filter critical-path
  gt tester batch "**/*.yaml" --exclude slow --stop-on-fail
  gt tester batch "**/*.yaml" --convoy parent-portal-tests`,
	Args: cobra.ExactArgs(1),
	RunE: runTesterBatch,
}

func init() {
	testerBatchCmd.Flags().IntVarP(&batchParallel, "parallel", "p", 1, "Number of scenarios to run simultaneously")
	testerBatchCmd.Flags().BoolVar(&batchStopOnFail, "stop-on-fail", false, "Stop batch on first failure")
	testerBatchCmd.Flags().StringVar(&batchConvoy, "convoy", "", "Create convoy bead with this name")
	testerBatchCmd.Flags().StringVar(&batchModel, "model", "", "Override model for all scenarios (haiku, sonnet, gemini)")
	testerBatchCmd.Flags().StringSliceVar(&batchFilter, "filter", nil, "Only run scenarios with these tags")
	testerBatchCmd.Flags().StringSliceVar(&batchExclude, "exclude", nil, "Skip scenarios with these tags")
	testerBatchCmd.Flags().BoolVar(&batchIncludeQuarantined, "include-quarantined", false, "Include quarantined tests")
	testerBatchCmd.Flags().StringVar(&batchCompareTo, "compare-to", "", "Compare to previous batch run")
	testerBatchCmd.Flags().BoolVar(&testerSkipPreflight, "skip-preflight", false, "Skip preflight checks (not recommended)")
	testerBatchCmd.Flags().StringVar(&batchOutputDir, "output", "test-results", "Output directory for results")

	testerCmd.AddCommand(testerBatchCmd)
}

func runTesterBatch(cmd *cobra.Command, args []string) error {
	pattern := args[0]

	config := batch.Config{
		Pattern:            pattern,
		Parallel:           batchParallel,
		StopOnFail:         batchStopOnFail,
		ConvoyName:         batchConvoy,
		Model:              batchModel,
		Environment:        testerEnv,
		FilterTags:         batchFilter,
		ExcludeTags:        batchExclude,
		IncludeQuarantined: batchIncludeQuarantined,
		CompareTo:          batchCompareTo,
		SkipPreflight:      testerSkipPreflight,
		OutputDir:          batchOutputDir,
	}

	if config.Environment == "" {
		config.Environment = "staging"
	}

	runner, err := batch.NewRunner(config)
	if err != nil {
		return fmt.Errorf("failed to create batch runner: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	fmt.Printf("Batch: %s\n", pattern)

	result, err := runner.Run(ctx)
	if err != nil {
		return fmt.Errorf("batch run failed: %w", err)
	}

	if testerJSON {
		data, _ := json.MarshalIndent(result, "", "  ")
		fmt.Println(string(data))
		return nil
	}

	printBatchResult(result)

	// Return error if any tests failed
	if result.Summary.Failed > 0 || result.Summary.Errors > 0 {
		return fmt.Errorf("batch completed with failures")
	}

	return nil
}

func printBatchResult(result *batch.BatchResult) {
	// Print header
	fmt.Printf("  Found: %d scenarios", result.ScenariosFound)
	if result.ScenariosSkipped > 0 {
		fmt.Printf(" (%d quarantined, skipped)", result.ScenariosSkipped)
	}
	fmt.Println()
	fmt.Printf("  Running: %d scenarios\n", result.ScenariosRun)
	fmt.Printf("  Parallel: %d\n", result.Config.Parallel)
	if result.ConvoyID != "" {
		fmt.Printf("  Convoy: %s\n", result.ConvoyID)
	}
	fmt.Println()

	if !result.Config.SkipPreflight {
		fmt.Println("Preflight...")
		fmt.Println("  ✓ All checks passed")
		fmt.Println()
	}

	// Print individual results
	fmt.Println("Running...")
	for _, r := range result.Results {
		printScenarioResult(r)
	}
	fmt.Println()

	// Print summary
	fmt.Println("Batch Complete")
	fmt.Printf("  Passed: %d/%d\n", result.Summary.Passed, result.ScenariosRun)
	fmt.Printf("  Failed: %d/%d\n", result.Summary.Failed, result.ScenariosRun)
	if result.Summary.Skipped > 0 {
		fmt.Printf("  Skipped: %d (quarantined)\n", result.Summary.Skipped)
	}
	fmt.Printf("  Total time: %s", formatDuration(result.TotalDuration))
	if result.Config.Parallel > 1 {
		fmt.Printf(" (parallel)")
	}
	fmt.Println()

	// Print observation summary
	if len(result.Summary.TotalObservations) > 0 {
		var obsStr []string
		total := 0
		for sev, count := range result.Summary.TotalObservations {
			obsStr = append(obsStr, fmt.Sprintf("%d %s", count, sev))
			total += count
		}
		fmt.Printf("  Total observations: %d issues (%s)\n", total, strings.Join(obsStr, ", "))
	}

	if result.Summary.TotalRetries > 0 {
		fmt.Printf("  Retries: %d\n", result.Summary.TotalRetries)
	}
	fmt.Println()

	// Print stability info
	hasStabilityInfo := result.Summary.FlakeRate > 0 ||
		len(result.Summary.AutoQuarantined) > 0 ||
		len(result.Summary.AutoUnquarantined) > 0 ||
		len(result.Summary.FlakyScenarios) > 0 ||
		len(result.Summary.NewQuarantineCandidates) > 0

	if hasStabilityInfo {
		fmt.Println("Stability:")
		if result.Summary.FlakeRate > 0 {
			failCount := result.Summary.Failed + result.Summary.Errors
			fmt.Printf("  Flake rate this batch: %.0f%% (%d/%d)\n",
				result.Summary.FlakeRate*100,
				failCount,
				result.ScenariosRun)
		}

		if len(result.Summary.AutoQuarantined) > 0 {
			fmt.Printf("  Auto-quarantined: %s\n",
				strings.Join(result.Summary.AutoQuarantined, ", "))
		}

		if len(result.Summary.AutoUnquarantined) > 0 {
			fmt.Printf("  Auto-unquarantined: %s\n",
				strings.Join(result.Summary.AutoUnquarantined, ", "))
		}

		if len(result.Summary.FlakyScenarios) > 0 {
			fmt.Printf("  Flaky (flagged): %s\n",
				strings.Join(result.Summary.FlakyScenarios, ", "))
		}

		if len(result.Summary.NewQuarantineCandidates) > 0 {
			fmt.Printf("  Quarantine candidates: %s (investigate)\n",
				strings.Join(result.Summary.NewQuarantineCandidates, ", "))
		}
	}

	// Print output location
	if result.ConvoyID != "" {
		fmt.Printf("\nConvoy: %s\n", result.ConvoyID)
	}
	fmt.Printf("Results: %s\n", result.OutputDir)
}

func printScenarioResult(r batch.ScenarioResult) {
	var status string
	switch r.Status {
	case batch.StatusPassed:
		status = "✓"
	case batch.StatusFailed:
		status = "✗"
	case batch.StatusError:
		status = "✗"
	case batch.StatusSkipped:
		status = "○"
	case batch.StatusRetrying:
		status = "↻"
	default:
		status = "?"
	}

	line := fmt.Sprintf("  %s %s", status, r.Scenario)

	if r.Duration > 0 {
		line += fmt.Sprintf(" (%s)", formatDuration(r.Duration))
	}

	if r.Status == batch.StatusSkipped {
		line += fmt.Sprintf(" - %s", r.SkipReason)
	} else if r.Status == batch.StatusFailed || r.Status == batch.StatusError {
		if r.Error != "" {
			line += fmt.Sprintf(" - %s", r.Error)
		}
	} else if len(r.Observations) > 0 {
		var obsStr []string
		for sev, count := range r.Observations {
			obsStr = append(obsStr, fmt.Sprintf("%d %s", count, sev))
		}
		if len(obsStr) > 0 {
			line += fmt.Sprintf(" - %s", strings.Join(obsStr, ", "))
		}
	}

	if r.RetryCount > 0 {
		line += fmt.Sprintf(" (retry %d)", r.RetryCount)
	}

	fmt.Println(line)
}
