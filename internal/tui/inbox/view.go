package inbox

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/charmbracelet/lipgloss"
)

// renderView renders the entire inbox view.
func (m Model) renderView() string {
	if m.width == 0 || m.height == 0 {
		return "Loading..."
	}

	// Render based on current mode
	switch m.mode {
	case ModeReply:
		return m.renderReplyView()
	case ModeThread:
		return m.renderThreadView()
	case ModeExpand:
		return m.renderExpandView()
	default:
		return m.renderListView()
	}
}

// renderListView renders the standard list + preview view.
func (m Model) renderListView() string {
	var b strings.Builder

	// Calculate dimensions
	// Reserve lines for: header (2), footer (2), borders (2)
	contentHeight := m.height - 6
	if contentHeight < 5 {
		contentHeight = 5
	}

	// Split width: 40% list, 60% preview (with divider)
	listWidth := m.width * 40 / 100
	if listWidth < 30 {
		listWidth = 30
	}
	previewWidth := m.width - listWidth - 1 // -1 for divider

	// Render header
	b.WriteString(m.renderHeader())
	b.WriteString("\n")

	// Render main content (list + preview)
	listView := m.renderList(listWidth, contentHeight)
	previewView := m.renderPreview(previewWidth, contentHeight)

	// Join list and preview horizontally
	b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, listView, m.renderDivider(contentHeight), previewView))
	b.WriteString("\n")

	// Render footer
	b.WriteString(m.renderFooter())

	return b.String()
}

// renderHeader renders the inbox header line.
func (m Model) renderHeader() string {
	// Count unread
	unread := 0
	for _, msg := range m.messages {
		if !msg.Read {
			unread++
		}
	}

	title := titleStyle.Render("GT INBOX")
	stats := dimStyle.Render(fmt.Sprintf("%d unread | %d messages", unread, len(m.messages)))

	// Error indicator
	if m.err != nil {
		stats = errorStyle.Render(fmt.Sprintf("Error: %v", m.err))
	}

	// Loading indicator
	if m.loading {
		stats = dimStyle.Render("Loading...")
	}

	return fmt.Sprintf("%s                                    %s", title, stats)
}

// renderList renders the message list pane.
func (m Model) renderList(width, height int) string {
	var b strings.Builder

	if len(m.messages) == 0 {
		if m.loading {
			b.WriteString(dimStyle.Render("Loading messages..."))
		} else if m.err != nil {
			b.WriteString(errorStyle.Render("Failed to load messages"))
		} else {
			b.WriteString(dimStyle.Render("(no messages)"))
		}
		// Pad to fill height
		for i := 1; i < height; i++ {
			b.WriteString("\n")
		}
		return b.String()
	}

	// Separate actionable (PROPOSAL, QUESTION, ALERT) from INFO
	actionable := make([]int, 0)
	info := make([]int, 0)
	for i, msg := range m.messages {
		if msg.IsActionable() {
			actionable = append(actionable, i)
		} else {
			info = append(info, i)
		}
	}

	// Calculate visible range (simple scrolling)
	visibleStart := 0
	visibleEnd := height - 1
	if len(m.messages) > height-1 && m.cursor > height/2 {
		visibleStart = m.cursor - height/2
		visibleEnd = visibleStart + height - 1
	}
	if visibleEnd > len(m.messages) {
		visibleEnd = len(m.messages)
		visibleStart = visibleEnd - height + 1
		if visibleStart < 0 {
			visibleStart = 0
		}
	}

	linesWritten := 0
	showedInfoSeparator := false

	for i := visibleStart; i < visibleEnd && linesWritten < height; i++ {
		// Show INFO separator when transitioning from actionable to info
		if !showedInfoSeparator && len(actionable) > 0 && len(info) > 0 {
			// Check if we're about to show the first INFO item
			isInfo := false
			for _, idx := range info {
				if idx == i {
					isInfo = true
					break
				}
			}
			if isInfo {
				sep := separatorStyle.Render("─────── INFO ───────")
				b.WriteString(truncateString(sep, width))
				b.WriteString("\n")
				linesWritten++
				showedInfoSeparator = true
				if linesWritten >= height {
					break
				}
			}
		}

		msg := m.messages[i]
		isSelected := i == m.cursor

		// Build message line
		line := m.renderMessageLine(&msg, width-2, isSelected)

		if isSelected {
			line = selectedStyle.Render(line)
		}

		b.WriteString(line)
		b.WriteString("\n")
		linesWritten++
	}

	// Pad remaining lines
	for linesWritten < height {
		b.WriteString("\n")
		linesWritten++
	}

	return b.String()
}

