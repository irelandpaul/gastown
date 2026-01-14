package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/steveyegge/gastown/internal/planner"
	"github.com/steveyegge/gastown/internal/rig"
	"github.com/steveyegge/gastown/internal/style"
	"github.com/steveyegge/gastown/internal/workspace"
)

// Shape command flags
var (
	shapeStatusJSON bool
	shapeShowJSON   bool
)

var shapeCmd = &cobra.Command{
	Use:     "shape",
	GroupID: GroupWork,
	Short:   "Shape specs through structured planning",
	RunE:    requireSubcommand,
	Long: `Shape feature specs through a structured planning process.

The shape command family manages the spec shaping workflow:
1. Start a new planning session with an idea
2. Answer clarifying questions
3. Review the generated proposal
4. Approve and generate the final spec

This implements the "Shape before you build" discipline for AI-driven development.`,
}

var shapeNewCmd = &cobra.Command{
	Use:   "new <title>",
	Short: "Start a new planning session",
	Long: `Start a new planning session for a feature idea.

Creates a new planning session in the .specs/ directory and a gt:planning bead
to track progress. The planner will ask clarifying questions to shape the spec.

Examples:
  gt shape new "Add user authentication"
  gt shape new "Implement dark mode toggle" --idea "Allow users to switch themes"`,
	Args: cobra.ExactArgs(1),
	RunE: runShapeNew,
}

var shapeStatusCmd = &cobra.Command{
	Use:   "status [session-id]",
	Short: "Show planning session status",
	Long: `Show the status of a planning session.

If no session ID is provided, shows the active session.

Examples:
  gt shape status
  gt shape status gt-plan-abc123
  gt shape status --json`,
	Args: cobra.MaximumNArgs(1),
	RunE: runShapeStatus,
}

var shapeShowCmd = &cobra.Command{
	Use:   "show <session-id>",
	Short: "Show planning session details",
	Long: `Show detailed information about a planning session.

Displays the raw idea, questions, answers, and artifact paths.

Examples:
  gt shape show gt-plan-abc123
  gt shape show gt-plan-abc123 --json`,
	Args: cobra.ExactArgs(1),
	RunE: runShapeShow,
}

var shapeListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all planning sessions",
	Long: `List all planning sessions in the .specs/ directory.

Shows session ID, title, status, and creation date.

Examples:
  gt shape list
  gt shape list --json`,
	RunE: runShapeList,
}

var shapeCancelCmd = &cobra.Command{
	Use:   "cancel <session-id>",
	Short: "Cancel a planning session",
	Long: `Cancel an active planning session.

Marks the session as cancelled and closes the associated bead.

Examples:
  gt shape cancel gt-plan-abc123`,
	Args: cobra.ExactArgs(1),
	RunE: runShapeCancel,
}

var shapeAnswerCmd = &cobra.Command{
	Use:   "answer <question-id> <answer>",
	Short: "Answer a clarifying question",
	Long: `Answer a clarifying question in the active planning session.

The planner asks questions to clarify requirements. Use this command
to provide answers that will be incorporated into the spec.

Examples:
  gt shape answer q1 "JWT tokens with refresh"
  gt shape answer q2 "Support Google and GitHub OAuth"`,
	Args: cobra.MinimumNArgs(2),
	RunE: runShapeAnswer,
}

// Flags for shape new
var shapeNewIdea string

func init() {
	// New command flags
	shapeNewCmd.Flags().StringVar(&shapeNewIdea, "idea", "", "Initial idea/description for the feature")

	// Status command flags
	shapeStatusCmd.Flags().BoolVar(&shapeStatusJSON, "json", false, "Output as JSON")

	// Show command flags
	shapeShowCmd.Flags().BoolVar(&shapeShowJSON, "json", false, "Output as JSON")

	// List command flags
	shapeListCmd.Flags().BoolVar(&shapeStatusJSON, "json", false, "Output as JSON")

	// Add subcommands
	shapeCmd.AddCommand(shapeNewCmd)
	shapeCmd.AddCommand(shapeStatusCmd)
	shapeCmd.AddCommand(shapeShowCmd)
	shapeCmd.AddCommand(shapeListCmd)
	shapeCmd.AddCommand(shapeCancelCmd)
	shapeCmd.AddCommand(shapeAnswerCmd)

	rootCmd.AddCommand(shapeCmd)
}

// getPlannerManager creates a planner manager for the current rig.
func getPlannerManager() (*planner.Manager, *rig.Rig, error) {
	// Find town root
	townRoot, err := workspace.FindFromCwdOrError()
	if err != nil {
		return nil, nil, fmt.Errorf("not in a Gas Town workspace: %w", err)
	}

	// Infer rig from cwd
	rigName, err := inferRigFromCwd(townRoot)
	if err != nil {
		return nil, nil, fmt.Errorf("could not determine rig: %w", err)
	}

	_, r, err := getRig(rigName)
	if err != nil {
		return nil, nil, err
	}

	mgr := planner.NewManager(r)
	return mgr, r, nil
}

func runShapeNew(cmd *cobra.Command, args []string) error {
	title := args[0]
	idea := shapeNewIdea
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
	fmt.Printf("\n  %s\n", style.Dim.Render("Use 'gt shape status' to check progress"))
	fmt.Printf("  %s\n", style.Dim.Render("Use 'gt shape answer' to respond to questions"))

	return nil
}

func runShapeStatus(cmd *cobra.Command, args []string) error {
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
				fmt.Printf("  %s\n", style.Dim.Render("Use 'gt shape new' to start one"))
				return nil
			}
			return fmt.Errorf("getting active session: %w", err)
		}
	}

	if shapeStatusJSON {
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

func runShapeShow(cmd *cobra.Command, args []string) error {
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

	if shapeShowJSON {
		output := struct {
			Session   *planner.PlanningSession  `json:"session"`
			Artifacts *planner.SpecArtifacts `json:"artifacts"`
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

func runShapeList(cmd *cobra.Command, args []string) error {
	mgr, r, err := getPlannerManager()
	if err != nil {
		return err
	}

	sessions, err := mgr.ListSessions()
	if err != nil {
		return fmt.Errorf("listing sessions: %w", err)
	}

	if shapeStatusJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(sessions)
	}

	if len(sessions) == 0 {
		fmt.Printf("%s No planning sessions in %s\n", style.Dim.Render("○"), r.Name)
		fmt.Printf("  %s\n", style.Dim.Render("Use 'gt shape new' to start one"))
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

func runShapeCancel(cmd *cobra.Command, args []string) error {
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

func runShapeAnswer(cmd *cobra.Command, args []string) error {
	questionID := args[0]
	answer := strings.Join(args[1:], " ")

	mgr, _, err := getPlannerManager()
	if err != nil {
		return err
	}

	session, err := mgr.GetActiveSession()
	if err != nil {
		if err == planner.ErrNoActiveSession {
			return fmt.Errorf("no active planning session - use 'gt shape new' to start one")
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

