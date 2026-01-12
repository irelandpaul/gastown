package telegram

import (
	"strings"
	"testing"
	"github.com/steveyegge/gastown/internal/mail"
)

func TestFormatMessage(t *testing.T) {
	msg := &mail.Message{
		From:    "gastown/Toast",
		Subject: "Hello World",
		Body:    "This is a test body.",
	}

	formatted := FormatMessage(msg)
	
	if !strings.Contains(formatted, "Hello World") {
		t.Errorf("Expected subject in formatted message, got: %s", formatted)
	}
	if !strings.Contains(formatted, "gastown/Toast") {
		t.Errorf("Expected sender in formatted message, got: %s", formatted)
	}
	// Note: EscapeText escapes dots and other characters
	if !strings.Contains(formatted, "This is a test body") {
		t.Errorf("Expected body in formatted message, got: %s", formatted)
	}
}

func TestFormatMessageMarkdownEscape(t *testing.T) {
	msg := &mail.Message{
		From:    "user",
		Subject: "Special *chars*",
		Body:    "More _chars_ and [links](http://example.com)",
	}

	formatted := FormatMessage(msg)
	
	// Characters should be escaped for MarkdownV2
	if strings.Contains(formatted, "*chars*") {
		t.Error("Expected '*' to be escaped in subject")
	}
	if strings.Contains(formatted, "_chars_") {
		t.Error("Expected '_' to be escaped in body")
	}
}
