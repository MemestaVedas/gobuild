package screens

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"charm.land/lipgloss/v2"
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
	titleColor := lipgloss.Color("#CBA6F7")
	title := "  HELP & KEYBINDINGS "

	titleRow := lipgloss.NewStyle().Foreground(titleColor).Bold(true).Render(title)

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
		"  :run [profile] Start specific build profile",
		"  :kill [id]     Kill specific build",
		"  :q             Quit app",
		"",
		"Mobile Setup:",
		"  Run ./install-mobile.sh to install the companion app.",
		"  The app will automatically discover GoBuild on your local network.",
	}

	content := lipgloss.NewStyle().Padding(1, 4).Foreground(lipgloss.Color("#CDD6F4")).Render(strings.Join(contentRows, "\n"))

	contentStyle := lipgloss.NewStyle().
		Width(h.width).
		Height(h.height)

	return lipgloss.JoinVertical(lipgloss.Left, titleRow, contentStyle.Render(content))
}

func (h *Help) Focus() {}
func (h *Help) Blur()  {}
