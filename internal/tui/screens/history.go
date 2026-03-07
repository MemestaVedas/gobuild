package screens

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type History struct {
	width  int
	height int
	cursor int
	items  []string
}

func NewHistory() *History {
	return &History{
		items: []string{
			"gobuild        | Cargo  | ✓ OK   | 112s | 0 Errors",
			"frontend-app   | NPM    | ✗ FAIL | 89s  | 3 Errors",
			"backend-api    | Make   | ✓ OK   | 34s  | 0 Errors",
		},
	}
}

func (h *History) Init() tea.Cmd { return nil }

func (h *History) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		h.width = msg.Width
		h.height = msg.Height - 3
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if h.cursor > 0 {
				h.cursor--
			}
		case "down", "j":
			if h.cursor < len(h.items)-1 {
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

	for i, item := range h.items {
		cursor := "  "
		if i == h.cursor {
			cursor = "▸ "
		}
		rows = append(rows, lipgloss.NewStyle().Foreground(lipgloss.Color("#CDD6F4")).Render("  "+cursor+item))
	}

	content := strings.Join(rows, "\n")
	contentStyle := lipgloss.NewStyle().
		Width(h.width).
		Height(h.height)

	return lipgloss.JoinVertical(lipgloss.Left, titleRow, contentStyle.Render(content))
}

func (h *History) Focus() {}
func (h *History) Blur()  {}
