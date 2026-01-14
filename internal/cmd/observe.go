package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/steveyegge/gastown/internal/beads"
	"github.com/steveyegge/gastown/internal/style"
)

var (
	observeSource string
	observeLabel  string
)

func init() {
	observeCmd.Flags().StringVar(&observeSource, "source", "", "Source bead ID to link the observation to")
	observeCmd.Flags().StringVar(&observeLabel, "label", "", "Additional label (e.g., gotcha, pattern, config)")
	rootCmd.AddCommand(observeCmd)
}

var observeCmd = &cobra.Command{
	Use:     "observe <message>",
	GroupID: GroupWork,
	Short:   "Capture an observation quickly",
	Long: `Capture observations as beads for knowledge retention.

This is a convenience wrapper around bd create for quick knowledge capture.
Observations are automatically labeled with knowledge:observation.

Examples:
  gt observe "What I learned"
  gt observe "API fails silently on empty input" --source=hq-abc
  gt observe "Config uses YAML not JSON" --source=hq-xyz --label=gotcha
  gt observe "Pattern: use factory for testing" --label=pattern`,
	Args: cobra.ExactArgs(1),
	RunE: runObserve,
}

func runObserve(cmd *cobra.Command, args []string) error {
	message := args[0]

	if message == "" {
		return fmt.Errorf("observation message cannot be empty")
	}

	// Get the current working directory for beads
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getting current directory: %w", err)
	}

	// Initialize beads
	bd := beads.New(beads.ResolveBeadsDir(cwd))

	// Build observation body
	var bodyLines []string

	// Add source reference if provided
	if observeSource != "" {
		bodyLines = append(bodyLines, fmt.Sprintf("Source: [[%s]]", observeSource))
		bodyLines = append(bodyLines, "")
	}

	// Add observer info from BD_ACTOR
	actor := os.Getenv("BD_ACTOR")
	if actor != "" {
		bodyLines = append(bodyLines, fmt.Sprintf("Observed by: %s", actor))
	}

	// Build labels - always include knowledge:observation
	labels := []string{"knowledge:observation"}
	if observeLabel != "" {
		labels = append(labels, observeLabel)
	}

	// Create the observation bead
	// Note: bd create uses --labels for multiple labels (comma-separated)
	// The beads.Create function converts Type to gt:<type> label, so we
	// need to use the Run method directly to set arbitrary labels
	createArgs := []string{
		"create",
		"--json",
		"--title=" + message,
		"--labels=" + strings.Join(labels, ","),
	}

	if len(bodyLines) > 0 {
		createArgs = append(createArgs, "--description="+strings.Join(bodyLines, "\n"))
	}

	if actor != "" {
		createArgs = append(createArgs, "--actor="+actor)
	}

	out, err := bd.Run(createArgs...)
	if err != nil {
		return fmt.Errorf("creating observation: %w", err)
	}

	// Parse the created issue ID from output (JSON format)
	// The output is a JSON object with an "id" field
	var created beads.Issue
	if err := json.Unmarshal(out, &created); err != nil {
		// If JSON parsing fails, just report success without ID
		fmt.Printf("%s Observation recorded\n", style.SuccessPrefix)
		return nil
	}

	fmt.Printf("%s Observation recorded: %s\n", style.SuccessPrefix, style.Bold.Render(created.ID))
	fmt.Printf("  Title: %s\n", message)
	if observeSource != "" {
		fmt.Printf("  Source: %s\n", observeSource)
	}
	if observeLabel != "" {
		fmt.Printf("  Label: %s\n", observeLabel)
	}

	return nil
}
