package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/steveyegge/gastown/internal/planner"
	"github.com/steveyegge/gastown/internal/planneragent"
	"github.com/steveyegge/gastown/internal/rig"
	"github.com/steveyegge/gastown/internal/style"
	"github.com/steveyegge/gastown/internal/workspace"
)

// Planner command flags
var (
	plannerStatusJSON bool
	plannerShowJSON   bool
	plannerRig        string
)

var plannerCmd = &cobra.Command{
	Use:     "planner",
	GroupID: GroupWork,
	Short:   "Plan specs through structured planning",
	RunE:    runPlannerDefault,
	Long: `Plan feature specs through a structured planning process.

Running 'gt planner' with no arguments opens an interactive planner session
where you can describe what you want to build and the planner guides you
through questions to shape the spec.

Subcommands are available for specific operations:
  gt planner start   - Start planner session in background
  gt planner attach  - Attach to running session
  gt planner new     - Create a new planning session record
  gt planner status  - Check session status

This implements the "Plan before you build" discipline for AI-driven development.`,
}

var plannerNewCmd = &cobra.Command{
	Use:   "new <title>",
	Short: "Start a new planning session",
	Long: `Start a new planning session for a feature idea.

Creates a new planning session in the .specs/ directory and a gt:planning bead
to track progress. The planner will ask clarifying questions to shape the spec.

Examples:
  gt planner new "Add user authentication"
  gt planner new "Implement dark mode toggle" --idea "Allow users to switch themes"`,
	Args: cobra.ExactArgs(1),
	RunE: runPlannerNew,
}

var plannerStatusCmd = &cobra.Command{
	Use:   "status [session-id]",
	Short: "Show planning session status",
	Long: `Show the status of a planning session.

If no session ID is provided, shows the active session.

Examples:
  gt planner status
  gt planner status gt-plan-abc123
  gt planner status --json`,
	Args: cobra.MaximumNArgs(1),
	RunE: runPlannerStatus,
}

var plannerShowCmd = &cobra.Command{
	Use:   "show <session-id>",
	Short: "Show planning session details",
	Long: `Show detailed information about a planning session.

Displays the raw idea, questions, answers, and artifact paths.

Examples:
  gt planner show gt-plan-abc123
  gt planner show gt-plan-abc123 --json`,
	Args: cobra.ExactArgs(1),
	RunE: runPlannerShow,
}

var plannerListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all planning sessions",
	Long: `List all planning sessions in the .specs/ directory.

Shows session ID, title, status, and creation date.

Examples:
  gt planner list
  gt planner list --json`,
	RunE: runPlannerList,
}

var plannerCancelCmd = &cobra.Command{
	Use:   "cancel <session-id>",
	Short: "Cancel a planning session",
	Long: `Cancel an active planning session.

Marks the session as cancelled and closes the associated bead.

Examples:
  gt planner cancel gt-plan-abc123`,
	Args: cobra.ExactArgs(1),
	RunE: runPlannerCancel,
}

var plannerAnswerCmd = &cobra.Command{
	Use:   "answer <question-id> <answer>",
	Short: "Answer a clarifying question",
	Long: `Answer a clarifying question in the active planning session.

The planner asks questions to clarify requirements. Use this command
to provide answers that will be incorporated into the spec.

Examples:
  gt planner answer q1 "JWT tokens with refresh"
  gt planner answer q2 "Support Google and GitHub OAuth"`,
	Args: cobra.MinimumNArgs(2),
	RunE: runPlannerAnswer,
}

// Flags for planner new
var plannerNewIdea string

// Flags for planner session management
var plannerAgentOverride string

// Session management commands
var plannerAgentStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the Planner agent session",
	Long: `Start the Planner agent tmux session for a rig.

Creates a new detached tmux session and launches Claude in the rig's .specs/ directory.
The Planner agent helps you shape feature specs through conversation.

Examples:
  gt planner start --rig gastown
  gt planner start --agent gemini`,
	RunE: runPlannerAgentStart,
}

var plannerAgentStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the Planner agent session",
	Long: `Stop the Planner agent tmux session.

Attempts graceful shutdown first (Ctrl-C), then kills the tmux session.`,
	RunE: runPlannerAgentStop,
}

