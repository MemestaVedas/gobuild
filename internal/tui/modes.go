package tui

// Mode defines the current global interaction state, following Neovim conventions.
type Mode int

const (
	ModeNormal  Mode = iota // Navigate, scroll, select
	ModeInsert              // Text entry in input fields
	ModeCommand             // Colon-prefix commands at the bottom bar
)

func (m Mode) String() string {
	switch m {
	case ModeInsert:
		return "INSERT"
	case ModeCommand:
		return "COMMAND"
	default:
		return "NORMAL"
	}
}
