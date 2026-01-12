package inbox

import (
	"os/exec"

	tea "github.com/charmbracelet/bubbletea"
)

// notifyAlert creates a command that sends a desktop notification.
func notifyAlert(subject string) tea.Cmd {
	return func() tea.Msg {
		// Ignore errors, it's just a notification
		_ = exec.Command("notify-send", "-u", "critical", "GT Alert", subject).Run()
		return nil
	}
}