var plannerAgentAttachCmd = &cobra.Command{
	Use:     "attach",
	Aliases: []string{"at"},
	Short:   "Attach to the Planner agent session",
	Long: `Attach to the running Planner agent tmux session.

Attaches the current terminal to the Planner's tmux session.
If the session is not running, it will be started automatically.
Detach with Ctrl-B D.

Examples:
  gt planner attach --rig gastown`,
	RunE: runPlannerAgentAttach,
}

var plannerAgentRestartCmd = &cobra.Command{
	Use:   "restart",
	Short: "Restart the Planner agent session",
	Long: `Restart the Planner agent tmux session.

Stops the current session (if running) and starts a fresh one.`,
	RunE: runPlannerAgentRestart,
}

func init() {
	// Persistent flag for rig selection (applies to all subcommands)
	plannerCmd.PersistentFlags().StringVar(&plannerRig, "rig", "", "Target rig (e.g., gastown, screencoach)")

	// New command flags
	plannerNewCmd.Flags().StringVar(&plannerNewIdea, "idea", "", "Initial idea/description for the feature")

	// Status command flags
	plannerStatusCmd.Flags().BoolVar(&plannerStatusJSON, "json", false, "Output as JSON")

	// Show command flags
	plannerShowCmd.Flags().BoolVar(&plannerShowJSON, "json", false, "Output as JSON")

	// List command flags
	plannerListCmd.Flags().BoolVar(&plannerStatusJSON, "json", false, "Output as JSON")

	// Agent session flags
	plannerAgentStartCmd.Flags().StringVar(&plannerAgentOverride, "agent", "", "Agent alias to use (overrides default)")
	plannerAgentAttachCmd.Flags().StringVar(&plannerAgentOverride, "agent", "", "Agent alias to use (overrides default)")
	plannerAgentRestartCmd.Flags().StringVar(&plannerAgentOverride, "agent", "", "Agent alias to use (overrides default)")

	// Add subcommands
	plannerCmd.AddCommand(plannerNewCmd)
	plannerCmd.AddCommand(plannerStatusCmd)
	plannerCmd.AddCommand(plannerShowCmd)
	plannerCmd.AddCommand(plannerListCmd)
	plannerCmd.AddCommand(plannerCancelCmd)
	plannerCmd.AddCommand(plannerAnswerCmd)

	// Add session management subcommands
	plannerCmd.AddCommand(plannerAgentStartCmd)
	plannerCmd.AddCommand(plannerAgentStopCmd)
	plannerCmd.AddCommand(plannerAgentAttachCmd)
	plannerCmd.AddCommand(plannerAgentRestartCmd)

	rootCmd.AddCommand(plannerCmd)
}

// getPlannerManager creates a planner manager for the current rig.
func getPlannerManager() (*planner.Manager, *rig.Rig, error) {
	// Find town root
	townRoot, err := workspace.FindFromCwdOrError()
	if err != nil {
		return nil, nil, fmt.Errorf("not in a Gas Town workspace: %w", err)
	}

	var rigName string

	// Use --rig flag if provided, otherwise infer from cwd
	if plannerRig != "" {
		rigName = plannerRig
	} else {
		rigName, err = inferRigFromCwd(townRoot)
		if err != nil {
			return nil, nil, fmt.Errorf("could not determine rig (use --rig flag or cd into a rig directory): %w", err)
		}
	}

	_, r, err := getRig(rigName)
	if err != nil {
		return nil, nil, err
	}

	mgr := planner.NewManager(r)
	return mgr, r, nil
}

// runPlannerDefault is the default behavior when 'gt planner' is run without arguments.
// It opens an interactive planner session (auto-starting if needed).
func runPlannerDefault(cmd *cobra.Command, args []string) error {
	return runPlannerAgentAttach(cmd, args)
}

