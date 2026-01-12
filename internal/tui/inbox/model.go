package inbox

import (
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
)

// ViewMode represents the current view mode of the inbox.
type ViewMode int

const (
	// ModeList shows the message list and preview.
	ModeList ViewMode = iota
	// ModeReply shows the reply input area.
	ModeReply
	// ModeThread shows the thread/conversation view.
	ModeThread
	// ModeExpand shows expanded bead details.
	ModeExpand
)

// ExpandedBead holds information about an expanded bead reference.
type ExpandedBead struct {
	ID          string
	Title       string
	Description string
	Status      string
	Type        string
	Priority    int
	Assignee    string
	Labels      []string
	CreatedAt   string
}

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

	// Phase 2: Reply mode
	mode       ViewMode
	replyInput textarea.Model
	replyingTo *Message // Message being replied to

	// Phase 2: Thread view
	threadMessages []Message // Messages in current thread

	// Phase 2: Status message (for confirmations)
	statusMsg string

	// Phase 3: Bead expansion
	expandedBeads []ExpandedBead // Expanded bead details for current message
	expandCursor  int            // Selected bead in expand view

	// Phase 4: Notifications
	lastFetch time.Time
	newCount  int // New messages since last view

	// Phase 5: Pagination
	page int
}

// New creates a new inbox TUI model.
func New(address, workDir string) Model {
	ti := textarea.New()
	ti.Placeholder = "Type your reply..."
	ti.CharLimit = 4000
	ti.SetWidth(60)
	ti.SetHeight(5)

	return Model{
		address:    address,
		workDir:    workDir,
		keys:       DefaultKeyMap(),
		help:       help.New(),
		messages:   make([]Message, 0),
		loading:    true,
		mode:       ModeList,
		replyInput: ti,
	}
}

// Init initializes the model and starts fetching messages.
func (m Model) Init() tea.Cmd {
	return tea.Batch(m.fetchMessages, m.tick())
}

// tickMsg is sent periodically to trigger a refresh.
type tickMsg time.Time

// tick creates a command that sends a tickMsg every 30 seconds.
func (m Model) tick() tea.Cmd {
	return tea.Tick(30*time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// fetchMessagesMsg is the result of fetching messages.
type fetchMessagesMsg struct {
	messages  []Message
	toArchive []string
	err       error
}

// fetchMessages fetches messages from the mailbox.
func (m Model) fetchMessages() tea.Msg {
	messages, toArchive, err := loadMessages(m.address, m.workDir)
	return fetchMessagesMsg{messages: messages, toArchive: toArchive, err: err}
}

// actionResultMsg is the result of an action (approve, reject, archive, reply).
type actionResultMsg struct {
	action  string // "approve", "reject", "archive", "reply"
	success bool
	err     error
}

// threadLoadedMsg is the result of loading a thread.
type threadLoadedMsg struct {
	messages []Message
	err      error
}

// beadsLoadedMsg is the result of loading bead details.
type beadsLoadedMsg struct {
	beads []ExpandedBead
	err   error
}

// Update handles messages and updates the model state.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.help.Width = msg.Width
		// Update textarea width for reply mode
		m.replyInput.SetWidth(m.width - 4)
		return m, nil

	case fetchMessagesMsg:
		m.loading = false
		m.err = msg.err

		// Phase 5: Auto-archive stacked INFO
		var archiveCmds []tea.Cmd
		for _, id := range msg.toArchive {
			archiveCmds = append(archiveCmds, m.doArchiveByID(id))
		}

		// Phase 4: Notification logic
		var notifyCmds []tea.Cmd
		if m.err == nil && !m.lastFetch.IsZero() {
			// Count new messages that weren't in the previous list
			knownIDs := make(map[string]bool)
			for _, msg := range m.messages {
				knownIDs[msg.ID] = true
			}

			newCount := 0
			for _, msg := range msg.messages {
				if !knownIDs[msg.ID] {
					newCount++
					// Check for ALERT
					if msg.Type == TypeAlert {
						notifyCmds = append(notifyCmds, notifyAlert(msg.Subject))
					}
				}
			}
			m.newCount += newCount
		}

		m.messages = msg.messages
		if m.lastFetch.IsZero() {
			m.lastFetch = time.Now()
		}

		cmds := append(archiveCmds, notifyCmds...)
		if len(cmds) > 0 {
			return m, tea.Batch(cmds...)
		}
		return m, nil

	case tickMsg:
		// Periodic refresh
		return m, tea.Batch(m.fetchMessages, m.tick())

	case actionResultMsg:
		if msg.success {
			if msg.action != "Auto-archived" {
				m.statusMsg = msg.action + " successful"
				// Refresh messages after action
				return m, m.fetchMessages
			}
			return m, nil
		}
		if msg.err != nil {
			m.statusMsg = msg.action + " failed: " + msg.err.Error()
		}
		return m, nil

	case threadLoadedMsg:
		if msg.err != nil {
			m.statusMsg = "Failed to load thread: " + msg.err.Error()
			return m, nil
		}
		m.threadMessages = msg.messages
		m.mode = ModeThread
		return m, nil

	case beadsLoadedMsg:
		if msg.err != nil {
			m.statusMsg = "Failed to load beads: " + msg.err.Error()
			return m, nil
		}
		m.expandedBeads = msg.beads
		m.mode = ModeExpand
		return m, nil

	case tea.KeyMsg:
		// Clear status message and new count on any key press
		m.statusMsg = ""
		m.newCount = 0

		// Handle mode-specific input
		switch m.mode {
		case ModeReply:
			return m.updateReplyMode(msg)
		case ModeThread:
			return m.updateThreadMode(msg)
		case ModeExpand:
			return m.updateExpandMode(msg)
		default:
			return m.updateListMode(msg)
		}
	}

	return m, nil
}

