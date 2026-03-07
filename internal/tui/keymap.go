package tui

import "github.com/charmbracelet/bubbles/key"

// KeyMap defines the global keybindings.
type KeyMap struct {
	// Mode transitions
	Esc     key.Binding
	Enter   key.Binding
	Command key.Binding

	// Tier 1 - Global Navigation (Normal)
	Tab1      key.Binding
	Tab2      key.Binding
	Tab3      key.Binding
	Tab4      key.Binding
	Tab5      key.Binding
	NextPanel key.Binding
	PrevPanel key.Binding
	Left      key.Binding
	Right     key.Binding
	Down      key.Binding
	Up        key.Binding
	Quit      key.Binding
	Help      key.Binding

	// Tier 2 - Build Control (Dashboard & Launcher)
	Run      key.Binding
	Kill     key.Binding
	FocusLog key.Binding
	OpenErr  key.Binding
	Details  key.Binding
	Tag      key.Binding

	// Tier 3 - Log Navigation (Log Panel)
	ScrollDown key.Binding
	ScrollUp   key.Binding
	HalfDown   key.Binding
	HalfUp     key.Binding
	Top        key.Binding
	Bottom     key.Binding
	Search     key.Binding
	NextMatch  key.Binding
	PrevMatch  key.Binding

	// Launcher / History
	Delete key.Binding
}

// DefaultKeyMap returns the specification-compliant default bindings.
func DefaultKeyMap() KeyMap {
	return KeyMap{
		Esc:     key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "normal mode")),
		Enter:   key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "select / insert")),
		Command: key.NewBinding(key.WithKeys(":"), key.WithHelp(":", "command mode")),

		Tab1:      key.NewBinding(key.WithKeys("1"), key.WithHelp("1", "dashboard")),
		Tab2:      key.NewBinding(key.WithKeys("2"), key.WithHelp("2", "launcher")),
		Tab3:      key.NewBinding(key.WithKeys("3"), key.WithHelp("3", "history")),
		Tab4:      key.NewBinding(key.WithKeys("4"), key.WithHelp("4", "plugins")),
		Tab5:      key.NewBinding(key.WithKeys("5", "?"), key.WithHelp("?", "help")),
		NextPanel: key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "next panel")),
		PrevPanel: key.NewBinding(key.WithKeys("shift+tab"), key.WithHelp("shift+tab", "prev panel")),
		Left:      key.NewBinding(key.WithKeys("ctrl+h"), key.WithHelp("ctrl+h", "left")),
		Right:     key.NewBinding(key.WithKeys("ctrl+l"), key.WithHelp("ctrl+l", "right")),
		Down:      key.NewBinding(key.WithKeys("ctrl+j"), key.WithHelp("ctrl+j", "down")),
		Up:        key.NewBinding(key.WithKeys("ctrl+k"), key.WithHelp("ctrl+k", "up")),
		Quit:      key.NewBinding(key.WithKeys("q"), key.WithHelp("q", "quit")),
		Help:      key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "help")),

		Run:      key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "run build")),
		Kill:     key.NewBinding(key.WithKeys("x"), key.WithHelp("x", "kill build")),
		FocusLog: key.NewBinding(key.WithKeys("f"), key.WithHelp("f", "focus log")),
		OpenErr:  key.NewBinding(key.WithKeys("o"), key.WithHelp("o", "open error")),
		Details:  key.NewBinding(key.WithKeys("d"), key.WithHelp("d", "details popup")),
		Tag:      key.NewBinding(key.WithKeys("t"), key.WithHelp("t", "tag build")),

		ScrollDown: key.NewBinding(key.WithKeys("j", "down"), key.WithHelp("j/↓", "down")),
		ScrollUp:   key.NewBinding(key.WithKeys("k", "up"), key.WithHelp("k/↑", "up")),
		HalfDown:   key.NewBinding(key.WithKeys("ctrl+d"), key.WithHelp("ctrl+d", "half down")),
		HalfUp:     key.NewBinding(key.WithKeys("ctrl+u"), key.WithHelp("ctrl+u", "half up")),
		Top:        key.NewBinding(key.WithKeys("g"), key.WithHelp("g", "top")),
		Bottom:     key.NewBinding(key.WithKeys("G"), key.WithHelp("G", "bottom(follow)")),
		Search:     key.NewBinding(key.WithKeys("/"), key.WithHelp("/", "search")),
		NextMatch:  key.NewBinding(key.WithKeys("n"), key.WithHelp("n", "next match")),
		PrevMatch:  key.NewBinding(key.WithKeys("N"), key.WithHelp("N", "prev match")),

		Delete: key.NewBinding(key.WithKeys("delete"), key.WithHelp("del", "delete")),
	}
}