func runPlannerNew(cmd *cobra.Command, args []string) error {
	title := args[0]
	idea := plannerNewIdea
	if idea == "" {
		idea = title // Use title as idea if not provided
	}

	mgr, r, err := getPlannerManager()
	if err != nil {
		return err
	}

	fmt.Printf("Creating planning session in %s...\n", r.Name)

	session, err := mgr.CreateSession(title, idea)
	if err != nil {
		return fmt.Errorf("creating session: %w", err)
	}

	fmt.Printf("%s Planning session created\n", style.Bold.Render("✓"))
	fmt.Printf("  ID: %s\n", session.ID)
	fmt.Printf("  Title: %s\n", session.Title)
	fmt.Printf("  Status: %s\n", style.Dim.Render(string(session.Status)))
	fmt.Printf("\n  %s\n", style.Dim.Render("Use 'gt planner status' to check progress"))
	fmt.Printf("  %s\n", style.Dim.Render("Use 'gt planner answer' to respond to questions"))

	return nil
}

func runPlannerStatus(cmd *cobra.Command, args []string) error {
	mgr, _, err := getPlannerManager()
	if err != nil {
		return err
	}

	var session *planner.PlanningSession

	if len(args) > 0 {
		// Load specific session
		session, err = mgr.LoadSession(args[0])
		if err != nil {
			return fmt.Errorf("loading session: %w", err)
		}
	} else {
		// Load active session
		session, err = mgr.GetActiveSession()
		if err != nil {
			if err == planner.ErrNoActiveSession {
				fmt.Printf("%s No active planning session\n", style.Dim.Render("○"))
				fmt.Printf("  %s\n", style.Dim.Render("Use 'gt planner new' to start one"))
				return nil
			}
			return fmt.Errorf("getting active session: %w", err)
		}
	}

	if plannerStatusJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(session)
	}

	// Human-readable output
	fmt.Printf("%s Planning Session: %s\n\n", style.Bold.Render("📋"), session.ID)
	fmt.Printf("  Title: %s\n", session.Title)

	statusStr := string(session.Status)
	switch session.Status {
	case planner.StatusQuestioning:
		statusStr = style.Bold.Render("● questioning")
	case planner.StatusReviewing:
		statusStr = style.Bold.Render("● reviewing")
	case planner.StatusApproved:
		statusStr = style.Bold.Render("✓ approved")
	case planner.StatusHandedOff:
		statusStr = style.Dim.Render("→ handed off")
	case planner.StatusCancelled:
		statusStr = style.Dim.Render("✗ cancelled")
	}
	fmt.Printf("  Status: %s\n", statusStr)
	fmt.Printf("  Created: %s\n", session.CreatedAt.Format("2006-01-02 15:04"))

	// Show unanswered questions
	unanswered := 0
	for _, q := range session.Questions {
		if q.Answer == "" {
			unanswered++
		}
	}
	if unanswered > 0 {
		fmt.Printf("\n  %s\n", style.Bold.Render("Pending Questions:"))
		for _, q := range session.Questions {
			if q.Answer == "" {
				fmt.Printf("    • [%s] %s\n", q.ID, q.Text)
			}
		}
	}

	return nil
}

