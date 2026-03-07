package screens

import tea "github.com/charmbracelet/bubbletea"

type Dashboard struct{}

func NewDashboard() *Dashboard                               { return &Dashboard{} }
func (d *Dashboard) Init() tea.Cmd                           { return nil }
func (d *Dashboard) Update(msg tea.Msg) (tea.Model, tea.Cmd) { return d, nil }
func (d *Dashboard) View() string                            { return "Dashboard [Stub]" }
func (d *Dashboard) Focus()                                  {}
func (d *Dashboard) Blur()                                   {}