// renderMessageLine renders a single message line for the list.
func (m Model) renderMessageLine(msg *Message, width int, selected bool) string {
	// Format: ▸ Subject                    Age  [Type]
	//   or:   ○ Subject                    Age  [Type]

	// Selection indicator
	indicator := "  "
	if selected {
		indicator = "▸ "
	}

	// Age
	age := msg.Age()

	// Badge with color (unless selected, then use selected colors)
	badge := msg.Type.Badge()
	if !selected {
		badge = BadgeStyle(msg.Type).Render(badge)
	}

	// Reply count indicator
	replyIndicator := ""
	if msg.ReplyCount > 0 {
		replyIndicator = fmt.Sprintf(" (%d)", msg.ReplyCount)
	}

	// Calculate available space for subject
	// indicator(2) + subject + "  " + age(4) + "  " + badge(3) + reply
	fixedWidth := 2 + 2 + 4 + 2 + 3 + len(replyIndicator)
	subjectWidth := width - fixedWidth
	if subjectWidth < 10 {
		subjectWidth = 10
	}

	subject := truncateString(msg.Subject, subjectWidth)
	// Pad subject to fixed width
	subject = padRight(subject, subjectWidth)

	return fmt.Sprintf("%s%s  %4s  %s%s", indicator, subject, age, badge, replyIndicator)
}

// renderDivider renders the vertical divider between list and preview.
func (m Model) renderDivider(height int) string {
	var b strings.Builder
	divider := "│"
	for i := 0; i < height; i++ {
		b.WriteString(divider)
		if i < height-1 {
			b.WriteString("\n")
		}
	}
	return dimStyle.Render(b.String())
}

// renderPreview renders the preview pane for the selected message.
func (m Model) renderPreview(width, height int) string {
	var b strings.Builder

	msg := m.SelectedMessage()
	if msg == nil {
		// No message selected
		b.WriteString(dimStyle.Render(" (no message selected)"))
		for i := 1; i < height; i++ {
			b.WriteString("\n")
		}
		return b.String()
	}

	linesWritten := 0

	// Header: Type and ID
	typeHeader := strings.ToUpper(string(msg.Type))
	typeStyle := BadgeStyle(msg.Type)
	header := fmt.Sprintf(" %s", typeStyle.Render(typeHeader))
	idPart := dimStyle.Render(msg.ID)
	headerLine := fmt.Sprintf("%s%s%s", header, strings.Repeat(" ", width-utf8.RuneCountInString(typeHeader)-utf8.RuneCountInString(msg.ID)-2), idPart)
	b.WriteString(truncateString(headerLine, width))
	b.WriteString("\n")
	linesWritten++

	// From line
	fromLine := fmt.Sprintf(" %s %s", previewLabelStyle.Render("From:"), msg.From)
	b.WriteString(truncateString(fromLine, width))
	b.WriteString("\n")
	linesWritten++

	// Bead references line (Phase 3)
	if len(msg.References) > 0 {
		refsLine := fmt.Sprintf(" %s %s",
			previewLabelStyle.Render("Refs:"),
			titleStyle.Render(strings.Join(msg.References, ", ")))
		b.WriteString(truncateString(refsLine, width))
		b.WriteString("\n")
		linesWritten++
	}

	// Separator
	b.WriteString(" " + dimStyle.Render(strings.Repeat("─", width-2)))
	b.WriteString("\n")
	linesWritten++

	// Body content (wrap lines, highlight bead references)
	bodyLines := wrapText(msg.Body, width-2)
	for _, line := range bodyLines {
		if linesWritten >= height-2 { // Reserve space for bottom actions
			break
		}
		// Highlight bead references in the line
		highlightedLine := highlightBeadRefs(line, msg.References)
		b.WriteString(" " + highlightedLine)
		b.WriteString("\n")
		linesWritten++
	}

	// Pad remaining lines
	for linesWritten < height-2 {
		b.WriteString("\n")
		linesWritten++
	}

	// Bottom separator
	b.WriteString(" " + dimStyle.Render(strings.Repeat("─", width-2)))
	b.WriteString("\n")
	linesWritten++

	// Quick actions hint based on message type
	actions := m.getQuickActionsHint(msg)
	b.WriteString(" " + dimStyle.Render(actions))

	return b.String()
}

