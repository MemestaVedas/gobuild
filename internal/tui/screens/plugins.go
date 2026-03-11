package screens

import (
	"strings"

	"github.com/MemestaVedas/gobuild/internal/tui/theme"
	tea "github.com/charmbracelet/bubbletea"
	"charm.land/lipgloss/v2"
)

type Plugins struct {
	width  int
	height int
	cursor int
	items  []string
	styles theme.Styles
}

func NewPlugins(styles theme.Styles) *Plugins {
	return &Plugins{
		styles: styles,
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
		p.height = msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if p.cursor > 0 { p.cursor-- }
		case "down", "j":
			if p.cursor < len(p.items)-1 { p.cursor++ }
		}
	}
	return p, nil
}

func (p *Plugins) View() string {
	if p.width <= 0 || p.height <= 0 { return "" }
	
	header := "  Plugin          | Status   | Description"
	var rows []string
	rows = append(rows, lipgloss.NewStyle().Foreground(p.styles.ColorFaint).Render(header))
	rows = append(rows, lipgloss.NewStyle().Foreground(p.styles.ColorBorderDim).Render(strings.Repeat("─", p.width-4)))

	for i, item := range p.items {
		cursor := "  "
		c := p.styles.ColorText
		if i == p.cursor {
			cursor = "▸ "
			c = p.styles.ColorAccent
		}
		rows = append(rows, lipgloss.NewStyle().Foreground(c).Render("  "+cursor+item))
	}

	content := strings.Join(rows, "\n")
	return p.panel("Available Plugins", p.width, p.height, content)
}

func (p *Plugins) panel(title string, w, h int, content string) string {
	bColor := p.styles.ColorBorderInactive
	titleStyled := lipgloss.NewStyle().Foreground(p.styles.ColorText).Render(" " + title + " ")
	
	titleBarWidth := w - 2
	if titleBarWidth < 0 { titleBarWidth = 0 }

	dashCount := titleBarWidth - lipgloss.Width(titleStyled) + 1
	if dashCount < 0 { dashCount = 0 }

	topLine := lipgloss.NewStyle().Foreground(bColor).Render("╭") +
		titleStyled +
		lipgloss.NewStyle().Foreground(bColor).Render(strings.Repeat("─", dashCount)) +
		lipgloss.NewStyle().Foreground(bColor).Render("╮")

	box := lipgloss.NewStyle().
		Width(w - 2).Height(h - 2).
		BorderLeft(true).BorderBottom(true).BorderRight(true).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(bColor).
		Render(content)

	return lipgloss.JoinVertical(lipgloss.Left, topLine, box)
}

func (p *Plugins) Focus() {}
func (p *Plugins) Blur()  {}
