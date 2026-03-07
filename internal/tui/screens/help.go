package screens

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Help struct {
	width  int
	height int
}

func NewHelp() *Help {
	return &Help{}
}

func (h *Help) Init() tea.Cmd { return nil }

func (h *Help) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		h.width = msg.Width
		h.height = msg.Height - 3
	}
	return h, nil
}

func (h *Help) View() string {
	borderColor := lipgloss.Color("#CBA6F7")
	title := "╭▸ HELP & KEYBINDINGS "

	border := lipgloss.RoundedBorder()
	titleRow := lipgloss.NewStyle().Foreground(borderColor).Bold(true).Render(title) +
		lipgloss.NewStyle().Foreground(borderColor).Render(strings.Repeat(border.Top, h.width-lipgloss.Width(title)-2)+"╮")

	contentRows := []string{
		"Global Navigation (Normal Mode):",
		"  1, 2, 3, 4, 5  Switch Tabs",
		"  Tab / Shift+Tab Switch Panels",
		"  :              Command Mode",
		"  q              Quit",
		"",
		"Build Controls:",
		"  r  Run Build       x  Kill Build",
		"  o  Open Error      f  Focus Log",
		"",
		"Command Mode (:):",
		"  :run [profile] Start specific build profile",
		"  :kill [id]     Kill specific build",
		"  :q             Quit app",
	}

	content := lipgloss.NewStyle().Padding(1, 4).Foreground(lipgloss.Color("#CDD6F4")).Render(strings.Join(contentRows, "\n"))

	contentStyle := lipgloss.NewStyle().
		Width(h.width - 2).
		Height(h.height - 2).
		Border(lipgloss.Border{
			Left: border.Left, Right: border.Right, Bottom: border.Bottom,
			BottomLeft: border.BottomLeft, BottomRight: border.BottomRight,
		}).
		BorderForeground(borderColor)

	return lipgloss.JoinVertical(lipgloss.Left, titleRow, contentStyle.Render(content))
}

func (h *Help) Focus() {}
func (h *Help) Blur()  {}
