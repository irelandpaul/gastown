package librarian

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/steveyegge/gastown/internal/beads"
	"github.com/steveyegge/gastown/internal/claude"
	"github.com/steveyegge/gastown/internal/config"
	"github.com/steveyegge/gastown/internal/constants"
	"github.com/steveyegge/gastown/internal/session"
	"github.com/steveyegge/gastown/internal/tmux"
)

// DefaultAgent is the default agent for the Librarian.
// Gemini is chosen for its large context window, ideal for research tasks.
const DefaultAgent = "gemini"

// Common errors
var (
	ErrNotRunning     = errors.New("librarian not running")
	ErrAlreadyRunning = errors.New("librarian already running")
)

// Manager handles Librarian lifecycle operations.
type Manager struct {
	townRoot string
}

// NewManager creates a new librarian manager for a town.
func NewManager(townRoot string) *Manager {
	return &Manager{
		townRoot: townRoot,
	}
}

// SessionName returns the tmux session name for the librarian.
// This is a package-level function for convenience.
func SessionName() string {
	return session.LibrarianSessionName()
}

// SessionName returns the tmux session name for the librarian.
func (m *Manager) SessionName() string {
	return SessionName()
}

// librarianDir returns the working directory for the librarian.
func (m *Manager) librarianDir() string {
	return filepath.Join(m.townRoot, "librarian")
}

// resolveAgent determines which agent to use.
// Returns the override if provided, otherwise returns DefaultAgent (gemini).
func resolveAgent(agentOverride string) string {
	if agentOverride != "" {
		return agentOverride
	}
	return DefaultAgent
}

// Start starts the librarian session.
// agentOverride optionally specifies a different agent alias to use.
// If empty, defaults to Gemini for its large context window.
func (m *Manager) Start(agentOverride string) error {
	t := tmux.NewTmux()
	sessionID := m.SessionName()

	// Check if session already exists
	running, _ := t.HasSession(sessionID)
	if running {
		// Session exists - check if the agent is actually running (healthy vs zombie)
		if t.IsClaudeRunning(sessionID) {
			return ErrAlreadyRunning
		}
		// Zombie - tmux alive but agent dead. Kill and recreate.
		if err := t.KillSession(sessionID); err != nil {
			return fmt.Errorf("killing zombie session: %w", err)
		}
	}

	// Ensure librarian directory exists
	librarianDir := m.librarianDir()
	if err := os.MkdirAll(librarianDir, 0755); err != nil {
		return fmt.Errorf("creating librarian directory: %w", err)
	}

	// Ensure Claude settings exist (works for any agent, not just Claude)
	if err := claude.EnsureSettingsForRole(librarianDir, "librarian"); err != nil {
		return fmt.Errorf("ensuring agent settings: %w", err)
	}

	// Resolve which agent to use (default: gemini)
	agent := resolveAgent(agentOverride)

	// Build startup command - the startup hook handles 'gt prime' automatically
	// Export GT_ROLE and BD_ACTOR in the command since tmux SetEnvironment only affects new panes
	startupCmd, err := config.BuildAgentStartupCommandWithAgentOverride("librarian", "librarian", "", "", agent)
	if err != nil {
		return fmt.Errorf("building startup command: %w", err)
	}

	// Create session with command directly to avoid send-keys race condition.
	// This runs the command as the pane's initial process, avoiding the shell
	// readiness timing issues that cause "bad pattern" and command-not-found errors.
	if err := t.NewSessionWithCommand(sessionID, librarianDir, startupCmd); err != nil {
		return fmt.Errorf("creating tmux session: %w", err)
	}

	// Set environment variables (non-fatal: session works without these)
	// Use centralized AgentEnv for consistency across all role startup paths
	envVars := config.AgentEnv(config.AgentEnvConfig{
		Role:     "librarian",
		TownRoot: m.townRoot,
		BeadsDir: beads.ResolveBeadsDir(m.townRoot),
	})
	for k, v := range envVars {
		_ = t.SetEnvironment(sessionID, k, v)
	}

	// Apply Librarian theming (non-fatal: theming failure doesn't affect operation)
	theme := tmux.LibrarianTheme()
	_ = t.ConfigureGasTownSession(sessionID, theme, "", "Librarian", "librarian")

	// Wait for agent to start (non-fatal)
	if err := t.WaitForCommand(sessionID, constants.SupportedShells, constants.ClaudeStartTimeout); err != nil {
		// Non-fatal - try to continue anyway
	}

	// Accept bypass permissions warning dialog if it appears.
	_ = t.AcceptBypassPermissionsWarning(sessionID)

	time.Sleep(constants.ShutdownNotifyDelay)

	// Inject startup nudge for predecessor discovery via /resume
	_ = session.StartupNudge(t, sessionID, session.StartupNudgeConfig{
		Recipient: "librarian",
		Sender:    "human",
		Topic:     "cold-start",
	}) // Non-fatal

	// GUPP: Gas Town Universal Propulsion Principle
	// Send the propulsion nudge to trigger autonomous execution.
	// Wait for beacon to be fully processed (needs to be separate prompt)
	time.Sleep(2 * time.Second)
	_ = t.NudgeSession(sessionID, session.PropulsionNudgeForRole("librarian", librarianDir)) // Non-fatal

	return nil
}

// Stop stops the librarian session.
func (m *Manager) Stop() error {
	t := tmux.NewTmux()
	sessionID := m.SessionName()

	// Check if session exists
	running, err := t.HasSession(sessionID)
	if err != nil {
		return fmt.Errorf("checking session: %w", err)
	}
	if !running {
		return ErrNotRunning
	}

	// Try graceful shutdown first (best-effort interrupt)
	_ = t.SendKeysRaw(sessionID, "C-c")
	time.Sleep(100 * time.Millisecond)

	// Kill the session
	if err := t.KillSession(sessionID); err != nil {
		return fmt.Errorf("killing session: %w", err)
	}

	return nil
}

// IsRunning checks if the librarian session is active.
func (m *Manager) IsRunning() (bool, error) {
	t := tmux.NewTmux()
	return t.HasSession(m.SessionName())
}

// Status returns information about the librarian session.
func (m *Manager) Status() (*tmux.SessionInfo, error) {
	t := tmux.NewTmux()
	sessionID := m.SessionName()

	running, err := t.HasSession(sessionID)
	if err != nil {
		return nil, fmt.Errorf("checking session: %w", err)
	}
	if !running {
		return nil, ErrNotRunning
	}

	return t.GetSessionInfo(sessionID)
}