// getQuickActionsHint returns the quick action hint for a message type.
func (m Model) getQuickActionsHint(msg *Message) string {
	var base string
	switch msg.Type {
	case TypeProposal:
		base = "[y] Approve  [n] Reject  [r] Reply"
	case TypeQuestion:
		base = "[r] Reply  [a] Archive"
	case TypeAlert:
		base = "[r] Reply  [a] Acknowledge"
	case TypeInfo:
		base = "[a] Archive"
	default:
		base = ""
	}

	// Add expand hint if message has bead references
	if len(msg.References) > 0 {
		if base != "" {
			base += "  "
		}
		base += fmt.Sprintf("[e] Expand (%d)", len(msg.References))
	}

	return base
}

// renderFooter renders the help footer.
func (m Model) renderFooter() string {
	// Show status message if present
	if m.statusMsg != "" {
		return titleStyle.Render(m.statusMsg)
	}

	if m.showHelp {
		return m.help.View(m.keys)
	}
	return helpStyle.Render("↑↓ nav | q quit | ? help")
}

// renderReplyView renders the reply composition view.
func (m Model) renderReplyView() string {
	var b strings.Builder

	// Header
	b.WriteString(titleStyle.Render("REPLY"))
	b.WriteString("\n\n")

	// Show what we're replying to
	if m.replyingTo != nil {
		b.WriteString(previewLabelStyle.Render("To: "))
		b.WriteString(m.replyingTo.From)
		b.WriteString("\n")
		b.WriteString(previewLabelStyle.Render("Re: "))
		b.WriteString(m.replyingTo.Subject)
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(dimStyle.Render(strings.Repeat("─", m.width-2)))
	b.WriteString("\n\n")

	// Textarea
	b.WriteString(m.replyInput.View())
	b.WriteString("\n\n")

	// Footer with instructions
	b.WriteString(dimStyle.Render(strings.Repeat("─", m.width-2)))
	b.WriteString("\n")
	b.WriteString(helpStyle.Render("Ctrl+D send | Esc cancel"))

	return b.String()
}

// renderThreadView renders the thread/conversation view.
func (m Model) renderThreadView() string {
	var b strings.Builder

	// Header
	b.WriteString(titleStyle.Render("THREAD"))
	if len(m.threadMessages) > 0 {
		b.WriteString("  ")
		b.WriteString(dimStyle.Render(fmt.Sprintf("(%d messages)", len(m.threadMessages))))
	}
	b.WriteString("\n")
	b.WriteString(dimStyle.Render(strings.Repeat("─", m.width-2)))
	b.WriteString("\n\n")

	// Thread messages (oldest first)
	contentHeight := m.height - 6
	linesUsed := 0

	for i, msg := range m.threadMessages {
		if linesUsed >= contentHeight-3 {
			b.WriteString(dimStyle.Render(fmt.Sprintf("... and %d more messages", len(m.threadMessages)-i)))
			b.WriteString("\n")
			break
		}

		// Message header: From and timestamp
		msgHeader := fmt.Sprintf("%s  %s", msg.From, dimStyle.Render(msg.Age()))
		b.WriteString(previewLabelStyle.Render(msgHeader))
		b.WriteString("\n")
		linesUsed++

		// Message body (truncate if needed)
		bodyLines := wrapText(msg.Body, m.width-4)
		maxBodyLines := 3
		for j, line := range bodyLines {
			if j >= maxBodyLines || linesUsed >= contentHeight-3 {
				if len(bodyLines) > maxBodyLines {
					b.WriteString(dimStyle.Render("  ..."))
					b.WriteString("\n")
					linesUsed++
				}
				break
			}
			b.WriteString("  ")
			b.WriteString(line)
			b.WriteString("\n")
			linesUsed++
		}

		// Separator between messages
		if i < len(m.threadMessages)-1 {
			b.WriteString("\n")
			linesUsed++
		}
	}

	// Pad remaining
	for linesUsed < contentHeight {
		b.WriteString("\n")
		linesUsed++
	}

	// Footer
	b.WriteString(dimStyle.Render(strings.Repeat("─", m.width-2)))
	b.WriteString("\n")
	b.WriteString(helpStyle.Render("r reply | Esc back"))

	return b.String()
}

// renderExpandView renders the expanded bead details view.
func (m Model) renderExpandView() string {
	var b strings.Builder

	// Header
	b.WriteString(titleStyle.Render("BEAD REFERENCES"))
	if len(m.expandedBeads) > 0 {
		b.WriteString("  ")
		b.WriteString(dimStyle.Render(fmt.Sprintf("(%d beads)", len(m.expandedBeads))))
	}
	b.WriteString("\n")
	b.WriteString(dimStyle.Render(strings.Repeat("─", m.width-2)))
	b.WriteString("\n\n")

	// Bead details
	contentHeight := m.height - 6
	linesUsed := 0

	// Visible range for beads (simple scrolling)
	visibleStart := 0
	if m.expandCursor >= contentHeight/4 {
		visibleStart = m.expandCursor - contentHeight/4
	}

	for i := visibleStart; i < len(m.expandedBeads); i++ {
		bead := m.expandedBeads[i]
		if linesUsed >= contentHeight-3 {
			b.WriteString(dimStyle.Render(fmt.Sprintf("... and %d more beads", len(m.expandedBeads)-i)))
			b.WriteString("\n")
			break
		}

		isSelected := i == m.expandCursor
		indicator := "  "
		if isSelected {
			indicator = "▸ "
		}

		// Bead ID and status
		statusColor := dimStyle
		if bead.Status == "open" {
			statusColor = alertBadgeStyle
		} else if bead.Status == "closed" {
			statusColor = infoBadgeStyle
		}

		priorityStr := ""
		switch bead.Priority {
		case 0:
			priorityStr = priorityUrgentStyle.Render(" [URGENT]")
		case 1:
			priorityStr = priorityHighStyle.Render(" [HIGH]")
		case 3:
			priorityStr = priorityLowStyle.Render(" [LOW]")
		}

		beadHeader := fmt.Sprintf("%s%s  %s  %s%s",
			indicator,
			titleStyle.Render(bead.ID),
			statusColor.Render("["+bead.Status+"]"),
			dimStyle.Render(bead.Type),
			priorityStr)

		if isSelected {
			b.WriteString(selectedStyle.Render(padRight(beadHeader, m.width-2)))
		} else {
			b.WriteString(beadHeader)
		}
		b.WriteString("\n")
		linesUsed++

		// Title
		if bead.Title != "" {
			b.WriteString("    ")
			b.WriteString(bead.Title)
			b.WriteString("\n")
			linesUsed++
		}

		// Labels
		if len(bead.Labels) > 0 {
			b.WriteString("    ")
			b.WriteString(previewLabelStyle.Render("Labels: "))
			b.WriteString(dimStyle.Render(strings.Join(bead.Labels, ", ")))
			b.WriteString("\n")
			linesUsed++
		}

		// Description (truncated)
		if bead.Description != "" && linesUsed < contentHeight-3 {
			descLines := wrapText(bead.Description, m.width-6)
			maxDescLines := 2
			for j, line := range descLines {
				if j >= maxDescLines || linesUsed >= contentHeight-3 {
					if len(descLines) > maxDescLines {
						b.WriteString(dimStyle.Render("    ..."))
						b.WriteString("\n")
						linesUsed++
					}
					break
				}
				b.WriteString("    ")
				b.WriteString(dimStyle.Render(line))
				b.WriteString("\n")
				linesUsed++
			}
		}

		// Assignee and CreatedAt
		if (bead.Assignee != "" || bead.CreatedAt != "") && linesUsed < contentHeight-3 {
			infoParts := []string{}
			if bead.Assignee != "" {
				infoParts = append(infoParts, previewLabelStyle.Render("Assignee: ")+bead.Assignee)
			}
			if bead.CreatedAt != "" {
				infoParts = append(infoParts, previewLabelStyle.Render("Created: ")+dimStyle.Render(bead.CreatedAt))
			}
			b.WriteString("    ")
			b.WriteString(strings.Join(infoParts, "  "))
			b.WriteString("\n")
			linesUsed++
		}

		// Separator between beads
		if i < len(m.expandedBeads)-1 {
			b.WriteString("\n")
			linesUsed++
		}
	}

	// Pad remaining
	for linesUsed < contentHeight {
		b.WriteString("\n")
		linesUsed++
	}

	// Footer
	b.WriteString(dimStyle.Render(strings.Repeat("─", m.width-2)))
	b.WriteString("\n")

	helpText := "↑↓ nav | Esc back"
	if m.expandCursor >= 0 && m.expandCursor < len(m.expandedBeads) {
		bead := m.expandedBeads[m.expandCursor]
		if bead.Status == "open" {
			helpText += " | h hook"
		}
	}
	b.WriteString(helpStyle.Render(helpText))

	return b.String()
}

// truncateString truncates a string to maxLen runes, adding "..." if truncated.
func truncateString(s string, maxLen int) string {
	if utf8.RuneCountInString(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return "..."
	}
	runes := []rune(s)
	return string(runes[:maxLen-3]) + "..."
}

// padRight pads a string with spaces to reach the target width.
func padRight(s string, width int) string {
	runeCount := utf8.RuneCountInString(s)
	if runeCount >= width {
		return s
	}
	return s + strings.Repeat(" ", width-runeCount)
}

// highlightBeadRefs highlights bead references in a line of text.
func highlightBeadRefs(line string, refs []string) string {
	if len(refs) == 0 {
		return line
	}

	result := line
	for _, ref := range refs {
		// Replace bead ID with highlighted version
		if strings.Contains(result, ref) {
			highlighted := titleStyle.Render(ref)
			result = strings.ReplaceAll(result, ref, highlighted)
		}
	}
	return result
}

// wrapText wraps text to fit within the given width.
func wrapText(text string, width int) []string {
	if width <= 0 {
		return []string{text}
	}

	var lines []string
	paragraphs := strings.Split(text, "\n")

	for _, para := range paragraphs {
		if para == "" {
			lines = append(lines, "")
			continue
		}

		words := strings.Fields(para)
		if len(words) == 0 {
			lines = append(lines, "")
			continue
		}

		currentLine := words[0]
		for _, word := range words[1:] {
			if utf8.RuneCountInString(currentLine)+1+utf8.RuneCountInString(word) <= width {
				currentLine += " " + word
			} else {
				lines = append(lines, currentLine)
				currentLine = word
			}
		}
		lines = append(lines, currentLine)
	}

	return lines
}
