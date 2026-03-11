package screens

import (
	"fmt"
	"strings"

	"github.com/MemestaVedas/gobuild/internal/core"
	"github.com/MemestaVedas/gobuild/internal/tui/theme"
	tea "github.com/charmbracelet/bubbletea"
	"charm.land/lipgloss/v2"
)

type History struct {
	width  int
	height int
	bm     *core.BuildManager
	cursor int
	styles theme.Styles
}

func NewHistory(bm *core.BuildManager, styles theme.Styles) *History {
	return &History{bm: bm, styles: styles}
}

func (h *History) Init() tea.Cmd { return nil }

func (h *History) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		h.width = msg.Width
		h.height = msg.Height
	case tea.KeyMsg:
		all := h.bm.All()
		switch msg.String() {
		case "up", "k":
			if h.cursor > 0 { h.cursor-- }
		case "down", "j":
			if h.cursor < len(all)-1 { h.cursor++ }
		}
	}
	return h, nil
}

func (h *History) View() string {
	if h.width <= 0 || h.height <= 0 { return "" }
	
	header := "  Project        | Tool   | Status | Time | Errors"
	var rows []string
	rows = append(rows, lipgloss.NewStyle().Foreground(h.styles.ColorFaint).Render(header))
	rows = append(rows, lipgloss.NewStyle().Foreground(h.styles.ColorBorderDim).Render(strings.Repeat("─", h.width-4)))

	all := h.bm.All()
	for i, b := range all {
		item := fmt.Sprintf("%-15s | %-6s | %-6s | %4ds | %d Errors",
			truncate(b.Name, 15), b.Tool.String(), b.State.String(), int(b.Elapsed().Seconds()), len(b.Errors))

		cursorStr := "  "
		c := h.styles.ColorText
		if i == h.cursor {
			cursorStr = "▸ "
			c = h.styles.ColorAccent
		}
		rows = append(rows, lipgloss.NewStyle().Foreground(c).Render("  "+cursorStr+item))
	}

	content := strings.Join(rows, "\n")
	
	return h.panel("Build History", h.width, h.height, content)
}

func (h *History) panel(title string, w, hP int, content string) string {
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

func (h *History) Focus() {}
func (h *History) Blur()  {}
