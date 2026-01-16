package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

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

var librarianSkillsCmd = &cobra.Command{
	Use:   "skills",
	Short: "List available skills",
	Long: `List all skills available for dynamic injection.

Skills are YAML files in <town>/librarian/skills/ that define:
- Triggers: conditions that activate the skill (labels, keywords, patterns)
- Content: files, patterns, docs, and notes to inject

Example skill file (librarian/skills/go-testing.yaml):
  id: go-testing
  name: Go Testing
  triggers:
    keywords: ["test", "testing"]
    labels: ["gt:testing"]
  content:
    files:
      - path: "*_test.go"
        description: "Test file patterns"
    patterns:
      - name: Table-driven tests
        description: Use testify/assert with table-driven tests`,
	RunE: runLibrarianSkills,
}

var librarianInjectCmd = &cobra.Command{
	Use:   "inject <bead-id>",
	Short: "Inject skills into a bead's enrichment",
	Long: `Perform dynamic skill injection for a bead.

This command:
1. Analyzes the bead's context (labels, title, description)
2. Matches applicable skills from the skills directory
3. Generates enrichment content with relevant files, patterns, and docs

Use --preview to see which skills would match without generating enrichment.`,
	Args: cobra.ExactArgs(1),
	RunE: runLibrarianInject,
}

var librarianMatchCmd = &cobra.Command{
	Use:   "match <bead-id>",
	Short: "Preview which skills match a bead",
	Long: `Preview skills that would be injected for a bead without generating enrichment.

Shows the bead context and lists all matching skills with their triggers.`,
	Args: cobra.ExactArgs(1),
	RunE: runLibrarianMatch,
}

var (
	injectDepth   string
	injectPreview bool
)

func init() {
	librarianCmd.AddCommand(librarianStartCmd)
	librarianCmd.AddCommand(librarianStopCmd)
	librarianCmd.AddCommand(librarianAttachCmd)
	librarianCmd.AddCommand(librarianStatusCmd)
	librarianCmd.AddCommand(librarianRestartCmd)
	librarianCmd.AddCommand(librarianSkillsCmd)
	librarianCmd.AddCommand(librarianInjectCmd)
	librarianCmd.AddCommand(librarianMatchCmd)

	librarianStartCmd.Flags().StringVar(&librarianAgentOverride, "agent", "", "Agent alias to use (default: gemini)")
	librarianAttachCmd.Flags().StringVar(&librarianAgentOverride, "agent", "", "Agent alias to use (default: gemini)")
	librarianRestartCmd.Flags().StringVar(&librarianAgentOverride, "agent", "", "Agent alias to use (default: gemini)")

	librarianInjectCmd.Flags().StringVar(&injectDepth, "depth", "standard", "Enrichment depth: quick, standard, or deep")
	librarianInjectCmd.Flags().BoolVar(&injectPreview, "preview", false, "Preview matches without generating enrichment")

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

func runLibrarianSkills(cmd *cobra.Command, args []string) error {
	townRoot, err := workspace.FindFromCwdOrError()
	if err != nil {
		return err
	}

	// Get current working directory as rig root
	rigRoot, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getting current directory: %w", err)
	}

	injector := librarian.NewInjector(townRoot, rigRoot)
	skills, err := injector.ListSkills()
	if err != nil {
		return err
	}

	skillsDir := injector.GetSkillsDir()

	if len(skills) == 0 {
		fmt.Printf("%s No skills found\n", style.Dim.Render("○"))
		fmt.Printf("\nCreate skills in: %s\n", style.Dim.Render(skillsDir))
		fmt.Printf("\nExample skill file:\n")
		fmt.Println(style.Dim.Render(`  # skills/go-testing.yaml
  id: go-testing
  name: Go Testing
  triggers:
    keywords: ["test", "testing"]
  content:
    patterns:
      - name: Table-driven tests
        description: Use testify with table-driven tests`))
		return nil
	}

	fmt.Printf("%s %d skills available\n\n", style.Bold.Render("●"), len(skills))

	for _, skill := range skills {
		fmt.Printf("  %s %s\n", style.Bold.Render(skill.ID), style.Dim.Render(fmt.Sprintf("(%s)", skill.Name)))
		if skill.Description != "" {
			fmt.Printf("    %s\n", skill.Description)
		}

		// Show triggers
		triggers := formatTriggers(skill.Triggers)
		if triggers != "" {
			fmt.Printf("    %s %s\n", style.Dim.Render("Triggers:"), triggers)
		}

		// Show content summary
		contentSummary := formatContentSummary(skill.Content)
		if contentSummary != "" {
			fmt.Printf("    %s %s\n", style.Dim.Render("Content:"), contentSummary)
		}
		fmt.Println()
	}

	fmt.Printf("Skills directory: %s\n", style.Dim.Render(skillsDir))
	return nil
}

