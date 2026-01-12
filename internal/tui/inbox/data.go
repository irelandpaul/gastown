package inbox

import (
	"strings"

	"github.com/steveyegge/gastown/internal/mail"
)

// loadMessages loads messages from the mailbox and converts them to inbox Messages.
func loadMessages(address, workDir string) ([]Message, error) {
	// Get mailbox
	router := mail.NewRouter(workDir)
	mailbox, err := router.GetMailbox(address)
	if err != nil {
		return nil, err
	}

	// Get all messages
	mailMessages, err := mailbox.List()
	if err != nil {
		return nil, err
	}

	// Convert to inbox messages
	messages := make([]Message, 0, len(mailMessages))
	for _, mm := range mailMessages {
		msg := convertMailMessage(mm)
		messages = append(messages, msg)
	}

	// Sort: actionable first (by type priority), then INFO
	// Within each group, newest first
	sortMessages(messages)

	return messages, nil
}

// convertMailMessage converts a mail.Message to an inbox.Message.
func convertMailMessage(mm *mail.Message) Message {
	return Message{
		ID:         mm.ID,
		Type:       inferMessageType(mm),
		Subject:    mm.Subject,
		Body:       mm.Body,
		From:       mm.From,
		Timestamp:  mm.Timestamp,
		Read:       mm.Read,
		ThreadID:   mm.ThreadID,
		ReplyCount: 0, // TODO: count thread replies
		References: extractReferences(mm.Body),
	}
}

// inferMessageType infers the inbox message type from a mail message.
// Uses subject prefixes and message type to determine the category.
func inferMessageType(mm *mail.Message) MessageType {
	subject := strings.ToLower(mm.Subject)

	// Check for explicit type markers in subject
	if strings.HasPrefix(subject, "[proposal]") ||
		strings.HasPrefix(subject, "proposal:") ||
		strings.Contains(subject, "approve?") ||
		strings.Contains(subject, "permission to") ||
		strings.Contains(subject, "ready to deploy") {
		return TypeProposal
	}

	if strings.HasPrefix(subject, "[question]") ||
		strings.HasPrefix(subject, "question:") ||
		strings.HasSuffix(subject, "?") && !strings.Contains(subject, "approve") {
		return TypeQuestion
	}

	if strings.HasPrefix(subject, "[alert]") ||
		strings.HasPrefix(subject, "alert:") ||
		strings.HasPrefix(subject, "[!]") ||
		strings.Contains(subject, "urgent") ||
		strings.Contains(subject, "stuck") ||
		strings.Contains(subject, "error") ||
		strings.Contains(subject, "failed") {
		return TypeAlert
	}

	// Check mail message type
	if mm.Type == mail.TypeTask {
		// Tasks that need decisions are proposals
		if strings.Contains(subject, "deploy") ||
			strings.Contains(subject, "release") ||
			strings.Contains(subject, "merge") {
			return TypeProposal
		}
		return TypeQuestion
	}

	// Default to INFO for notifications and replies
	return TypeInfo
}

// extractReferences extracts bead IDs referenced in the message body.
// Looks for patterns like gt-abc, bd-xyz, hq-123, sc-456.
func extractReferences(body string) []string {
	// Simple pattern matching for bead IDs
	// Format: 2-3 lowercase letters + hyphen + alphanumeric
	var refs []string
	words := strings.Fields(body)

	for _, word := range words {
		// Clean punctuation
		word = strings.Trim(word, ".,;:!?()[]{}\"'")

		// Check if it looks like a bead ID
		if len(word) >= 5 && len(word) <= 20 {
			hyphenIdx := strings.Index(word, "-")
			if hyphenIdx >= 2 && hyphenIdx <= 3 {
				prefix := word[:hyphenIdx]
				allLower := true
				for _, c := range prefix {
					if c < 'a' || c > 'z' {
						allLower = false
						break
					}
				}
				if allLower {
					refs = append(refs, word)
				}
			}
		}
	}

	return refs
}

// sortMessages sorts messages with actionable items first, then INFO.
// Within each group, newest first.
func sortMessages(messages []Message) {
	// Simple bubble sort (messages list is typically small)
	n := len(messages)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			if shouldSwap(messages[j], messages[j+1]) {
				messages[j], messages[j+1] = messages[j+1], messages[j]
			}
		}
	}
}

// shouldSwap returns true if a should come after b in the sorted order.
func shouldSwap(a, b Message) bool {
	// Priority order: ALERT > PROPOSAL > QUESTION > INFO
	aPriority := typePriority(a.Type)
	bPriority := typePriority(b.Type)

	if aPriority != bPriority {
		return aPriority > bPriority // Higher priority = lower number, should come first
	}

	// Same type: newer first
	return a.Timestamp.Before(b.Timestamp)
}

// typePriority returns the sort priority for a message type.
// Lower number = higher priority (appears first).
func typePriority(t MessageType) int {
	switch t {
	case TypeAlert:
		return 0
	case TypeProposal:
		return 1
	case TypeQuestion:
		return 2
	case TypeInfo:
		return 3
	default:
		return 4
	}
}
