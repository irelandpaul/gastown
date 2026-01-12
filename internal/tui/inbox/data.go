package inbox

import (
	"regexp"
	"strings"

	"github.com/steveyegge/gastown/internal/mail"
)

// loadMessages loads messages from the mailbox and converts them to inbox Messages.
func loadMessages(address, workDir string, ls *LearningSystem) ([]Message, []string, error) {
	// Get mailbox
	router := mail.NewRouter(workDir)
	mailbox, err := router.GetMailbox(address)
	if err != nil {
		return nil, nil, err
	}

	// Get all messages
	mailMessages, err := mailbox.List()
	if err != nil {
		return nil, nil, err
	}

	// Convert to inbox messages
	messages := make([]Message, 0, len(mailMessages))
	for _, mm := range mailMessages {
		msg := convertMailMessage(mm, ls)
		messages = append(messages, msg)
	}

	// Phase 5: Replace mode for INFO (status updates don't stack)
	filtered, toArchive := filterStackedInfo(messages)

	// Sort: actionable first (by type priority), then INFO
	// Within each group, newest first
	sortMessages(filtered)

	return filtered, toArchive, nil
}

// convertMailMessage converts a mail.Message to an inbox.Message.
func convertMailMessage(mm *mail.Message, ls *LearningSystem) Message {
	msg := Message{
		ID:         mm.ID,
		Type:       InferMessageType(mm),
		Subject:    mm.Subject,
		Body:       mm.Body,
		From:       mm.From,
		Timestamp:  mm.Timestamp,
		Read:       mm.Read,
		ThreadID:   mm.ThreadID,
		ReplyCount: 0, // TODO: count thread replies
		References: extractReferences(mm.Body),
	}

	// Apply learning system overrides
	if ls != nil {
		if overriddenType, ok := ls.Classify(msg); ok {
			msg.Type = overriddenType
		}
	}

	return msg
}

// InferMessageType infers the inbox message type from a mail message.
// Uses subject prefixes and message type to determine the category.
func InferMessageType(mm *mail.Message) MessageType {
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

// beadIDRegex matches bead IDs like gt-123, hq-cv-abc, bd-xyz.
// Pattern: 2-4 lowercase letters, followed by one or more hyphenated alphanumeric segments.
var beadIDRegex = regexp.MustCompile(`\b[a-z]{2,4}(-[a-z0-9.]+)+\b`)

// extractReferences extracts bead IDs referenced in the message body.
// Looks for patterns like gt-abc, bd-xyz, hq-123, sc-456.
func extractReferences(body string) []string {
	matches := beadIDRegex.FindAllString(body, -1)
	if len(matches) == 0 {
		return nil
	}

	// Use a map to de-duplicate and preserve order
	seen := make(map[string]bool)
	var refs []string
	for _, match := range matches {
		// Bead IDs are typically short (e.g. gt-123, hq-cv-abc)
		// We limit to 25 characters to avoid matching long hyphenated text.
		if len(match) > 25 {
			continue
		}
		if !seen[match] {
			seen[match] = true
			refs = append(refs, match)
		}
	}

	return refs
}

// filterStackedInfo removes older INFO messages from the same sender with the same subject.
// This implements Phase 5 "Replace mode for INFO".
// Returns filtered messages and messages that should be archived.
func filterStackedInfo(messages []Message) ([]Message, []string) {
	type key struct {
		from    string
		subject string
	}

	// Track newest message for each (from, subject) pair
	newest := make(map[key]Message)
	var toArchive []string

	for _, msg := range messages {
		if msg.Type != TypeInfo {
			continue
		}

		k := key{from: msg.From, subject: msg.Subject}
		if existing, ok := newest[k]; ok {
			if msg.Timestamp.After(existing.Timestamp) {
				toArchive = append(toArchive, existing.ID)
				newest[k] = msg
			} else {
				toArchive = append(toArchive, msg.ID)
			}
		} else {
			newest[k] = msg
		}
	}

	var filtered []Message
	for _, msg := range messages {
		if msg.Type != TypeInfo {
			filtered = append(filtered, msg)
			continue
		}
		// Only keep if it's the newest for its key
		k := key{from: msg.From, subject: msg.Subject}
		if newest[k].ID == msg.ID {
			filtered = append(filtered, msg)
		}
	}

	return filtered, toArchive
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