// updateListMode handles key input in list mode.
func (m Model) updateListMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
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

	case key.Matches(msg, m.keys.Approve):
		// y - approve (only for proposals)
		if sel := m.SelectedMessage(); sel != nil && sel.Type == TypeProposal {
			return m, m.doApprove(sel)
		}
		return m, nil

	case key.Matches(msg, m.keys.Reject):
		// n - reject (only for proposals)
		if sel := m.SelectedMessage(); sel != nil && sel.Type == TypeProposal {
			return m, m.doReject(sel)
		}
		return m, nil

	case key.Matches(msg, m.keys.Reply):
		// r - enter reply mode
		if sel := m.SelectedMessage(); sel != nil {
			m.mode = ModeReply
			m.replyingTo = sel
			m.replyInput.Reset()
			m.replyInput.Focus()
			return m, nil
		}
		return m, nil

	case key.Matches(msg, m.keys.Archive):
		// a - archive message
		if sel := m.SelectedMessage(); sel != nil {
			return m, m.doArchive(sel)
		}
		return m, nil

	case key.Matches(msg, m.keys.ArchiveInfo):
		// A - archive all INFO
		return m, m.doArchiveInfo()

	case key.Matches(msg, m.keys.MarkAllRead):
		// M - mark all read
		return m, m.doMarkAllRead()

	case key.Matches(msg, m.keys.ArchiveOld):
		// D - archive old
		return m, m.doArchiveOld()

	case key.Matches(msg, m.keys.NextPage):
		// ] - next page
		if (m.page+1)*100 < len(m.messages) {
			m.page++
			m.cursor = m.page * 100
		}
		return m, nil

	case key.Matches(msg, m.keys.PrevPage):
		// [ - prev page
		if m.page > 0 {
			m.page--
			m.cursor = m.page * 100
		}
		return m, nil

	case key.Matches(msg, m.keys.Tab):
		// tab - show thread view if message has thread
		if sel := m.SelectedMessage(); sel != nil && sel.ThreadID != "" {
			return m, m.loadThread(sel.ThreadID)
		}
		return m, nil

	case key.Matches(msg, m.keys.Expand):
		// e - expand bead references
		if sel := m.SelectedMessage(); sel != nil && len(sel.References) > 0 {
			return m, m.loadBeads(sel.References)
		}
		return m, nil
	}

	return m, nil
}

// updateExpandMode handles key input in expand mode.
func (m Model) updateExpandMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Quit), msg.Type == tea.KeyEsc:
		// Exit expand view back to list
		m.mode = ModeList
		m.expandedBeads = nil
		m.expandCursor = 0
		return m, nil

	case key.Matches(msg, m.keys.Up):
		if m.expandCursor > 0 {
			m.expandCursor--
		}
		return m, nil

	case key.Matches(msg, m.keys.Down):
		if m.expandCursor < len(m.expandedBeads)-1 {
			m.expandCursor++
		}
		return m, nil

	case key.Matches(msg, m.keys.Hook):
		if m.expandCursor >= 0 && m.expandCursor < len(m.expandedBeads) {
			bead := m.expandedBeads[m.expandCursor]
			if bead.Status == "open" {
				return m, m.doHook(bead.ID)
			}
		}
		return m, nil
	}

	return m, nil
}