func runPlannerShow(cmd *cobra.Command, args []string) error {
	sessionID := args[0]

	mgr, _, err := getPlannerManager()
	if err != nil {
		return err
	}

	session, err := mgr.LoadSession(sessionID)
	if err != nil {
		return fmt.Errorf("loading session: %w", err)
	}

	artifacts, err := mgr.GetSessionArtifacts(sessionID)
	if err != nil {
		return fmt.Errorf("getting artifacts: %w", err)
	}

	if plannerShowJSON {
		output := struct {
			Session   *planner.PlanningSession `json:"session"`
			Artifacts *planner.SpecArtifacts   `json:"artifacts"`
		}{
			Session:   session,
			Artifacts: artifacts,
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(output)
	}

	// Human-readable output
	fmt.Printf("%s Planning Session: %s\n\n", style.Bold.Render("📋"), session.ID)
	fmt.Printf("  Title: %s\n", session.Title)
	fmt.Printf("  Status: %s\n", session.Status)
	fmt.Printf("  Created: %s\n", session.CreatedAt.Format("2006-01-02 15:04"))

	if session.RawIdea != "" {
		fmt.Printf("\n  %s\n", style.Bold.Render("Raw Idea:"))
		// Truncate long ideas
		idea := session.RawIdea
		if len(idea) > 200 {
			idea = idea[:200] + "..."
		}
		fmt.Printf("    %s\n", style.Dim.Render(idea))
	}

	// Show questions and answers
	if len(session.Questions) > 0 {
		fmt.Printf("\n  %s\n", style.Bold.Render("Questions:"))
		for _, q := range session.Questions {
			status := "○"
			if q.Answer != "" {
				status = "✓"
			}
			fmt.Printf("    %s [%s] %s\n", status, q.ID, q.Text)
			if q.Answer != "" {
				fmt.Printf("      → %s\n", style.Dim.Render(q.Answer))
			}
		}
	}

	// Show artifacts
	fmt.Printf("\n  %s\n", style.Bold.Render("Artifacts:"))
	if artifacts.RawIdeaPath != "" {
		fmt.Printf("    • raw-idea.md: %s\n", style.Dim.Render(artifacts.RawIdeaPath))
	}
	if artifacts.RequirementsPath != "" {
		fmt.Printf("    • requirements.md: %s\n", style.Dim.Render(artifacts.RequirementsPath))
	}
	if artifacts.ProposalPath != "" {
		fmt.Printf("    • proposal.md: %s\n", style.Dim.Render(artifacts.ProposalPath))
	}
	if artifacts.SpecPath != "" {
		fmt.Printf("    • SPEC.md: %s\n", style.Dim.Render(artifacts.SpecPath))
	}
	if artifacts.TasksPath != "" {
		fmt.Printf("    • tasks.md: %s\n", style.Dim.Render(artifacts.TasksPath))
	}
	for agent, path := range artifacts.ReviewPaths {
		fmt.Printf("    • %s-review.md: %s\n", agent, style.Dim.Render(path))
	}

	return nil
}

func runPlannerList(cmd *cobra.Command, args []string) error {
	mgr, r, err := getPlannerManager()
	if err != nil {
		return err
	}

	sessions, err := mgr.ListSessions()
	if err != nil {
		return fmt.Errorf("listing sessions: %w", err)
	}

	if plannerStatusJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(sessions)
	}

	if len(sessions) == 0 {
		fmt.Printf("%s No planning sessions in %s\n", style.Dim.Render("○"), r.Name)
		fmt.Printf("  %s\n", style.Dim.Render("Use 'gt planner new' to start one"))
		return nil
	}

	fmt.Printf("%s Planning Sessions: %s\n\n", style.Bold.Render("📋"), r.Name)

	for _, s := range sessions {
		statusIcon := "○"
		switch s.Status {
		case planner.StatusQuestioning, planner.StatusReviewing:
			statusIcon = "●"
		case planner.StatusApproved:
			statusIcon = "✓"
		case planner.StatusHandedOff:
			statusIcon = "→"
		case planner.StatusCancelled:
			statusIcon = "✗"
		}

		ageStr := formatAge(s.CreatedAt)

		fmt.Printf("  %s %s - %s\n", statusIcon, s.ID, s.Title)
		fmt.Printf("    %s | %s\n", style.Dim.Render(string(s.Status)), style.Dim.Render(ageStr))
	}

	return nil
}

func runPlannerCancel(cmd *cobra.Command, args []string) error {
	sessionID := args[0]

	mgr, _, err := getPlannerManager()
	if err != nil {
		return err
	}

	if err := mgr.CancelSession(sessionID); err != nil {
		return fmt.Errorf("cancelling session: %w", err)
	}

	fmt.Printf("%s Planning session %s cancelled\n", style.Bold.Render("✓"), sessionID)
	return nil
}

