package screens

import (
	"strings"

	"github.com/MemestaVedas/gobuild/internal/tui/theme"
	tea "github.com/charmbracelet/bubbletea"
	"charm.land/lipgloss/v2"
)

type Help struct {
	width  int
	height int
	styles theme.Styles
}

func NewHelp(styles theme.Styles) *Help {
	return &Help{styles: styles}
}

func (h *Help) Init() tea.Cmd { return nil }

func (h *Help) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		h.width = msg.Width
		h.height = msg.Height
	}
	return h, nil
}

func (h *Help) View() string {
	if h.width <= 0 || h.height <= 0 { return "" }
	
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

	content := lipgloss.NewStyle().Padding(1, 2).Foreground(h.styles.ColorText).Render(strings.Join(contentRows, "\n"))
	return h.panel("Help & Keybindings", h.width, h.height, content)
}

func (h *Help) panel(title string, w, hP int, content string) string {
	bColor := h.styles.ColorBorderInactive
	titleStyled := lipgloss.NewStyle().Foreground(h.styles.ColorText).Render(" " + title + " ")
	
	titleBarWidth := w - 2
	if titleBarWidth < 0 { titleBarWidth = 0 }

	dashCount := titleBarWidth - lipgloss.Width(titleStyled) + 1
	if dashCount < 0 { dashCount = 0 }

	topLine := lipgloss.NewStyle().Foreground(bColor).Render("╭") +
		titleStyled +
		lipgloss.NewStyle().Foreground(bColor).Render(strings.Repeat("─", dashCount)) +
		lipgloss.NewStyle().Foreground(bColor).Render("╮")

	box := lipgloss.NewStyle().
		Width(w - 2).Height(hP - 2).
		BorderLeft(true).BorderBottom(true).BorderRight(true).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(bColor).
		Render(content)

	return lipgloss.JoinVertical(lipgloss.Left, topLine, box)
}

func (h *Help) Focus() {}
func (h *Help) Blur()  {}
