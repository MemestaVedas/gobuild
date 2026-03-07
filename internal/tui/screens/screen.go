package screens

import (
	tea "github.com/charmbracelet/bubbletea"
)

// Screen represents a single tab in the TUI (Dashboard, Launcher, etc.).
type Screen interface {
	Init() tea.Cmd
	Update(msg tea.Msg) (tea.Model, tea.Cmd)
	View() string
	Focus()
	Blur()
}