func runLibrarianInject(cmd *cobra.Command, args []string) error {
	beadID := args[0]

	townRoot, err := workspace.FindFromCwdOrError()
	if err != nil {
		return err
	}

	rigRoot, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getting current directory: %w", err)
	}

	// Parse depth
	var depth librarian.EnrichmentDepth
	switch strings.ToLower(injectDepth) {
	case "quick":
		depth = librarian.DepthQuick
	case "standard":
		depth = librarian.DepthStandard
	case "deep":
		depth = librarian.DepthDeep
	default:
		return fmt.Errorf("invalid depth: %s (use quick, standard, or deep)", injectDepth)
	}

	injector := librarian.NewInjector(townRoot, rigRoot)

	// Preview mode
	if injectPreview {
		skills, ctx, err := injector.PreviewMatches(beadID)
		if err != nil {
			return err
		}
		printMatchPreview(ctx, skills)
		return nil
	}

	// Full injection
	result, err := injector.InjectForBead(beadID, depth)
	if err != nil {
		return err
	}

	// Print result
	fmt.Printf("%s Skill injection complete\n\n", style.Bold.Render("✓"))
	fmt.Printf("  Bead: %s\n", style.Bold.Render(beadID))
	fmt.Printf("  Depth: %s\n", depth)
	fmt.Printf("  Skills matched: %d\n", len(result.MatchedSkills))
	fmt.Printf("  Files: %d | Patterns: %d | Docs: %d\n",
		result.Stats.FilesCount,
		result.Stats.PatternsCount,
		result.Stats.DocsCount)

	if len(result.MatchedSkills) > 0 {
		skillNames := make([]string, len(result.MatchedSkills))
		for i, s := range result.MatchedSkills {
			skillNames[i] = s.Name
		}
		fmt.Printf("  Injected: %s\n", strings.Join(skillNames, ", "))
	}

	fmt.Println()
	fmt.Println(style.Dim.Render("─── Enrichment ───"))
	fmt.Println()
	fmt.Println(result.Enrichment)

	return nil
}

func runLibrarianMatch(cmd *cobra.Command, args []string) error {
	beadID := args[0]

	townRoot, err := workspace.FindFromCwdOrError()
	if err != nil {
		return err
	}

	rigRoot, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getting current directory: %w", err)
	}

	injector := librarian.NewInjector(townRoot, rigRoot)

	skills, ctx, err := injector.PreviewMatches(beadID)
	if err != nil {
		return err
	}

	printMatchPreview(ctx, skills)
	return nil
}

func printMatchPreview(ctx *librarian.BeadContext, skills []*librarian.Skill) {
	fmt.Printf("%s Bead Context\n\n", style.Bold.Render("●"))
	fmt.Printf("  ID: %s\n", style.Bold.Render(ctx.ID))
	fmt.Printf("  Title: %s\n", ctx.Title)
	if len(ctx.Labels) > 0 {
		fmt.Printf("  Labels: %s\n", strings.Join(ctx.Labels, ", "))
	}
	if ctx.Type != "" {
		fmt.Printf("  Type: %s\n", ctx.Type)
	}
	if ctx.ParentID != "" {
		fmt.Printf("  Parent: %s\n", ctx.ParentID)
	}
	if ctx.Description != "" {
		desc := ctx.Description
		if len(desc) > 100 {
			desc = desc[:100] + "..."
		}
		fmt.Printf("  Description: %s\n", style.Dim.Render(desc))
	}

	fmt.Println()

	if len(skills) == 0 {
		fmt.Printf("%s No skills matched\n", style.Dim.Render("○"))
		fmt.Println("\nTip: Create skills in <town>/librarian/skills/ to enable automatic injection.")
		return
	}

	fmt.Printf("%s %d skills matched\n\n", style.Bold.Render("●"), len(skills))

	for _, skill := range skills {
		fmt.Printf("  %s %s\n", style.Bold.Render("→"), skill.Name)
		if skill.Description != "" {
			fmt.Printf("    %s\n", style.Dim.Render(skill.Description))
		}
		triggers := formatTriggers(skill.Triggers)
		if triggers != "" {
			fmt.Printf("    %s %s\n", style.Dim.Render("Matched by:"), triggers)
		}
	}
}

func formatTriggers(t librarian.SkillTriggers) string {
	var parts []string
	if len(t.Labels) > 0 {
		parts = append(parts, fmt.Sprintf("labels=%s", strings.Join(t.Labels, ",")))
	}
	if len(t.Keywords) > 0 {
		parts = append(parts, fmt.Sprintf("keywords=%s", strings.Join(t.Keywords, ",")))
	}
	if len(t.TitlePatterns) > 0 {
		parts = append(parts, fmt.Sprintf("title=%d patterns", len(t.TitlePatterns)))
	}
	if len(t.DescriptionPatterns) > 0 {
		parts = append(parts, fmt.Sprintf("desc=%d patterns", len(t.DescriptionPatterns)))
	}
	if len(t.BeadTypes) > 0 {
		parts = append(parts, fmt.Sprintf("types=%s", strings.Join(t.BeadTypes, ",")))
	}
	return strings.Join(parts, ", ")
}

func formatContentSummary(c librarian.SkillContent) string {
	var parts []string
	if len(c.Files) > 0 {
		parts = append(parts, fmt.Sprintf("%d files", len(c.Files)))
	}
	if len(c.Patterns) > 0 {
		parts = append(parts, fmt.Sprintf("%d patterns", len(c.Patterns)))
	}
	if len(c.Documentation) > 0 {
		parts = append(parts, fmt.Sprintf("%d docs", len(c.Documentation)))
	}
	if len(c.ContextNotes) > 0 {
		parts = append(parts, fmt.Sprintf("%d notes", len(c.ContextNotes)))
	}
	return strings.Join(parts, ", ")
}

// getSkillsPath returns the path to the skills directory for the current workspace.
func getSkillsPath() (string, error) {
	townRoot, err := workspace.FindFromCwdOrError()
	if err != nil {
		return "", err
	}
	return filepath.Join(townRoot, "librarian", "skills"), nil
}
