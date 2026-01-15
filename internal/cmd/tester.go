package cmd

import (
	"github.com/spf13/cobra"
)

// Tester command flags
var (
	testerJSON       bool
	testerVerbose    bool
	testerEnv        string
	testerSkipPreflight bool
)

var testerCmd = &cobra.Command{
	Use:     "tester",
	GroupID: GroupDiag,
	Short:   "AI user testing commands",
	RunE:    requireSubcommand,
	Long: `AI User Testing commands for running scenario-based tests.

The tester spawns AI agents (Claude) to act as users navigating your
application, observing UX issues through the eyes of specific personas.

RUNNING TESTS:
  gt tester run <scenario.yaml>      Run a single test scenario
  gt tester preflight                Check environment before testing

MANAGING SCENARIOS:
  gt tester list                     List available scenarios
  gt tester validate <pattern>       Validate scenario files

VIEWING RESULTS:
  gt tester results [date]           View test results
  gt tester review                   Review and validate observations
  gt tester artifacts <run-path>     Open test artifacts

BATCH EXECUTION:
  gt tester batch <pattern>          Run multiple scenarios

STABILITY:
  gt tester flaky                    View flaky test metrics
  gt tester metrics                  View overall stability metrics

Examples:
  gt tester preflight                 # Check if ready to run tests
  gt tester run scenarios/signup.yaml # Run single scenario
  gt tester run signup.yaml --headed  # Run with visible browser
  gt tester batch "**/*.yaml"         # Run all scenarios`,
}

func init() {
	// Register subcommands
	testerCmd.AddCommand(testerPreflightCmd)
	testerCmd.AddCommand(testerRunCmd)

	rootCmd.AddCommand(testerCmd)
}