// updateReplyMode handles key input in reply mode.
func (m Model) updateReplyMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		// Cancel reply
		m.mode = ModeList
		m.replyingTo = nil
		m.replyInput.Blur()
		return m, nil

	case tea.KeyCtrlD:
		// Send reply (Ctrl+D as alternative to Enter since Enter adds newlines)
		if m.replyingTo != nil && m.replyInput.Value() != "" {
			cmd := m.doReply(m.replyingTo, m.replyInput.Value())
			m.mode = ModeList
			m.replyingTo = nil
			m.replyInput.Blur()
			return m, cmd
		}
		return m, nil
	}

	// Pass to textarea
	var cmd tea.Cmd
	m.replyInput, cmd = m.replyInput.Update(msg)
	return m, cmd
}

// updateThreadMode handles key input in thread mode.
func (m Model) updateThreadMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Quit), msg.Type == tea.KeyEsc:
		// Exit thread view back to list
		m.mode = ModeList
		m.threadMessages = nil
		return m, nil

	case key.Matches(msg, m.keys.Reply):
		// r - reply to thread (reply to original message)
		if len(m.threadMessages) > 0 {
			// Reply to the first message in thread (the original)
			original := m.threadMessages[0]
			m.mode = ModeReply
			m.replyingTo = &original
			m.replyInput.Reset()
			m.replyInput.Focus()
		}
		return m, nil
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

// doApprove creates a command to approve a proposal.
func (m Model) doApprove(msg *Message) tea.Cmd {
	return func() tea.Msg {
		err := approveMessage(msg.ID, m.address, m.workDir)
		return actionResultMsg{
			action:  "Approved",
			success: err == nil,
			err:     err,
		}
	}
}

// doReject creates a command to reject a proposal.
func (m Model) doReject(msg *Message) tea.Cmd {
	return func() tea.Msg {
		err := rejectMessage(msg.ID, m.address, m.workDir)
		return actionResultMsg{
			action:  "Rejected",
			success: err == nil,
			err:     err,
		}
	}
}

// doArchive creates a command to archive a message.
func (m Model) doArchive(msg *Message) tea.Cmd {
	return m.doArchiveByID(msg.ID)
}

// doArchiveByID creates a command to archive a message by its ID.
func (m Model) doArchiveByID(id string) tea.Cmd {
	return func() tea.Msg {
		err := archiveMessage(id, m.address, m.workDir)
		return actionResultMsg{
			action:  "Auto-archived",
			success: err == nil,
			err:     err,
		}
	}
}

// doArchiveInfo creates a command to archive all INFO messages.
func (m Model) doArchiveInfo() tea.Cmd {
	return func() tea.Msg {
		err := archiveInfo(m.address, m.workDir)
		return actionResultMsg{
			action:  "Archived all info",
			success: err == nil,
			err:     err,
		}
	}
}

// doMarkAllRead creates a command to mark all messages as read.
func (m Model) doMarkAllRead() tea.Cmd {
	return func() tea.Msg {
		err := markAllRead(m.address, m.workDir)
		return actionResultMsg{
			action:  "Marked all read",
			success: err == nil,
			err:     err,
		}
	}
}

// doArchiveOld creates a command to archive old messages.
func (m Model) doArchiveOld() tea.Cmd {
	return func() tea.Msg {
		err := archiveOld(m.address, m.workDir)
		return actionResultMsg{
			action:  "Archived old messages",
			success: err == nil,
			err:     err,
		}
	}
}

// doReply creates a command to send a reply.
func (m Model) doReply(msg *Message, body string) tea.Cmd {
	return func() tea.Msg {
		err := sendReply(msg, body, m.address, m.workDir)
		return actionResultMsg{
			action:  "Reply sent",
			success: err == nil,
			err:     err,
		}
	}
}

// loadThread creates a command to load thread messages.
func (m Model) loadThread(threadID string) tea.Cmd {
	return func() tea.Msg {
		messages, err := loadThreadMessages(threadID, m.address, m.workDir)
		// Convert mail.Message to inbox.Message
		var inboxMsgs []Message
		for _, mm := range messages {
			inboxMsgs = append(inboxMsgs, convertToInboxMessage(mm))
		}
		return threadLoadedMsg{
			messages: inboxMsgs,
			err:      err,
		}
	}
}

// loadBeads creates a command to load bead details.
func (m Model) loadBeads(beadIDs []string) tea.Cmd {
	return func() tea.Msg {
		beads, err := fetchBeadDetails(beadIDs, m.workDir)
		return beadsLoadedMsg{
			beads: beads,
			err:   err,
		}
	}
}

// doHook creates a command to hook a bead.
func (m Model) doHook(beadID string) tea.Cmd {
	return func() tea.Msg {
		err := hookBead(beadID, m.address, m.workDir)
		return actionResultMsg{
			action:  "Hooked " + beadID,
			success: err == nil,
			err:     err,
		}
	}
}
