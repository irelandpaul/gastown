package inbox

import (
	"fmt"
	"os/exec"
	"time"

	"github.com/steveyegge/gastown/internal/beads"
	"github.com/steveyegge/gastown/internal/mail"
)

// approveMessage approves a proposal message.
// This sends a reply with "[APPROVED]" prefix and closes the message.
func approveMessage(msgID, address, workDir string) error {
	router := mail.NewRouter(workDir)
	mailbox, err := router.GetMailbox(address)
	if err != nil {
		return fmt.Errorf("getting mailbox: %w", err)
	}

	// Get the original message
	original, err := mailbox.Get(msgID)
	if err != nil {
		return fmt.Errorf("getting message: %w", err)
	}

	// Send approval reply
	reply := mail.NewReplyMessage(
		address,       // from
		original.From, // to (reply to sender)
		"Re: "+original.Subject,
		"[APPROVED] ✓",
		original,
	)

	if err := router.Send(reply); err != nil {
		return fmt.Errorf("sending approval: %w", err)
	}

	// Mark original as read (closes the message)
	if err := mailbox.MarkRead(msgID); err != nil {
		return fmt.Errorf("marking read: %w", err)
	}

	return nil
}

// rejectMessage rejects a proposal message.
// This sends a reply with "[REJECTED]" prefix and closes the message.
func rejectMessage(msgID, address, workDir string) error {
	router := mail.NewRouter(workDir)
	mailbox, err := router.GetMailbox(address)
	if err != nil {
		return fmt.Errorf("getting mailbox: %w", err)
	}

	// Get the original message
	original, err := mailbox.Get(msgID)
	if err != nil {
		return fmt.Errorf("getting message: %w", err)
	}

	// Send rejection reply
	reply := mail.NewReplyMessage(
		address,       // from
		original.From, // to (reply to sender)
		"Re: "+original.Subject,
		"[REJECTED] ✗",
		original,
	)

	if err := router.Send(reply); err != nil {
		return fmt.Errorf("sending rejection: %w", err)
	}

	// Mark original as read (closes the message)
	if err := mailbox.MarkRead(msgID); err != nil {
		return fmt.Errorf("marking read: %w", err)
	}

	return nil
}

// archiveMessage archives a message.
func archiveMessage(msgID, address, workDir string) error {
	router := mail.NewRouter(workDir)
	mailbox, err := router.GetMailbox(address)
	if err != nil {
		return fmt.Errorf("getting mailbox: %w", err)
	}

	if err := mailbox.Archive(msgID); err != nil {
		return fmt.Errorf("archiving: %w", err)
	}

	return nil
}

// sendReply sends a reply to a message.
func sendReply(original *Message, body, address, workDir string) error {
	router := mail.NewRouter(workDir)

	// Convert inbox.Message to mail.Message for reply
	mailOriginal := &mail.Message{
		ID:       original.ID,
		From:     original.From,
		Subject:  original.Subject,
		Body:     original.Body,
		ThreadID: original.ThreadID,
	}

	// Create reply
	reply := mail.NewReplyMessage(
		address,       // from
		original.From, // to (reply to sender)
		"Re: "+original.Subject,
		body,
		mailOriginal,
	)

	if err := router.Send(reply); err != nil {
		return fmt.Errorf("sending reply: %w", err)
	}

	return nil
}

// loadThreadMessages loads all messages in a thread.
func loadThreadMessages(threadID, address, workDir string) ([]*mail.Message, error) {
	router := mail.NewRouter(workDir)
	mailbox, err := router.GetMailbox(address)
	if err != nil {
		return nil, fmt.Errorf("getting mailbox: %w", err)
	}

	messages, err := mailbox.ListByThread(threadID)
	if err != nil {
		return nil, fmt.Errorf("loading thread: %w", err)
	}

	return messages, nil
}

// convertToInboxMessage converts a mail.Message to an inbox.Message.
func convertToInboxMessage(mm *mail.Message) Message {
	return Message{
		ID:         mm.ID,
		Type:       inferTypeFromMail(mm),
		Subject:    mm.Subject,
		Body:       mm.Body,
		From:       mm.From,
		Timestamp:  mm.Timestamp,
		Read:       mm.Read,
		ThreadID:   mm.ThreadID,
		ReplyCount: 0,
		References: extractReferences(mm.Body),
	}
}

// inferTypeFromMail infers the inbox message type from a mail message.
func inferTypeFromMail(mm *mail.Message) MessageType {
	// Reuse the existing inference logic from data.go
	return inferMessageType(mm)
}

// fetchBeadDetails fetches details for multiple bead IDs.
func fetchBeadDetails(beadIDs []string, workDir string) ([]ExpandedBead, error) {
	if len(beadIDs) == 0 {
		return nil, nil
	}

	b := beads.New(workDir)
	issueMap, err := b.ShowMultiple(beadIDs)
	if err != nil {
		return nil, fmt.Errorf("fetching beads: %w", err)
	}

	var result []ExpandedBead
	for _, id := range beadIDs {
		issue, ok := issueMap[id]
		if !ok {
			// Bead not found, add placeholder
			result = append(result, ExpandedBead{
				ID:    id,
				Title: "(not found)",
			})
			continue
		}

		result = append(result, ExpandedBead{
			ID:          issue.ID,
			Title:       issue.Title,
			Description: issue.Description,
			Status:      issue.Status,
			Type:        issue.Type,
			Priority:    issue.Priority,
			Assignee:    issue.Assignee,
			Labels:      issue.Labels,
			CreatedAt:   issue.CreatedAt,
		})
	}

	return result, nil
}

// hookBead hooks a bead for the current agent.
func hookBead(beadID, address, workDir string) error {
	// Use 'gt sling <bead-id>' to hook the bead to the current agent.
	// This is the unified work dispatch command.
	cmd := exec.Command("gt", "sling", beadID)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("hooking bead %s: %s", beadID, string(output))
	}
	return nil
}

// archiveInfo archives all INFO messages in the inbox.
func archiveInfo(address, workDir string) error {
	router := mail.NewRouter(workDir)
	mailbox, err := router.GetMailbox(address)
	if err != nil {
		return err
	}

	messages, err := mailbox.List()
	if err != nil {
		return err
	}

	for _, mm := range messages {
		if inferMessageType(mm) == TypeInfo {
			if err := mailbox.Archive(mm.ID); err != nil {
				// Continue on error for other messages
				continue
			}
		}
	}

	return nil
}

// markAllRead marks all messages in the inbox as read (closes them in beads).
func markAllRead(address, workDir string) error {
	router := mail.NewRouter(workDir)
	mailbox, err := router.GetMailbox(address)
	if err != nil {
		return err
	}

	messages, err := mailbox.List()
	if err != nil {
		return err
	}

	for _, mm := range messages {
		if err := mailbox.MarkRead(mm.ID); err != nil {
			// Continue on error for other messages
			continue
		}
	}

	return nil
}

// archiveOld archives messages older than 24 hours.
func archiveOld(address, workDir string) error {
	router := mail.NewRouter(workDir)
	mailbox, err := router.GetMailbox(address)
	if err != nil {
		return err
	}

	messages, err := mailbox.List()
	if err != nil {
		return err
	}

	cutoff := time.Now().Add(-24 * time.Hour)
	for _, mm := range messages {
		if mm.Timestamp.Before(cutoff) {
			if err := mailbox.Archive(mm.ID); err != nil {
				// Continue on error for other messages
				continue
			}
		}
	}

	return nil
}
