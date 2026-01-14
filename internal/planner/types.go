// Package planner provides the spec shaping and planning agent.
// The planner transforms vague feature requests into well-shaped specifications
// through a structured Q&A and review process before work is assigned to polecats.
package planner

import (
	"time"

	"github.com/steveyegge/gastown/internal/agent"
)

// Bead type labels for planning workflow.
// These are used with the beads system's label mechanism (gt:TYPE format).
const (
	// LabelPlanning is for active planning sessions where Q&A is ongoing.
	LabelPlanning = "gt:planning"

	// LabelProposal is for draft proposals under review.
	LabelProposal = "gt:proposal"

	// LabelSpec is for approved specifications ready for execution.
	LabelSpec = "gt:spec"
)

// State is an alias for agent.State for backwards compatibility.
type State = agent.State

// State constants - re-exported from agent package for backwards compatibility.
const (
	StateStopped = agent.StateStopped
	StateRunning = agent.StateRunning
	StatePaused  = agent.StatePaused
)

// PlanningStatus represents the status of a planning session.
type PlanningStatus string

const (
	// StatusQuestioning means the planner is asking clarifying questions.
	StatusQuestioning PlanningStatus = "questioning"

	// StatusReviewing means the proposal is under review by review agents.
	StatusReviewing PlanningStatus = "reviewing"

	// StatusApproved means the human has approved the proposal.
	StatusApproved PlanningStatus = "approved"

	// StatusHandedOff means the spec has been sent to Mayor for execution.
	StatusHandedOff PlanningStatus = "handed_off"

	// StatusCancelled means the planning session was cancelled.
	StatusCancelled PlanningStatus = "cancelled"
)

// PlanningSession represents an active planning/shaping session.
type PlanningSession struct {
	// ID is the unique identifier for this planning session (e.g., gt-plan-abc).
	ID string `json:"id"`

	// Title is the human-readable title of the feature being planned.
	Title string `json:"title"`

	// Status is the current status of the planning session.
	Status PlanningStatus `json:"status"`

	// RigName is the rig this planning session is for.
	RigName string `json:"rig_name"`

	// CreatedAt is when the planning session started.
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt is when the planning session was last updated.
	UpdatedAt time.Time `json:"updated_at"`

	// RawIdea is the original feature idea/request.
	RawIdea string `json:"raw_idea,omitempty"`

	// ProposalBeadID is the ID of the proposal bead (if created).
	ProposalBeadID string `json:"proposal_bead_id,omitempty"`

	// SpecBeadID is the ID of the spec bead (if created).
	SpecBeadID string `json:"spec_bead_id,omitempty"`

	// Questions are the clarifying questions asked by the planner.
	Questions []Question `json:"questions,omitempty"`

	// ReviewStatus tracks the status of each review agent.
	ReviewStatus map[string]ReviewResult `json:"review_status,omitempty"`
}

// Question represents a clarifying question from the planner.
type Question struct {
	// ID is the unique identifier for this question.
	ID string `json:"id"`

	// Text is the question text.
	Text string `json:"text"`

	// Answer is the human's answer (empty if not yet answered).
	Answer string `json:"answer,omitempty"`

	// AskedAt is when the question was asked.
	AskedAt time.Time `json:"asked_at"`

	// AnsweredAt is when the question was answered (zero if not answered).
	AnsweredAt *time.Time `json:"answered_at,omitempty"`
}

// ReviewResult represents the result of a review agent's evaluation.
type ReviewResult struct {
	// Agent is the review agent name (pm, developer, security, ralph).
	Agent string `json:"agent"`

	// Status is the review status (approved, needs_changes, rejected, confused).
	Status string `json:"status"`

	// Summary is a brief summary of the review findings.
	Summary string `json:"summary,omitempty"`

	// ReviewedAt is when the review was completed.
	ReviewedAt time.Time `json:"reviewed_at"`
}

// Planner represents the planner agent state for a rig.
type Planner struct {
	// RigName is the rig this planner manages.
	RigName string `json:"rig_name"`

	// State is the current running state.
	State State `json:"state"`

	// StartedAt is when the planner was started.
	StartedAt *time.Time `json:"started_at,omitempty"`

	// ActiveSession is the currently active planning session (if any).
	ActiveSessionID string `json:"active_session_id,omitempty"`
}

// SpecArtifacts represents the artifacts produced by a completed planning session.
type SpecArtifacts struct {
	// RawIdeaPath is the path to raw-idea.md
	RawIdeaPath string `json:"raw_idea_path,omitempty"`

	// RequirementsPath is the path to requirements.md
	RequirementsPath string `json:"requirements_path,omitempty"`

	// ProposalPath is the path to proposal.md
	ProposalPath string `json:"proposal_path,omitempty"`

	// SpecPath is the path to SPEC.md
	SpecPath string `json:"spec_path,omitempty"`

	// TasksPath is the path to tasks.md
	TasksPath string `json:"tasks_path,omitempty"`

	// ReviewPaths maps review agent names to their review file paths
	ReviewPaths map[string]string `json:"review_paths,omitempty"`
}