func runPlannerAnswer(cmd *cobra.Command, args []string) error {
	questionID := args[0]
	answer := strings.Join(args[1:], " ")

	mgr, _, err := getPlannerManager()
	if err != nil {
		return err
	}

	session, err := mgr.GetActiveSession()
	if err != nil {
		if err == planner.ErrNoActiveSession {
			return fmt.Errorf("no active planning session - use 'gt planner new' to start one")
		}
		return fmt.Errorf("getting active session: %w", err)
	}

	// Find and update the question
	found := false
	for i := range session.Questions {
		if session.Questions[i].ID == questionID {
			session.Questions[i].Answer = answer
			now := time.Now()
			session.Questions[i].AnsweredAt = &now
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("question %s not found in session %s", questionID, session.ID)
	}

	if err := mgr.SaveSession(session); err != nil {
		return fmt.Errorf("saving session: %w", err)
	}

	fmt.Printf("%s Answer recorded for question %s\n", style.Bold.Render("✓"), questionID)

	// Check if all questions are answered
	unanswered := 0
	for _, q := range session.Questions {
		if q.Answer == "" {
			unanswered++
		}
	}
	if unanswered > 0 {
		fmt.Printf("  %s\n", style.Dim.Render(fmt.Sprintf("%d questions remaining", unanswered)))
	} else {
		fmt.Printf("  %s\n", style.Dim.Render("All questions answered - ready for review"))
	}

	return nil
}

// getPlannerAgentManager returns a planner agent manager for the current rig.
func getPlannerAgentManager() (*planneragent.Manager, *rig.Rig, error) {
	// Find town root
	townRoot, err := workspace.FindFromCwdOrError()
	if err != nil {
		return nil, nil, fmt.Errorf("not in a Gas Town workspace: %w", err)
	}

	var rigName string

	// Use --rig flag if provided, otherwise infer from cwd
	if plannerRig != "" {
		rigName = plannerRig
	} else {
		rigName, err = inferRigFromCwd(townRoot)
		if err != nil {
			return nil, nil, fmt.Errorf("could not determine rig (use --rig flag or cd into a rig directory): %w", err)
		}
	}

	_, r, err := getRig(rigName)
	if err != nil {
		return nil, nil, err
	}

	mgr := planneragent.NewManager(townRoot, r)
	return mgr, r, nil
}

func runPlannerAgentStart(cmd *cobra.Command, args []string) error {
	mgr, r, err := getPlannerAgentManager()
	if err != nil {
		return err
	}

	fmt.Printf("Starting Planner session for %s...\n", r.Name)
	if err := mgr.Start(plannerAgentOverride); err != nil {
		if err == planneragent.ErrAlreadyRunning {
			return fmt.Errorf("Planner session already running. Attach with: gt planner attach --rig %s", r.Name)
		}
		return err
	}

	fmt.Printf("%s Planner session started. Attach with: %s\n",
		style.Bold.Render("✓"),
		style.Dim.Render(fmt.Sprintf("gt planner attach --rig %s", r.Name)))

	return nil
}

func runPlannerAgentStop(cmd *cobra.Command, args []string) error {
	mgr, r, err := getPlannerAgentManager()
	if err != nil {
		return err
	}

	fmt.Printf("Stopping Planner session for %s...\n", r.Name)
	if err := mgr.Stop(); err != nil {
		if err == planneragent.ErrNotRunning {
			return fmt.Errorf("Planner session is not running")
		}
		return err
	}

	fmt.Printf("%s Planner session stopped.\n", style.Bold.Render("✓"))
	return nil
}

func runPlannerAgentAttach(cmd *cobra.Command, args []string) error {
	mgr, r, err := getPlannerAgentManager()
	if err != nil {
		return err
	}

	running, err := mgr.IsRunning()
	if err != nil {
		return fmt.Errorf("checking session: %w", err)
	}
	if !running {
		// Auto-start if not running
		fmt.Printf("Planner session for %s not running, starting...\n", r.Name)
		if err := mgr.Start(plannerAgentOverride); err != nil {
			return err
		}
	}

	// Use shared attach helper
	return attachToTmuxSession(mgr.SessionName())
}

func runPlannerAgentRestart(cmd *cobra.Command, args []string) error {
	mgr, _, err := getPlannerAgentManager()
	if err != nil {
		return err
	}

	// Stop if running (ignore not-running error)
	if err := mgr.Stop(); err != nil && err != planneragent.ErrNotRunning {
		return fmt.Errorf("stopping session: %w", err)
	}

	// Start fresh
	return runPlannerAgentStart(cmd, args)
}
