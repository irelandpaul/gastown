// Package inbox provides a TUI for viewing and managing agent messages.
package inbox

import (
	"fmt"
	"time"
)

// MessageType indicates the purpose of an inbox message.
// These are the 4 simplified types from the v5 inbox spec.
type MessageType string

const (
	// TypeProposal indicates a message needing yes/no decision.
	// Quick actions: y=approve, n=reject
	TypeProposal MessageType = "proposal"

	// TypeQuestion indicates a message needing open-ended input.
	// Quick actions: r=reply
	TypeQuestion MessageType = "question"

	// TypeAlert indicates urgent attention needed NOW.
	// Quick actions: r=reply, a=acknowledge
	TypeAlert MessageType = "alert"

	// TypeInfo is for FYI messages (reports, digests, handoffs, notes).
	// Quick actions: a=archive
	TypeInfo MessageType = "info"
)

// Badge returns the display badge for the message type.
func (t MessageType) Badge() string {
	switch t {
	case TypeProposal:
		return "[P]"
	case TypeQuestion:
		return "[Q]"
	case TypeAlert:
		return "[!]"
	case TypeInfo:
		return "[i]"
	default:
		return "[?]"
	}
}

// Message represents an inbox message for display in the TUI.
type Message struct {
	// ID is the unique message identifier (beads issue ID).
	ID string

	// Type is the message type (proposal, question, alert, info).
	Type MessageType

	// Subject is the message subject line.
	Subject string

	// Body is the full message content.
	Body string

	// From is the sender address.
	From string

	// Timestamp is when the message was sent.
	Timestamp time.Time

	// Read indicates if the message has been read.
	Read bool

	// ThreadID groups related messages.
	ThreadID string

	// ReplyCount is the number of replies in the thread.
	ReplyCount int

	// References are bead IDs referenced in the message body.
	References []string
}

// Age returns the age of the message as a human-readable string.
func (m *Message) Age() string {
	d := time.Since(m.Timestamp)
	if d < time.Minute {
		return "<1m"
	}
	if d < time.Hour {
		return formatDuration(int(d.Minutes()), "m")
	}
	if d < 24*time.Hour {
		return formatDuration(int(d.Hours()), "h")
	}
	return formatDuration(int(d.Hours()/24), "d")
}

func formatDuration(value int, unit string) string {
	if value == 0 {
		value = 1
	}
	// Use fmt.Sprintf for proper number formatting
	return fmt.Sprintf("%d%s", value, unit)
}

// IsActionable returns true if this message type requires a decision or response.
func (m *Message) IsActionable() bool {
	return m.Type == TypeProposal || m.Type == TypeQuestion || m.Type == TypeAlert
}
