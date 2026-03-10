package screens

import (
	"fmt"
	"strings"

	"github.com/MemestaVedas/gobuild/internal/core"
	tea "github.com/charmbracelet/bubbletea"
	"charm.land/lipgloss/v2"
)

type History struct {
	width  int
	height int
	bm     *core.BuildManager
	cursor int
}

func NewHistory(bm *core.BuildManager) *History {
	return &History{bm: bm}
}

func (h *History) Init() tea.Cmd { return nil }

func (h *History) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		h.width = msg.Width
		h.height = msg.Height - 3
	case tea.KeyMsg:
		all := h.bm.All()
		switch msg.String() {
		case "up", "k":
			if h.cursor > 0 {
				h.cursor--
			}
		case "down", "j":
			if h.cursor < len(all)-1 {
				h.cursor++
			}
		}
	}
	return h, nil
}

func (h *History) View() string {
	titleColor := lipgloss.Color("#CBA6F7")
	title := "  BUILD HISTORY (Last 50) "

	titleRow := lipgloss.NewStyle().Foreground(titleColor).Bold(true).Render(title)

	header := "  Project        | Tool   | Status | Time | Errors"
	var rows []string
	rows = append(rows, "")
	rows = append(rows, lipgloss.NewStyle().Foreground(lipgloss.Color("#6C7086")).Render(header))
	rows = append(rows, "  "+safeRepeat("-", h.width-4))

	all := h.bm.All()
	for i, b := range all {
		item := fmt.Sprintf("%-15s | %-6s | %-6s | %4ds | %d Errors",
			b.Name, b.Tool.String(), b.State.String(), int(b.Elapsed().Seconds()), len(b.Errors))

		cursorStr := "  "
		if i == h.cursor {
			cursorStr = "▸ "
		}
		rows = append(rows, lipgloss.NewStyle().Foreground(lipgloss.Color("#CDD6F4")).Render("  "+cursorStr+item))
	}

	content := strings.Join(rows, "\n")
	contentStyle := lipgloss.NewStyle().
		Width(h.width).
		Height(h.height)

	return lipgloss.JoinVertical(lipgloss.Left, titleRow, contentStyle.Render(content))
}

func (h *History) Focus() {}
func (h *History) Blur()  {}
