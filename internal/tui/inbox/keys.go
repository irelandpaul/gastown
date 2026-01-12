package inbox

import "github.com/charmbracelet/bubbles/key"

// KeyMap defines the key bindings for the inbox TUI.
type KeyMap struct {
	// Navigation
	Up       key.Binding
	Down     key.Binding
	PageUp   key.Binding
	PageDown key.Binding
	Top      key.Binding
	Bottom   key.Binding

	// Actions
	Approve     key.Binding
	Reject      key.Binding
	Reply       key.Binding
	Archive     key.Binding
	ArchiveInfo key.Binding // Phase 5: Archive all INFO messages
	MarkAllRead key.Binding // Phase 5: Mark all messages as read
	ArchiveOld  key.Binding // Phase 5: Archive old messages
	Expand      key.Binding // Phase 3: Expand bead references
	Hook        key.Binding // Phase 3: Hook/claim bead
	Learn       key.Binding // Phase 6: Learn message type

	// General
	NextPage key.Binding // Phase 5: Next page of messages
	PrevPage key.Binding // Phase 5: Previous page of messages
	Tab      key.Binding
	Help     key.Binding
	Quit     key.Binding
}

// DefaultKeyMap returns the default key bindings for the inbox TUI.
func DefaultKeyMap() KeyMap {
	return KeyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "down"),
		),
		PageUp: key.NewBinding(
			key.WithKeys("pgup", "ctrl+u"),
			key.WithHelp("pgup", "page up"),
		),
		PageDown: key.NewBinding(
			key.WithKeys("pgdown", "ctrl+d"),
			key.WithHelp("pgdn", "page down"),
		),
		Top: key.NewBinding(
			key.WithKeys("home", "g"),
			key.WithHelp("g", "top"),
		),
		Bottom: key.NewBinding(
			key.WithKeys("end", "G"),
			key.WithHelp("G", "bottom"),
		),
		Approve: key.NewBinding(
			key.WithKeys("y"),
			key.WithHelp("y", "approve [P]"),
		),
		Reject: key.NewBinding(
			key.WithKeys("n"),
			key.WithHelp("n", "reject [P]"),
		),
		Reply: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "reply"),
		),
		Archive: key.NewBinding(
			key.WithKeys("a"),
			key.WithHelp("a", "archive"),
		),
		ArchiveInfo: key.NewBinding(
			key.WithKeys("A"),
			key.WithHelp("A", "archive all info"),
		),
		MarkAllRead: key.NewBinding(
			key.WithKeys("M"),
			key.WithHelp("M", "mark all read"),
		),
		ArchiveOld: key.NewBinding(
			key.WithKeys("D"),
			key.WithHelp("D", "archive old"),
		),
		NextPage: key.NewBinding(
			key.WithKeys("]"),
			key.WithHelp("]", "next page"),
		),
		PrevPage: key.NewBinding(
			key.WithKeys("["),
			key.WithHelp("[", "prev page"),
		),
		Expand: key.NewBinding(
			key.WithKeys("e"),
			key.WithHelp("e", "expand beads"),
		),
		Hook: key.NewBinding(
			key.WithKeys("h"),
			key.WithHelp("h", "hook bead"),
		),
		Learn: key.NewBinding(
			key.WithKeys("L"),
			key.WithHelp("L", "learn type"),
		),
		Tab: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "switch pane"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "help"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "esc", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
	}
}

// ShortHelp returns keybindings to show in the mini help view.
func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Archive, k.Quit, k.Help}
}

// FullHelp returns keybindings for the expanded help view.
func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.PageUp, k.PageDown},
		{k.Top, k.Bottom, k.NextPage, k.PrevPage, k.Tab},
		{k.Approve, k.Reject, k.Reply, k.Archive},
		{k.ArchiveInfo, k.MarkAllRead, k.ArchiveOld},
		{k.Expand, k.Hook, k.Learn},
		{k.Help, k.Quit},
	}
}