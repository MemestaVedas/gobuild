package screens

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Plugins struct {
	width  int
	height int
	cursor int
	items  []string
}

func NewPlugins() *Plugins {
	return &Plugins{
		items: []string{
			"slack-notify    | ● Active | Post to Slack on build end",
			"discord-notify  | ○ Off    | Post to Discord channel",
			"git-detect      | ● Active | Link builds to git commits",
		},
	}
}

func (p *Plugins) Init() tea.Cmd { return nil }

func (p *Plugins) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		p.width = msg.Width
		p.height = msg.Height - 3
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if p.cursor > 0 {
				p.cursor--
			}
		case "down", "j":
			if p.cursor < len(p.items)-1 {
				p.cursor++
			}
		}
	}
	return p, nil
}

func (p *Plugins) View() string {
	borderColor := lipgloss.Color("#CBA6F7")
	title := "╭▸ PLUGINS "

	border := lipgloss.RoundedBorder()
	titleRow := lipgloss.NewStyle().Foreground(borderColor).Bold(true).Render(title) +
		lipgloss.NewStyle().Foreground(borderColor).Render(strings.Repeat(border.Top, p.width-lipgloss.Width(title)-2)+"╮")

	header := "  Plugin          | Status   | Description"
	var rows []string
	rows = append(rows, lipgloss.NewStyle().Foreground(lipgloss.Color("#6C7086")).Render(header))
	rows = append(rows, strings.Repeat("-", p.width-4))

	for i, item := range p.items {
		cursor := "  "
		if i == p.cursor {
			cursor = "▸ "
		}
		rows = append(rows, lipgloss.NewStyle().Foreground(lipgloss.Color("#CDD6F4")).Render(cursor+item))
	}

	content := strings.Join(rows, "\n")
	contentStyle := lipgloss.NewStyle().
		Width(p.width - 2).
		Height(p.height - 2).
		Border(lipgloss.Border{
			Left: border.Left, Right: border.Right, Bottom: border.Bottom,
			BottomLeft: border.BottomLeft, BottomRight: border.BottomRight,
		}).
		BorderForeground(borderColor)

	return lipgloss.JoinVertical(lipgloss.Left, titleRow, contentStyle.Render(content))
}

func (p *Plugins) Focus() {}
func (p *Plugins) Blur()  {}
