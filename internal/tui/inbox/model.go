package inbox

import (
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// Model is the bubbletea model for the inbox TUI.
type Model struct {
	// Messages is the list of messages to display.
	messages []Message

	// cursor is the currently selected message index.
	cursor int

	// address is the inbox address being viewed.
	address string

	// workDir is the beads working directory.
	workDir string

	// UI state
	keys     KeyMap
	help     help.Model
	showHelp bool
	width    int
	height   int

	// Error state
	err error

	// Loading state
	loading bool
}

// New creates a new inbox TUI model.
func New(address, workDir string) Model {
	return Model{
		address:  address,
		workDir:  workDir,
		keys:     DefaultKeyMap(),
		help:     help.New(),
		messages: make([]Message, 0),
		loading:  true,
	}
}

// Init initializes the model and starts fetching messages.
func (m Model) Init() tea.Cmd {
	return m.fetchMessages
}

// fetchMessagesMsg is the result of fetching messages.
type fetchMessagesMsg struct {
	messages []Message
	err      error
}

// fetchMessages fetches messages from the mailbox.
func (m Model) fetchMessages() tea.Msg {
	messages, err := loadMessages(m.address, m.workDir)
	return fetchMessagesMsg{messages: messages, err: err}
}

// Update handles messages and updates the model state.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.help.Width = msg.Width
		return m, nil

	case fetchMessagesMsg:
		m.loading = false
		m.err = msg.err
		m.messages = msg.messages
		return m, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit

		case key.Matches(msg, m.keys.Help):
			m.showHelp = !m.showHelp
			return m, nil

		case key.Matches(msg, m.keys.Up):
			if m.cursor > 0 {
				m.cursor--
			}
			return m, nil

		case key.Matches(msg, m.keys.Down):
			if m.cursor < len(m.messages)-1 {
				m.cursor++
			}
			return m, nil

		case key.Matches(msg, m.keys.Top):
			m.cursor = 0
			return m, nil

		case key.Matches(msg, m.keys.Bottom):
			if len(m.messages) > 0 {
				m.cursor = len(m.messages) - 1
			}
			return m, nil

		case key.Matches(msg, m.keys.PageUp):
			m.cursor -= 10
			if m.cursor < 0 {
				m.cursor = 0
			}
			return m, nil

		case key.Matches(msg, m.keys.PageDown):
			m.cursor += 10
			if m.cursor >= len(m.messages) {
				m.cursor = len(m.messages) - 1
			}
			if m.cursor < 0 {
				m.cursor = 0
			}
			return m, nil
		}
	}

	return m, nil
}

// View renders the model to a string.
func (m Model) View() string {
	return m.renderView()
}

// SelectedMessage returns the currently selected message, or nil if none.
func (m Model) SelectedMessage() *Message {
	if m.cursor >= 0 && m.cursor < len(m.messages) {
		return &m.messages[m.cursor]
	}
	return nil
}
