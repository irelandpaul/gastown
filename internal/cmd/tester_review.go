package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/steveyegge/gastown/internal/style"
	"github.com/steveyegge/gastown/internal/ui"
)

// Review command flags
var (
	reviewScenario     string
	reviewDate         string
	reviewInteractive  bool
	reviewValidate     int
	reviewFalsePos     int
	reviewResultsDir   string
)

var testerReviewCmd = &cobra.Command{
	Use:   "review",
	Short: "Review and validate observations",
	Long: `Review and validate observations from AI user testing.

By default, lists all pending observations that need human review.
Use --interactive for step-by-step review mode.

Observations are marked as pending review if they have:
- Low or medium confidence (agent uncertain)
- P0/P1 severity (critical issues need validation)

Examples:
  gt tester review                      # List pending observations
  gt tester review --interactive        # Interactive review mode
  gt tester review --scenario signup    # Filter by scenario
  gt tester review --date 2026-01-15    # Filter by date
  gt tester review --validate 1         # Validate observation #1
  gt tester review --false-positive 2   # Mark #2 as false positive`,
	RunE: runTesterReview,
}

// PendingObservation represents an observation pending review with context
type PendingObservation struct {
	Index       int         `json:"index"`
	Scenario    string      `json:"scenario"`
	RunID       string      `json:"run_id"`
	RunPath     string      `json:"run_path"`
	Observation Observation `json:"observation"`
	ResultFile  string      `json:"result_file"`
}

func init() {
	testerReviewCmd.Flags().StringVar(&reviewScenario, "scenario", "", "Filter by scenario name")
	testerReviewCmd.Flags().StringVar(&reviewDate, "date", "", "Filter by date (YYYY-MM-DD)")
	testerReviewCmd.Flags().BoolVar(&reviewInteractive, "interactive", false, "Interactive review mode")
	testerReviewCmd.Flags().IntVar(&reviewValidate, "validate", 0, "Validate observation by number")
	testerReviewCmd.Flags().IntVar(&reviewFalsePos, "false-positive", 0, "Mark observation as false positive by number")
	testerReviewCmd.Flags().StringVar(&reviewResultsDir, "results-dir", "test-results", "Test results directory")
	testerReviewCmd.Flags().BoolVar(&testerJSON, "json", false, "Output as JSON")

	testerCmd.AddCommand(testerReviewCmd)
}

func runTesterReview(cmd *cobra.Command, args []string) error {
	// Find all pending observations
	pending, err := findPendingObservations(reviewResultsDir, reviewScenario, reviewDate)
	if err != nil {
		return fmt.Errorf("finding observations: %w", err)
	}

	// Handle --validate flag
	if reviewValidate > 0 {
		return validateObservation(pending, reviewValidate, true)
	}

	// Handle --false-positive flag
	if reviewFalsePos > 0 {
		return validateObservation(pending, reviewFalsePos, false)
	}

	// Interactive mode
	if reviewInteractive {
		return runInteractiveReview(pending)
	}

	// Default: list pending observations
	return listPendingObservations(pending)
}

