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

	// Separator
	b.WriteString(" " + dimStyle.Render(strings.Repeat("─", width-2)))
	b.WriteString("\n")
	linesWritten++

	// Body content (wrap lines)
	bodyLines := wrapText(msg.Body, width-2)
	for _, line := range bodyLines {
		if linesWritten >= height-2 { // Reserve space for bottom actions
			break
		}
		b.WriteString(" " + line)
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
	switch msg.Type {
	case TypeProposal:
		return "[y] Approve  [n] Reject  [Enter]"
	case TypeQuestion:
		return "[r] Reply  [a] Archive"
	case TypeAlert:
		return "[r] Reply  [a] Acknowledge"
	case TypeInfo:
		return "[a] Archive"
	default:
		return ""
	}
}

// renderFooter renders the help footer.
func (m Model) renderFooter() string {
	if m.showHelp {
		return m.help.View(m.keys)
	}
	return helpStyle.Render("↑↓ nav | q quit | ? help")
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
