package aid

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/steveyegge/gastown/internal/claude"
	"github.com/steveyegge/gastown/internal/config"
	"github.com/steveyegge/gastown/internal/constants"
	"github.com/steveyegge/gastown/internal/session"
	"github.com/steveyegge/gastown/internal/tmux"
)

// Common errors
var (
	ErrNotRunning     = errors.New("aid not running")
	ErrAlreadyRunning = errors.New("aid already running")
)

// Manager handles Mayor's Aid lifecycle operations.
type Manager struct {
	townRoot string
}

// NewManager creates a new aid manager for a town.
func NewManager(townRoot string) *Manager {
	return &Manager{
		townRoot: townRoot,
	}
}

// SessionName returns the tmux session name for the aid.
// This is a package-level function for convenience.
func SessionName() string {
	return session.MayorAidSessionName()
}

// SessionName returns the tmux session name for the aid.
func (m *Manager) SessionName() string {
	return SessionName()
}

// aidDir returns the working directory for the aid.
func (m *Manager) aidDir() string {
	return filepath.Join(m.townRoot, "mayor", "aid")
}

// Start starts the aid session.
// agentOverride optionally specifies a different agent alias to use.
func (m *Manager) Start(agentOverride string) error {
	t := tmux.NewTmux()
	sessionID := m.SessionName()

	// Check if session already exists
	running, _ := t.HasSession(sessionID)
	if running {
		// Session exists - check if Claude is actually running (healthy vs zombie)
		if t.IsClaudeRunning(sessionID) {
			return ErrAlreadyRunning
		}
		// Zombie - tmux alive but Claude dead. Kill and recreate.
		if err := t.KillSession(sessionID); err != nil {
			return fmt.Errorf("killing zombie session: %w", err)
		}
	}

	// Ensure aid directory exists
	aidDir := m.aidDir()
	if err := os.MkdirAll(aidDir, 0755); err != nil {
		return fmt.Errorf("creating aid directory: %w", err)
	}

	// Ensure Claude settings exist
	if err := claude.EnsureSettingsForRole(aidDir, "aid"); err != nil {
		return fmt.Errorf("ensuring Claude settings: %w", err)
	}

	// Build startup command first - the startup hook handles 'gt prime' automatically
	// Export GT_ROLE and BD_ACTOR in the command since tmux SetEnvironment only affects new panes
	startupCmd, err := config.BuildAgentStartupCommandWithAgentOverride("aid", "aid", "", "", "", agentOverride)
	if err != nil {
		return fmt.Errorf("building startup command: %w", err)
	}

	// Create session with command directly to avoid send-keys race condition.
	// This runs the command as the pane's initial process, avoiding the shell
	// readiness timing issues that cause "bad pattern" and command-not-found errors.
	if err := t.NewSessionWithCommand(sessionID, aidDir, startupCmd); err != nil {
		return fmt.Errorf("creating tmux session: %w", err)
	}

	// Set environment variables (non-fatal: session works without these)
	// Use centralized AgentEnv for consistency across all role startup paths
	envVars := config.AgentEnv(config.AgentEnvConfig{
		Role:     "aid",
		TownRoot: m.townRoot,
	})
	for k, v := range envVars {
		_ = t.SetEnvironment(sessionID, k, v)
	}

	// Apply Aid theming (non-fatal: theming failure doesn't affect operation)
	theme := tmux.AidTheme()
	_ = t.ConfigureGasTownSession(sessionID, theme, "", "Aid", "aid")

	// Wait for Claude to start (non-fatal)
	if err := t.WaitForCommand(sessionID, constants.SupportedShells, constants.ClaudeStartTimeout); err != nil {
		// Non-fatal - try to continue anyway
	}

	// Accept bypass permissions warning dialog if it appears.
	_ = t.AcceptBypassPermissionsWarning(sessionID)

	time.Sleep(constants.ShutdownNotifyDelay)

	// Inject startup nudge for predecessor discovery via /resume
	_ = session.StartupNudge(t, sessionID, session.StartupNudgeConfig{
		Recipient: "aid",
		Sender:    "human",
		Topic:     "cold-start",
	}) // Non-fatal

	// GUPP: Gas Town Universal Propulsion Principle
	// Send the propulsion nudge to trigger autonomous execution.
	// Wait for beacon to be fully processed (needs to be separate prompt)
	time.Sleep(2 * time.Second)
	_ = t.NudgeSession(sessionID, session.PropulsionNudgeForRole("aid", aidDir)) // Non-fatal

	return nil
}

// Stop stops the aid session.
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

// IsRunning checks if the aid session is active.
func (m *Manager) IsRunning() (bool, error) {
	t := tmux.NewTmux()
	return t.HasSession(m.SessionName())
}

// Status returns information about the aid session.
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
