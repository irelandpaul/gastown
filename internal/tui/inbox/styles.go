package inbox

import (
	"github.com/charmbracelet/lipgloss"
)

// Color definitions matching the spec
var (
	// Message type colors
	colorProposal = lipgloss.Color("11") // Yellow
	colorQuestion = lipgloss.Color("14") // Cyan
	colorAlert    = lipgloss.Color("9")  // Red
	colorInfo     = lipgloss.Color("8")  // Gray

	// UI colors
	colorSelected  = lipgloss.Color("236") // Selection background
	colorBorder    = lipgloss.Color("240") // Border color
	colorTitle     = lipgloss.Color("12")  // Title color (blue)
	colorDim       = lipgloss.Color("8")   // Dimmed text
	colorNormal    = lipgloss.Color("15")  // Normal text (white)
	colorUnread    = lipgloss.Color("15")  // Unread indicator
	colorRead      = lipgloss.Color("8")   // Read indicator (dimmed)
	colorHeader    = lipgloss.Color("15")  // Header text
	colorHeaderDim = lipgloss.Color("8")   // Dimmed header parts
)

// Styles for the inbox TUI
var (
	// Title style for "GT INBOX" header
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorTitle)

	// Selected item style
	selectedStyle = lipgloss.NewStyle().
			Background(colorSelected).
			Foreground(colorNormal)

	// Normal message row styles
	messageStyle = lipgloss.NewStyle().
			Foreground(colorNormal)

	dimStyle = lipgloss.NewStyle().
			Foreground(colorDim)

	// Badge styles by type
	proposalBadgeStyle = lipgloss.NewStyle().
				Foreground(colorProposal)

	questionBadgeStyle = lipgloss.NewStyle().
				Foreground(colorQuestion)

	alertBadgeStyle = lipgloss.NewStyle().
			Foreground(colorAlert).
			Bold(true)

	infoBadgeStyle = lipgloss.NewStyle().
			Foreground(colorInfo)

	// Preview pane styles
	previewHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(colorHeader)

	previewLabelStyle = lipgloss.NewStyle().
				Foreground(colorHeaderDim)

	previewBodyStyle = lipgloss.NewStyle().
				Foreground(colorNormal)

	// Border styles
	borderStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(colorBorder)

	// Help footer style
	helpStyle = lipgloss.NewStyle().
			Foreground(colorDim)

	// Error style
	errorStyle = lipgloss.NewStyle().
			Foreground(colorAlert)

	// Separator line for INFO section
	separatorStyle = lipgloss.NewStyle().
			Foreground(colorDim)

	// Unread indicator styles
	unreadStyle = lipgloss.NewStyle().
			Foreground(colorUnread).
			Bold(true)

	readStyle = lipgloss.NewStyle().
			Foreground(colorRead)

	// Priority styles
	priorityUrgentStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Bold(true)
	priorityHighStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("11"))
	priorityNormalStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("12"))
	priorityLowStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
)

// BadgeStyle returns the appropriate style for a message type badge.
func BadgeStyle(t MessageType) lipgloss.Style {
	switch t {
	case TypeProposal:
		return proposalBadgeStyle
	case TypeQuestion:
		return questionBadgeStyle
	case TypeAlert:
		return alertBadgeStyle
	case TypeInfo:
		return infoBadgeStyle
	default:
		return dimStyle
	}
}
