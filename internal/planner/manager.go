package planner

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/steveyegge/gastown/internal/agent"
	"github.com/steveyegge/gastown/internal/beads"
	"github.com/steveyegge/gastown/internal/rig"
	"github.com/steveyegge/gastown/internal/util"
)

// Common errors
var (
	ErrNoActiveSession = errors.New("no active planning session")
	ErrSessionExists   = errors.New("planning session already exists")
	ErrSessionNotFound = errors.New("planning session not found")
)

// Manager handles planner lifecycle and planning session operations.
type Manager struct {
	rig          *rig.Rig
	workDir      string
	stateManager *agent.StateManager[Planner]
	beads        *beads.Beads
}

// NewManager creates a new planner manager for a rig.
func NewManager(r *rig.Rig) *Manager {
	return &Manager{
		rig:     r,
		workDir: r.Path,
		stateManager: agent.NewStateManager[Planner](r.Path, "planner.json", func() *Planner {
			return &Planner{
				RigName: r.Name,
				State:   StateStopped,
			}
		}),
		beads: beads.New(r.Path),
	}
}

// specsDir returns the path to the .specs directory for this rig.
func (m *Manager) specsDir() string {
	return filepath.Join(m.rig.Path, ".specs")
}

// sessionDir returns the path to a specific planning session's directory.
func (m *Manager) sessionDir(sessionID string) string {
	return filepath.Join(m.specsDir(), sessionID)
}

// EnsureSpecsDir ensures the .specs directory exists.
func (m *Manager) EnsureSpecsDir() error {
	return os.MkdirAll(m.specsDir(), 0755)
}

// Status returns the current planner status.
func (m *Manager) Status() (*Planner, error) {
	return m.stateManager.Load()
}

// ListSessions returns all planning sessions in the .specs directory.
func (m *Manager) ListSessions() ([]*PlanningSession, error) {
	specsDir := m.specsDir()

	// Check if .specs directory exists
	if _, err := os.Stat(specsDir); os.IsNotExist(err) {
		return []*PlanningSession{}, nil
	}

	entries, err := os.ReadDir(specsDir)
	if err != nil {
		return nil, fmt.Errorf("reading specs directory: %w", err)
	}

	var sessions []*PlanningSession
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		session, err := m.LoadSession(entry.Name())
		if err != nil {
			// Skip directories that don't contain valid sessions
			continue
		}
		sessions = append(sessions, session)
	}

	return sessions, nil
}

// LoadSession loads a planning session from disk.
func (m *Manager) LoadSession(sessionID string) (*PlanningSession, error) {
	sessionFile := filepath.Join(m.sessionDir(sessionID), "session.json")

	data, err := os.ReadFile(sessionFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrSessionNotFound
		}
		return nil, fmt.Errorf("reading session file: %w", err)
	}

	var session PlanningSession
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, fmt.Errorf("parsing session file: %w", err)
	}

	return &session, nil
}

// SaveSession saves a planning session to disk.
func (m *Manager) SaveSession(session *PlanningSession) error {
	sessionDir := m.sessionDir(session.ID)
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		return fmt.Errorf("creating session directory: %w", err)
	}

	sessionFile := filepath.Join(sessionDir, "session.json")
	session.UpdatedAt = time.Now()

	return util.AtomicWriteJSON(sessionFile, session)
}

