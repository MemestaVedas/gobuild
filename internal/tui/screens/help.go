package screens

import tea "github.com/charmbracelet/bubbletea"

type Help struct{}

func NewHelp() *Help                                    { return &Help{} }
func (h *Help) Init() tea.Cmd                           { return nil }
func (h *Help) Update(msg tea.Msg) (tea.Model, tea.Cmd) { return h, nil }
func (h *Help) View() string                            { return "Help [Stub]" }
func (h *Help) Focus()                                  {}
func (h *Help) Blur()                                   {}
