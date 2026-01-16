package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/steveyegge/gastown/internal/librarian"
	"github.com/steveyegge/gastown/internal/style"
	"github.com/steveyegge/gastown/internal/workspace"
)

var librarianCmd = &cobra.Command{
	Use:     "librarian",
	GroupID: GroupAgents,
	Short:   "Manage the Librarian session",
	RunE:    requireSubcommand,
	Long: `Manage the Librarian tmux session.

The Librarian is a research-focused agent that enriches beads with relevant
context before polecats start working. It solves the "cold start" problem
where agents waste context window searching for documentation and prior work.

Key characteristics:
- Runs on Gemini by default (large context window for research)
- Researches docs, prior work, and codebase patterns
- Attaches "Required Reading" to beads
- Front-loads context so polecats can execute immediately

Model selection:
  gt librarian start              # Uses Gemini (default)
  gt librarian start --agent claude  # Uses Claude instead`,
}

var librarianAgentOverride string

var librarianStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the Librarian session",
	Long: `Start the Librarian tmux session.

Creates a new detached tmux session for the Librarian and launches the agent.
The session runs in the ~/gt/librarian directory.

By default, Librarian uses Gemini for its large context window, ideal for
research-heavy tasks. Override with --agent for other models.`,
	RunE: runLibrarianStart,
}

var librarianStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the Librarian session",
	Long: `Stop the Librarian tmux session.

Attempts graceful shutdown first (Ctrl-C), then kills the tmux session.`,
	RunE: runLibrarianStop,
}

var librarianAttachCmd = &cobra.Command{
	Use:     "attach",
	Aliases: []string{"at"},
	Short:   "Attach to the Librarian session",
	Long: `Attach to the running Librarian tmux session.

Attaches the current terminal to the Librarian's tmux session.
Detach with Ctrl-B D.`,
	RunE: runLibrarianAttach,
}

var librarianStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check Librarian session status",
	Long:  `Check if the Librarian tmux session is currently running.`,
	RunE:  runLibrarianStatus,
}

var librarianRestartCmd = &cobra.Command{
	Use:   "restart",
	Short: "Restart the Librarian session",
	Long: `Restart the Librarian tmux session.

Stops the current session (if running) and starts a fresh one.`,
	RunE: runLibrarianRestart,
}

func init() {
	librarianCmd.AddCommand(librarianStartCmd)
	librarianCmd.AddCommand(librarianStopCmd)
	librarianCmd.AddCommand(librarianAttachCmd)
	librarianCmd.AddCommand(librarianStatusCmd)
	librarianCmd.AddCommand(librarianRestartCmd)

	librarianStartCmd.Flags().StringVar(&librarianAgentOverride, "agent", "", "Agent alias to use (default: gemini)")
	librarianAttachCmd.Flags().StringVar(&librarianAgentOverride, "agent", "", "Agent alias to use (default: gemini)")
	librarianRestartCmd.Flags().StringVar(&librarianAgentOverride, "agent", "", "Agent alias to use (default: gemini)")

	rootCmd.AddCommand(librarianCmd)
}

// getLibrarianManager returns a librarian manager for the current workspace.
func getLibrarianManager() (*librarian.Manager, error) {
	townRoot, err := workspace.FindFromCwdOrError()
	if err != nil {
		return nil, fmt.Errorf("not in a Gas Town workspace: %w", err)
	}
	return librarian.NewManager(townRoot), nil
}

// getLibrarianSessionName returns the Librarian session name.
func getLibrarianSessionName() string {
	return librarian.SessionName()
}

func runLibrarianStart(cmd *cobra.Command, args []string) error {
	mgr, err := getLibrarianManager()
	if err != nil {
		return err
	}

	agentName := librarianAgentOverride
	if agentName == "" {
		agentName = librarian.DefaultAgent
	}

	fmt.Printf("Starting Librarian session with %s...\n", agentName)
	if err := mgr.Start(librarianAgentOverride); err != nil {
		if err == librarian.ErrAlreadyRunning {
			return fmt.Errorf("Librarian session already running. Attach with: gt librarian attach")
		}
		return err
	}

	fmt.Printf("%s Librarian session started. Attach with: %s\n",
		style.Bold.Render("✓"),
		style.Dim.Render("gt librarian attach"))

	return nil
}

func runLibrarianStop(cmd *cobra.Command, args []string) error {
	mgr, err := getLibrarianManager()
	if err != nil {
		return err
	}

	fmt.Println("Stopping Librarian session...")
	if err := mgr.Stop(); err != nil {
		if err == librarian.ErrNotRunning {
			return fmt.Errorf("Librarian session is not running")
		}
		return err
	}

	fmt.Printf("%s Librarian session stopped.\n", style.Bold.Render("✓"))
	return nil
}

func runLibrarianAttach(cmd *cobra.Command, args []string) error {
	mgr, err := getLibrarianManager()
	if err != nil {
		return err
	}

	running, err := mgr.IsRunning()
	if err != nil {
		return fmt.Errorf("checking session: %w", err)
	}
	if !running {
		// Auto-start if not running
		agentName := librarianAgentOverride
		if agentName == "" {
			agentName = librarian.DefaultAgent
		}
		fmt.Printf("Librarian session not running, starting with %s...\n", agentName)
		if err := mgr.Start(librarianAgentOverride); err != nil {
			return err
		}
	}

	// Use shared attach helper (smart: links if inside tmux, attaches if outside)
	return attachToTmuxSession(mgr.SessionName())
}

func runLibrarianStatus(cmd *cobra.Command, args []string) error {
	mgr, err := getLibrarianManager()
	if err != nil {
		return err
	}

	info, err := mgr.Status()
	if err != nil {
		if err == librarian.ErrNotRunning {
			fmt.Printf("%s Librarian session is %s\n",
				style.Dim.Render("○"),
				"not running")
			fmt.Printf("\nStart with: %s\n", style.Dim.Render("gt librarian start"))
			return nil
		}
		return fmt.Errorf("checking status: %w", err)
	}

	status := "detached"
	if info.Attached {
		status = "attached"
	}
	fmt.Printf("%s Librarian session is %s\n",
		style.Bold.Render("●"),
		style.Bold.Render("running"))
	fmt.Printf("  Status: %s\n", status)
	fmt.Printf("  Model: %s (default)\n", librarian.DefaultAgent)
	fmt.Printf("  Created: %s\n", info.Created)
	fmt.Printf("\nAttach with: %s\n", style.Dim.Render("gt librarian attach"))

	return nil
}

func runLibrarianRestart(cmd *cobra.Command, args []string) error {
	mgr, err := getLibrarianManager()
	if err != nil {
		return err
	}

	// Stop if running (ignore not-running error)
	if err := mgr.Stop(); err != nil && err != librarian.ErrNotRunning {
		return fmt.Errorf("stopping session: %w", err)
	}

	// Start fresh
	return runLibrarianStart(cmd, args)
}