// findPendingObservations scans test results for observations needing review
func findPendingObservations(resultsDir, scenarioFilter, dateFilter string) ([]PendingObservation, error) {
	var pending []PendingObservation
	index := 1

	// Check if results directory exists
	if _, err := os.Stat(resultsDir); os.IsNotExist(err) {
		return pending, nil // No results yet
	}

	// Walk the results directory
	err := filepath.Walk(resultsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}

		// Look for observations.json files
		if info.Name() != "observations.json" {
			return nil
		}

		// Load the observation result
		result, err := LoadObservationResult(path)
		if err != nil {
			return nil // Skip invalid files
		}

		// Apply filters
		if scenarioFilter != "" && !strings.Contains(strings.ToLower(result.Scenario), strings.ToLower(scenarioFilter)) {
			return nil
		}

		if dateFilter != "" {
			// Extract date from path (expected: test-results/YYYY-MM-DD/...)
			pathDate := extractDateFromPath(path)
			if pathDate != dateFilter {
				return nil
			}
		}

		// Find pending observations in this result
		runPath := filepath.Dir(path)
		for _, obs := range result.Observations {
			if obs.Validated == nil && obs.RequiresHumanReview() {
				pending = append(pending, PendingObservation{
					Index:       index,
					Scenario:    result.Scenario,
					RunID:       result.RunID,
					RunPath:     runPath,
					Observation: obs,
					ResultFile:  path,
				})
				index++
			}
		}

		// Also check for observations that haven't been reviewed at all
		for _, obs := range result.Observations {
			if obs.Validated == nil && !obs.RequiresHumanReview() {
				// Include all unreviewed observations, not just those requiring review
				alreadyAdded := false
				for _, p := range pending {
					if p.ResultFile == path && p.Observation.Description == obs.Description {
						alreadyAdded = true
						break
					}
				}
				if !alreadyAdded {
					pending = append(pending, PendingObservation{
						Index:       index,
						Scenario:    result.Scenario,
						RunID:       result.RunID,
						RunPath:     runPath,
						Observation: obs,
						ResultFile:  path,
					})
					index++
				}
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Sort by scenario then by run
	sort.Slice(pending, func(i, j int) bool {
		if pending[i].Scenario != pending[j].Scenario {
			return pending[i].Scenario < pending[j].Scenario
		}
		return pending[i].RunID < pending[j].RunID
	})

	// Re-index after sorting
	for i := range pending {
		pending[i].Index = i + 1
	}

	return pending, nil
}

// extractDateFromPath extracts the date from a test results path
func extractDateFromPath(path string) string {
	// Expected format: test-results/YYYY-MM-DD/scenario/run-xxx/observations.json
	parts := strings.Split(path, string(filepath.Separator))
	for _, part := range parts {
		if len(part) == 10 && part[4] == '-' && part[7] == '-' {
			// Looks like a date
			if _, err := time.Parse("2006-01-02", part); err == nil {
				return part
			}
		}
	}
	return ""
}

// listPendingObservations displays pending observations in non-interactive mode
func listPendingObservations(pending []PendingObservation) error {
	if len(pending) == 0 {
		fmt.Println("\n✓ No observations pending review.")
		return nil
	}

	// JSON output
	if testerJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(pending)
	}

	// Human-readable output
	fmt.Printf("\n%s %d observations\n\n", style.Bold.Render("Pending Review:"), len(pending))

	for _, p := range pending {
		fmt.Printf("%d. %s %s [%s]\n", p.Index, p.Scenario, p.RunID, p.Observation.Timestamp)

		// Format severity with color
		severityStr := fmt.Sprintf("%s/%s %s", p.Observation.Severity, p.Observation.Confidence, p.Observation.Type)
		switch p.Observation.Severity {
		case SeverityP0:
			fmt.Printf("   %s: %s\n", ui.RenderFail(severityStr), p.Observation.Description)
		case SeverityP1:
			fmt.Printf("   %s: %s\n", ui.RenderWarn(severityStr), p.Observation.Description)
		default:
			fmt.Printf("   %s: %s\n", severityStr, p.Observation.Description)
		}

		if p.Observation.Location != "" {
			fmt.Printf("   Location: %s\n", p.Observation.Location)
		}
		if p.Observation.Screenshot != "" {
			fmt.Printf("   Screenshot: %s\n", p.Observation.Screenshot)
		}
		fmt.Println()
		fmt.Printf("   Validate: %s\n", ui.RenderCommand(fmt.Sprintf("gt tester review --validate %d", p.Index)))
		fmt.Printf("   Mark false positive: %s\n", ui.RenderCommand(fmt.Sprintf("gt tester review --false-positive %d", p.Index)))
		fmt.Println()
	}

	return nil
}

// validateObservation marks an observation as validated or false positive
func validateObservation(pending []PendingObservation, index int, isValid bool) error {
	// Find the observation
	var target *PendingObservation
	for i := range pending {
		if pending[i].Index == index {
			target = &pending[i]
			break
		}
	}

	if target == nil {
		return fmt.Errorf("observation #%d not found (valid range: 1-%d)", index, len(pending))
	}

	// Load the result file
	result, err := LoadObservationResult(target.ResultFile)
	if err != nil {
		return fmt.Errorf("loading result file: %w", err)
	}

	// Find and update the observation
	updated := false
	for i := range result.Observations {
		if result.Observations[i].Description == target.Observation.Description &&
			result.Observations[i].Timestamp == target.Observation.Timestamp {
			result.Observations[i].Validated = &isValid
			if !isValid {
				t := true
				result.Observations[i].FalsePositive = &t
			} else {
				f := false
				result.Observations[i].FalsePositive = &f
			}
			updated = true
			break
		}
	}

	if !updated {
		return fmt.Errorf("could not find observation in result file")
	}

	// Write back the result
	if err := result.WriteToFile(filepath.Dir(target.ResultFile)); err != nil {
		return fmt.Errorf("saving result: %w", err)
	}

	// Print confirmation
	if isValid {
		fmt.Printf("\n%s Observation #%d validated.\n", ui.RenderPassIcon(), index)
		fmt.Printf("   %s: %s\n", target.Scenario, target.Observation.Description)
	} else {
		fmt.Printf("\n%s Observation #%d marked as false positive.\n", ui.RenderWarnIcon(), index)
		fmt.Printf("   %s: %s\n", target.Scenario, target.Observation.Description)
	}

	return nil
}

// runInteractiveReview runs the interactive review mode
func runInteractiveReview(pending []PendingObservation) error {
	if len(pending) == 0 {
		fmt.Println("\n✓ No observations pending review.")
		return nil
	}

	fmt.Printf("\n%s Interactive Review Mode\n", style.Bold.Render("🔍"))
	fmt.Printf("   %d observations to review\n\n", len(pending))

	reader := bufio.NewReader(os.Stdin)

	for i, p := range pending {
		fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
		fmt.Printf("Reviewing: %s %s [%s] (%d/%d)\n\n", p.Scenario, p.RunID, p.Observation.Timestamp, i+1, len(pending))

		// Show observation details
		severityStr := fmt.Sprintf("%s/%s %s", p.Observation.Severity, p.Observation.Confidence, p.Observation.Type)
		switch p.Observation.Severity {
		case SeverityP0:
			fmt.Printf("  %s\n", ui.RenderFail(severityStr))
		case SeverityP1:
			fmt.Printf("  %s\n", ui.RenderWarn(severityStr))
		default:
			fmt.Printf("  %s\n", severityStr)
		}

		fmt.Printf("  %s\n", p.Observation.Description)

		if p.Observation.Location != "" {
			fmt.Printf("  Location: %s\n", p.Observation.Location)
		}

		// Try to open screenshot if available
		if p.Observation.Screenshot != "" {
			screenshotPath := filepath.Join(p.RunPath, "screenshots", p.Observation.Screenshot)
			if _, err := os.Stat(screenshotPath); err == nil {
				fmt.Printf("\n  Screenshot: %s\n", p.Observation.Screenshot)
				fmt.Printf("  Opening screenshot...\n")
				openFile(screenshotPath)
			} else {
				fmt.Printf("\n  Screenshot: %s (not found)\n", p.Observation.Screenshot)
			}
		}

		fmt.Println()
		fmt.Println("  [v] Validate  [f] False positive  [s] Skip  [q] Quit")
		fmt.Print("  > ")

		input, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("reading input: %w", err)
		}

		input = strings.TrimSpace(strings.ToLower(input))

		switch input {
		case "v", "validate":
			if err := validateObservation(pending, p.Index, true); err != nil {
				fmt.Printf("  %s Error: %v\n", ui.RenderFailIcon(), err)
			} else {
				fmt.Printf("  %s Validated!\n", ui.RenderPassIcon())
			}

		case "f", "false", "false-positive", "fp":
			if err := validateObservation(pending, p.Index, false); err != nil {
				fmt.Printf("  %s Error: %v\n", ui.RenderFailIcon(), err)
			} else {
				fmt.Printf("  %s Marked as false positive.\n", ui.RenderWarnIcon())
			}

		case "s", "skip":
			fmt.Printf("  %s Skipped.\n", ui.RenderMuted("→"))

		case "q", "quit", "exit":
			fmt.Println("\n  Exiting review mode.")
			return nil

		default:
			fmt.Printf("  %s Unknown command, skipping.\n", ui.RenderWarnIcon())
		}

		fmt.Println()
	}

	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("\n%s Review complete!\n", ui.RenderPassIcon())

	return nil
}

// openFile attempts to open a file with the system default application
func openFile(path string) {
	var cmd *exec.Cmd

	// Try different openers based on what's available
	if _, err := exec.LookPath("xdg-open"); err == nil {
		cmd = exec.Command("xdg-open", path)
	} else if _, err := exec.LookPath("open"); err == nil {
		cmd = exec.Command("open", path)
	} else {
		// Can't open, just print the path
		fmt.Printf("  View: %s\n", path)
		return
	}

	cmd.Start() // Don't wait for completion
}
