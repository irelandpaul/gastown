package planneragent

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
	"github.com/steveyegge/gastown/internal/rig"
	"github.com/steveyegge/gastown/internal/session"
	"github.com/steveyegge/gastown/internal/tmux"
)

// Common errors
var (
	ErrNotRunning     = errors.New("planner not running")
	ErrAlreadyRunning = errors.New("planner already running")
)

// Manager handles Planner agent lifecycle operations.
type Manager struct {
	townRoot string
	rig      *rig.Rig
}

// NewManager creates a new planner manager for a rig.
func NewManager(townRoot string, r *rig.Rig) *Manager {
	return &Manager{
		townRoot: townRoot,
		rig:      r,
	}
}

// SessionName returns the tmux session name for the planner.
func (m *Manager) SessionName() string {
	return session.PlannerSessionName(m.rig.Name)
}

// plannerDir returns the working directory for the planner.
// The planner works from the rig's .specs directory.
func (m *Manager) plannerDir() string {
	return filepath.Join(m.rig.Path, ".specs")
}

// Start starts the planner session.
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

	// Ensure specs directory exists
	specsDir := m.plannerDir()
	if err := os.MkdirAll(specsDir, 0755); err != nil {
		return fmt.Errorf("creating specs directory: %w", err)
	}

	// Ensure Claude settings exist for planner role
	if err := claude.EnsureSettingsForRole(specsDir, "planner"); err != nil {
		return fmt.Errorf("ensuring Claude settings: %w", err)
	}

	// Build startup command
	startupCmd, err := config.BuildAgentStartupCommandWithAgentOverride("planner", "planner", m.rig.Name, "", agentOverride)
	if err != nil {
		return fmt.Errorf("building startup command: %w", err)
	}

	// Create session with command directly
	if err := t.NewSessionWithCommand(sessionID, specsDir, startupCmd); err != nil {
		return fmt.Errorf("creating tmux session: %w", err)
	}

	// Set environment variables
	envVars := config.AgentEnv(config.AgentEnvConfig{
		Role:     "planner",
		Rig:      m.rig.Name,
		TownRoot: m.townRoot,
		BeadsDir: beads.ResolveBeadsDir(m.townRoot),
	})
	for k, v := range envVars {
		_ = t.SetEnvironment(sessionID, k, v)
	}

	// Apply Planner theming
	theme := tmux.PlannerTheme()
	_ = t.ConfigureGasTownSession(sessionID, theme, m.rig.Name, "Planner", "planner")

	// Wait for Claude to start
	if err := t.WaitForCommand(sessionID, constants.SupportedShells, constants.ClaudeStartTimeout); err != nil {
		// Non-fatal - try to continue anyway
	}

	// Accept bypass permissions warning dialog if it appears
	_ = t.AcceptBypassPermissionsWarning(sessionID)

	time.Sleep(constants.ShutdownNotifyDelay)

	// Inject startup nudge for predecessor discovery
	_ = session.StartupNudge(t, sessionID, session.StartupNudgeConfig{
		Recipient: "planner",
		Sender:    "human",
		Topic:     "cold-start",
	})

	// GUPP nudge
	time.Sleep(2 * time.Second)
	_ = t.NudgeSession(sessionID, session.PropulsionNudgeForRole("planner", specsDir))

	return nil
}

// Stop stops the planner session.
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

	// Try graceful shutdown first
	_ = t.SendKeysRaw(sessionID, "C-c")
	time.Sleep(100 * time.Millisecond)

	// Kill the session
	if err := t.KillSession(sessionID); err != nil {
		return fmt.Errorf("killing session: %w", err)
	}

	return nil
}

// IsRunning checks if the planner session is active.
func (m *Manager) IsRunning() (bool, error) {
	t := tmux.NewTmux()
	return t.HasSession(m.SessionName())
}

// Status returns information about the planner session.
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
