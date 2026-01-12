package telegram

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/steveyegge/gastown/internal/mail"
	"github.com/steveyegge/gastown/internal/tui/inbox"
)

// FormatMessage formats a Gas Town message for Telegram.
func FormatMessage(msg *mail.Message) string {
	msgType := inbox.InferMessageType(msg)
	badge := msgType.Badge()

	text := fmt.Sprintf("%s *%s*\nFrom: %s\n\n%s",
		badge,
		tgbotapi.EscapeText(tgbotapi.ModeMarkdownV2, msg.Subject),
		tgbotapi.EscapeText(tgbotapi.ModeMarkdownV2, msg.From),
		tgbotapi.EscapeText(tgbotapi.ModeMarkdownV2, msg.Body),
	)

	// Trim if too long for Telegram (4096 chars)
	if len(text) > 4000 {
		text = text[:3997] + "..."
	}
	
	return text
}

