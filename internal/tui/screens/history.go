package screens

import tea "github.com/charmbracelet/bubbletea"

type History struct{}

func NewHistory() *History                                 { return &History{} }
func (h *History) Init() tea.Cmd                           { return nil }
func (h *History) Update(msg tea.Msg) (tea.Model, tea.Cmd) { return h, nil }
func (h *History) View() string                            { return "History [Stub]" }
func (h *History) Focus()                                  {}
func (h *History) Blur()                                   {}
