package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/steveyegge/gastown/internal/aid"
	"github.com/steveyegge/gastown/internal/style"
	"github.com/steveyegge/gastown/internal/workspace"
)

var aidCmd = &cobra.Command{
	Use:     "aid",
	GroupID: GroupAgents,
	Short:   "Manage the Mayor's Aid session",
	RunE:    requireSubcommand,
	Long: `Manage the Mayor's Aid tmux session.

The Mayor's Aid is the tactical executor for Gas Town, running alongside the Mayor
in a dedicated tmux session. While Mayor coordinates and plans, Aid implements.

Key characteristics:
- Works on beads slung by Mayor
- Focused execution context (can compact freely)
- Bead-only work constraint
- Mails Mayor for review when work is complete`,
}

var aidAgentOverride string

var aidStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the Mayor's Aid session",
	Long: `Start the Mayor's Aid tmux session.

Creates a new detached tmux session for the Aid and launches Claude.
The session runs in the ~/gt/mayor/aid directory.`,
	RunE: runAidStart,
}

var aidStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the Mayor's Aid session",
	Long: `Stop the Mayor's Aid tmux session.

Attempts graceful shutdown first (Ctrl-C), then kills the tmux session.`,
	RunE: runAidStop,
}

var aidAttachCmd = &cobra.Command{
	Use:     "attach",
	Aliases: []string{"at"},
	Short:   "Attach to the Mayor's Aid session",
	Long: `Attach to the running Mayor's Aid tmux session.

Attaches the current terminal to the Aid's tmux session.
Detach with Ctrl-B D.`,
	RunE: runAidAttach,
}

var aidStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check Mayor's Aid session status",
	Long:  `Check if the Mayor's Aid tmux session is currently running.`,
	RunE:  runAidStatus,
}

var aidRestartCmd = &cobra.Command{
	Use:   "restart",
	Short: "Restart the Mayor's Aid session",
	Long: `Restart the Mayor's Aid tmux session.

Stops the current session (if running) and starts a fresh one.`,
	RunE: runAidRestart,
}

func init() {
	aidCmd.AddCommand(aidStartCmd)
	aidCmd.AddCommand(aidStopCmd)
	aidCmd.AddCommand(aidAttachCmd)
	aidCmd.AddCommand(aidStatusCmd)
	aidCmd.AddCommand(aidRestartCmd)

	aidStartCmd.Flags().StringVar(&aidAgentOverride, "agent", "", "Agent alias to run the Aid with (overrides town default)")
	aidAttachCmd.Flags().StringVar(&aidAgentOverride, "agent", "", "Agent alias to run the Aid with (overrides town default)")
	aidRestartCmd.Flags().StringVar(&aidAgentOverride, "agent", "", "Agent alias to run the Aid with (overrides town default)")

	rootCmd.AddCommand(aidCmd)
}

// getAidManager returns an aid manager for the current workspace.
func getAidManager() (*aid.Manager, error) {
	townRoot, err := workspace.FindFromCwdOrError()
	if err != nil {
		return nil, fmt.Errorf("not in a Gas Town workspace: %w", err)
	}
	return aid.NewManager(townRoot), nil
}

// getAidSessionName returns the Aid session name.
func getAidSessionName() string {
	return aid.SessionName()
}

func runAidStart(cmd *cobra.Command, args []string) error {
	mgr, err := getAidManager()
	if err != nil {
		return err
	}

	fmt.Println("Starting Mayor's Aid session...")
	if err := mgr.Start(aidAgentOverride); err != nil {
		if err == aid.ErrAlreadyRunning {
			return fmt.Errorf("Aid session already running. Attach with: gt aid attach")
		}
		return err
	}

	fmt.Printf("%s Aid session started. Attach with: %s\n",
		style.Bold.Render("✓"),
		style.Dim.Render("gt aid attach"))

	return nil
}

func runAidStop(cmd *cobra.Command, args []string) error {
	mgr, err := getAidManager()
	if err != nil {
		return err
	}

	fmt.Println("Stopping Mayor's Aid session...")
	if err := mgr.Stop(); err != nil {
		if err == aid.ErrNotRunning {
			return fmt.Errorf("Aid session is not running")
		}
		return err
	}

	fmt.Printf("%s Aid session stopped.\n", style.Bold.Render("✓"))
	return nil
}

func runAidAttach(cmd *cobra.Command, args []string) error {
	mgr, err := getAidManager()
	if err != nil {
		return err
	}

	running, err := mgr.IsRunning()
	if err != nil {
		return fmt.Errorf("checking session: %w", err)
	}
	if !running {
		// Auto-start if not running
		fmt.Println("Aid session not running, starting...")
		if err := mgr.Start(aidAgentOverride); err != nil {
			return err
		}
	}

	// Use shared attach helper (smart: links if inside tmux, attaches if outside)
	return attachToTmuxSession(mgr.SessionName())
}

func runAidStatus(cmd *cobra.Command, args []string) error {
	mgr, err := getAidManager()
	if err != nil {
		return err
	}

	info, err := mgr.Status()
	if err != nil {
		if err == aid.ErrNotRunning {
			fmt.Printf("%s Aid session is %s\n",
				style.Dim.Render("○"),
				"not running")
			fmt.Printf("\nStart with: %s\n", style.Dim.Render("gt aid start"))
			return nil
		}
		return fmt.Errorf("checking status: %w", err)
	}

	status := "detached"
	if info.Attached {
		status = "attached"
	}
	fmt.Printf("%s Aid session is %s\n",
		style.Bold.Render("●"),
		style.Bold.Render("running"))
	fmt.Printf("  Status: %s\n", status)
	fmt.Printf("  Created: %s\n", info.Created)
	fmt.Printf("\nAttach with: %s\n", style.Dim.Render("gt aid attach"))

	return nil
}

func runAidRestart(cmd *cobra.Command, args []string) error {
	mgr, err := getAidManager()
	if err != nil {
		return err
	}

	// Stop if running (ignore not-running error)
	if err := mgr.Stop(); err != nil && err != aid.ErrNotRunning {
		return fmt.Errorf("stopping session: %w", err)
	}

	// Start fresh
	return runAidStart(cmd, args)
}
