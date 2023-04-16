package table

import "github.com/charmbracelet/bubbles/key"

// KeyMap defines keybindings. It satisfies to the help.KeyMap interface, which
// is used to render the menu menu.
type KeyMap struct {
	// Keybindings used when browsing the list.
	NextPage      key.Binding
	PrevPage      key.Binding
	GoToStart     key.Binding
	GoToEnd       key.Binding
	Filter        key.Binding
	ClearFilter   key.Binding
	ToggleDeleted key.Binding

	// Keybindings used when setting a filter.
	CancelWhileFiltering key.Binding
	AcceptWhileFiltering key.Binding

	// Help toggle keybindings.
	ShowFullHelp  key.Binding
	CloseFullHelp key.Binding

	// The quit-no-matter-what keybinding. This will be caught when filtering.
	ForceQuit key.Binding
}

// DefaultKeyMap is a default set of keybindings.
var DefaultKeyMap = KeyMap{
	// Browsing.
	PrevPage: key.NewBinding(
		key.WithKeys("left", "h", "pgup", "b"),
		key.WithHelp("←/h/pgup", "prev page"),
	),
	NextPage: key.NewBinding(
		key.WithKeys("right", "l", "pgdown", "f"),
		key.WithHelp("→/l/pgdn", "next page"),
	),
	GoToStart: key.NewBinding(
		key.WithKeys("home", "g"),
		key.WithHelp("g/home", "go to start"),
	),
	GoToEnd: key.NewBinding(
		key.WithKeys("end", "G"),
		key.WithHelp("G/end", "go to end"),
	),
	Filter: key.NewBinding(
		key.WithKeys("/"),
		key.WithHelp("/", "filter"),
	),
	ClearFilter: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "clear filter"),
	),
	ToggleDeleted: key.NewBinding(
		key.WithKeys("d"),
		key.WithHelp("d", "show/hide deleted"),
	),

	// Filtering.
	CancelWhileFiltering: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "cancel"),
	),
	AcceptWhileFiltering: key.NewBinding(
		key.WithKeys("enter", "tab", "shift+tab", "ctrl+k", "up", "ctrl+j", "down"),
		key.WithHelp("enter", "apply filter"),
	),

	// Toggle help.
	ShowFullHelp: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "more"),
	),
	CloseFullHelp: key.NewBinding(
		key.WithKeys("?", "esc"),
		key.WithHelp("?/esc", "close help"),
	),

	// Quitting.
	ForceQuit: key.NewBinding(key.WithKeys("ctrl+c")),
}
