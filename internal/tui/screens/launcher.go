package screens

import tea "github.com/charmbracelet/bubbletea"

type Launcher struct{}

func NewLauncher() *Launcher                                { return &Launcher{} }
func (l *Launcher) Init() tea.Cmd                           { return nil }
func (l *Launcher) Update(msg tea.Msg) (tea.Model, tea.Cmd) { return l, nil }
func (l *Launcher) View() string                            { return "Launcher [Stub]" }
func (l *Launcher) Focus()                                  {}
func (l *Launcher) Blur()                                   {}
