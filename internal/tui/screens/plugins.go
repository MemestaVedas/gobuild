package screens

import tea "github.com/charmbracelet/bubbletea"

type Plugins struct{}

func NewPlugins() *Plugins                                 { return &Plugins{} }
func (p *Plugins) Init() tea.Cmd                           { return nil }
func (p *Plugins) Update(msg tea.Msg) (tea.Model, tea.Cmd) { return p, nil }
func (p *Plugins) View() string                            { return "Plugins [Stub]" }
func (p *Plugins) Focus()                                  {}
func (p *Plugins) Blur()                                   {}