// CreateSession creates a new planning session.
func (m *Manager) CreateSession(title, rawIdea string) (*PlanningSession, error) {
	// Ensure .specs directory exists
	if err := m.EnsureSpecsDir(); err != nil {
		return nil, err
	}

	// Create planning bead first to get the auto-generated ID
	beadOpts := beads.CreateOptions{
		Title:       fmt.Sprintf("Planning: %s", title),
		Type:        "planning", // Will be converted to gt:planning label
		Priority:    2,
		Description: rawIdea,
	}

	bead, err := m.beads.Create(beadOpts)
	if err != nil {
		return nil, fmt.Errorf("creating planning bead: %w", err)
	}

	// Use the bead ID as the session ID
	session := &PlanningSession{
		ID:        bead.ID,
		Title:     title,
		Status:    StatusQuestioning,
		RigName:   m.rig.Name,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		RawIdea:   rawIdea,
	}

	// Create session directory structure
	sessionDir := m.sessionDir(session.ID)
	planningDir := filepath.Join(sessionDir, "planning")
	if err := os.MkdirAll(planningDir, 0755); err != nil {
		return nil, fmt.Errorf("creating planning directory: %w", err)
	}

	// Write raw-idea.md
	rawIdeaPath := filepath.Join(planningDir, "raw-idea.md")
	rawIdeaContent := fmt.Sprintf(`# Raw Idea: %s

**Date**: %s
**Session**: %s

## Description

%s

## Next Steps

This raw idea will be refined into detailed requirements through Q&A.
`, title, time.Now().Format("2006-01-02"), session.ID, rawIdea)

	if err := os.WriteFile(rawIdeaPath, []byte(rawIdeaContent), 0644); err != nil {
		return nil, fmt.Errorf("writing raw-idea.md: %w", err)
	}

	// Save session metadata
	if err := m.SaveSession(session); err != nil {
		return nil, err
	}

	// Update planner state with active session
	planner, err := m.stateManager.Load()
	if err != nil {
		return nil, err
	}
	planner.ActiveSessionID = session.ID
	if err := m.stateManager.Save(planner); err != nil {
		return nil, err
	}

	return session, nil
}

// GetActiveSession returns the currently active planning session.
func (m *Manager) GetActiveSession() (*PlanningSession, error) {
	planner, err := m.stateManager.Load()
	if err != nil {
		return nil, err
	}

	if planner.ActiveSessionID == "" {
		return nil, ErrNoActiveSession
	}

	return m.LoadSession(planner.ActiveSessionID)
}

// CancelSession cancels a planning session.
func (m *Manager) CancelSession(sessionID string) error {
	session, err := m.LoadSession(sessionID)
	if err != nil {
		return err
	}

	session.Status = StatusCancelled
	if err := m.SaveSession(session); err != nil {
		return err
	}

	// Close the bead
	if err := m.beads.CloseWithReason("Planning session cancelled", sessionID); err != nil {
		// Non-fatal: bead might not exist or already be closed
	}

	// Clear active session if this was it
	planner, err := m.stateManager.Load()
	if err != nil {
		return err
	}
	if planner.ActiveSessionID == sessionID {
		planner.ActiveSessionID = ""
		if err := m.stateManager.Save(planner); err != nil {
			return err
		}
	}

	return nil
}

// GetSessionArtifacts returns the paths to all artifacts for a planning session.
func (m *Manager) GetSessionArtifacts(sessionID string) (*SpecArtifacts, error) {
	sessionDir := m.sessionDir(sessionID)

	artifacts := &SpecArtifacts{
		ReviewPaths: make(map[string]string),
	}

	// Check for planning artifacts
	planningDir := filepath.Join(sessionDir, "planning")
	if rawIdea := filepath.Join(planningDir, "raw-idea.md"); fileExists(rawIdea) {
		artifacts.RawIdeaPath = rawIdea
	}
	if requirements := filepath.Join(planningDir, "requirements.md"); fileExists(requirements) {
		artifacts.RequirementsPath = requirements
	}

	// Check for proposal artifacts
	proposalDir := filepath.Join(sessionDir, "proposal")
	if proposal := filepath.Join(proposalDir, "proposal.md"); fileExists(proposal) {
		artifacts.ProposalPath = proposal
	}

	// Check for review artifacts
	reviewsDir := filepath.Join(proposalDir, "reviews")
	reviewFiles := []string{"pm-review.md", "developer-review.md", "security-review.md", "ralph-review.md"}
	for _, rf := range reviewFiles {
		reviewPath := filepath.Join(reviewsDir, rf)
		if fileExists(reviewPath) {
			// Extract agent name from filename (e.g., "pm-review.md" -> "pm")
			agentName := rf[:len(rf)-len("-review.md")]
			artifacts.ReviewPaths[agentName] = reviewPath
		}
	}

	// Check for spec artifacts
	specDir := filepath.Join(sessionDir, "spec")
	if spec := filepath.Join(specDir, "SPEC.md"); fileExists(spec) {
		artifacts.SpecPath = spec
	}
	if tasks := filepath.Join(specDir, "tasks.md"); fileExists(tasks) {
		artifacts.TasksPath = tasks
	}

	return artifacts, nil
}

// generateShortID generates a short random ID for session naming.
func generateShortID() string {
	// Use current time-based ID for uniqueness
	return fmt.Sprintf("%x", time.Now().UnixNano()%0xFFFFFF)
}

// fileExists checks if a file exists.
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
