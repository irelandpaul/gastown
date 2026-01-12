package cmd

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/steveyegge/gastown/internal/tui/inbox"
)

// Inbox command flags
var (
	inboxOnce bool // Exit after displaying (no interactive mode)
)

var inboxCmd = &cobra.Command{
	Use:     "inbox [address]",
	GroupID: GroupComm,
	Short:   "Interactive inbox for agent messages",
	Long: `View and manage agent messages in an interactive TUI.

The inbox provides a split-view interface with message list and preview pane.
Messages are categorized by type:

  [P] PROPOSAL  - Needs yes/no decision (y=approve, n=reject)
  [Q] QUESTION  - Needs open-ended input (r=reply)
  [!] ALERT     - Urgent attention needed (r=reply, a=acknowledge)
  [i] INFO      - FYI (reports, digests, handoffs) (a=archive)

Actionable messages (PROPOSAL, QUESTION, ALERT) appear first, followed by
INFO messages after a separator.

NAVIGATION:
  ↑/k, ↓/j     Move up/down
  g, G         Go to top/bottom
  pgup, pgdn   Page up/down
  L            Learn message type (classification override)
  q, Esc       Quit

Examples:
  gt inbox                    # Your inbox (auto-detected identity)
  gt inbox mayor/             # Mayor's inbox
  gt inbox gastown/Toast      # Polecat's inbox
  gt inbox --once             # Show and exit (non-interactive)`,
	Args: cobra.MaximumNArgs(1),
	RunE: runInbox,
}

func init() {
	inboxCmd.Flags().BoolVar(&inboxOnce, "once", false, "Show inbox and exit (non-interactive)")

	rootCmd.AddCommand(inboxCmd)
}

func runInbox(cmd *cobra.Command, args []string) error {
	// Determine which inbox to view
	address := ""
	if len(args) > 0 {
		address = args[0]
	} else {
		address = detectSender()
	}

	// Get work directory for beads
	workDir, err := findMailWorkDir()
	if err != nil {
		return fmt.Errorf("not in a Gas Town workspace: %w", err)
	}

	// Non-interactive mode: just show messages and exit
	if inboxOnce {
		return runMailInbox(cmd, args)
	}

	// Interactive TUI mode
	m := inbox.New(address, workDir)
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err = p.Run()
	return err
}
