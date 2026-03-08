package screens

import (
	"fmt"
	"strings"

	"github.com/MemestaVedas/gobuild/internal/core"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Dashboard struct {
	width  int
	height int
	active int // 0: Active, 1: Flamechart, 2: Log, 3: Errors, 4: Stats
	bm     *core.BuildManager
}

func NewDashboard(bm *core.BuildManager) *Dashboard {
	return &Dashboard{bm: bm}
}

func (d *Dashboard) Init() tea.Cmd { return nil }

func (d *Dashboard) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		d.width = msg.Width
		// Reserve 3 lines for tabs, statusBar and hintbar
		d.height = msg.Height - 3
	case tea.KeyMsg:
		switch msg.String() {
		case "tab":
			d.active = (d.active + 1) % 5
		case "shift+tab":
			d.active = (d.active - 1)
			if d.active < 0 {
				d.active = 4
			}
		case "ctrl+h", "h": // left
			if d.active == 1 || d.active == 2 || d.active == 4 {
				d.active = 0
			}
		case "ctrl+l", "l": // right
			if d.active == 0 || d.active == 3 {
				d.active = 2
			}
		case "ctrl+k", "k": // up
			switch d.active {
			case 3:
				d.active = 0
			case 2:
				d.active = 1
			case 4:
				d.active = 2
			}
		case "ctrl+j", "j": // down
			switch d.active {
			case 0:
				d.active = 3
			case 1:
				d.active = 2
			case 2:
				d.active = 4
			}
		}
	}
	return d, nil
}

func (d *Dashboard) View() string {
	if d.width <= 0 || d.height <= 0 {
		return ""
	}

	leftWidth := 45
	rightWidth := d.width - leftWidth

	activeCount := len(d.bm.Active())
	// Build the panels
	activeBuildsList := d.renderActiveBuilds(leftWidth, d.height/2)
	activeBuilds := d.renderBox(fmt.Sprintf("ACTIVE BUILDS (%d)", activeCount), leftWidth, d.height/2, d.active == 0, activeBuildsList)

	errors := d.renderBox("ERROR ANALYSIS (0)", leftWidth, d.height-(d.height/2), d.active == 3, "")

	flamechart := d.renderBox("FLAMECHART", rightWidth, d.height/4, d.active == 1, "")
	logs := d.renderBox("LOG OUTPUT", rightWidth, d.height/2, d.active == 2, "")
	stats := d.renderBox("SYSTEM STATS", rightWidth, d.height-(d.height/4)-(d.height/2), d.active == 4, "")

	leftCol := lipgloss.JoinVertical(lipgloss.Left, activeBuilds, errors)
	rightCol := lipgloss.JoinVertical(lipgloss.Left, flamechart, logs, stats)

	return lipgloss.JoinHorizontal(lipgloss.Top, leftCol, rightCol)
}

func (d *Dashboard) renderBox(title string, width, height int, focused bool, content string) string {
	borderColor := lipgloss.Color("#45475A") // inactive
	titleColor := lipgloss.Color("#6C7086")
	titlePrefix := "╭ "

	if focused {
		borderColor = lipgloss.Color("#CBA6F7") // active mauve
		titleColor = lipgloss.Color("#CBA6F7")
		titlePrefix = "╭▸ "
	}

	t := lipgloss.NewStyle().Foreground(titleColor).Bold(focused).Render(title)

	// Create Custom border with embedded title
	border := lipgloss.RoundedBorder()
	titleRow := titlePrefix + t + " " + safeRepeat(border.Top, width-lipgloss.Width(title)-4) + "╮"

	contentStyle := lipgloss.NewStyle().
		Width(width - 2).
		Height(height - 2).
		Border(lipgloss.Border{
			Left: border.Left, Right: border.Right, Bottom: border.Bottom,
			BottomLeft: border.BottomLeft, BottomRight: border.BottomRight,
		}).
		BorderForeground(borderColor)

	return lipgloss.JoinVertical(lipgloss.Left,
		lipgloss.NewStyle().Foreground(borderColor).Render(titleRow),
		contentStyle.Render(content))
}

func (d *Dashboard) renderActiveBuilds(width, height int) string {
	active := d.bm.Active()
	if len(active) == 0 {
		return "\n\n   No active builds found.\n   Start a command or wait\n   for auto-discovery."
	}

	var rows []string
	for _, b := range active {
		progress := int(b.Progress * 10)
		barColor := "#A6E3A1" // green
		if b.State == core.StateFailed {
			barColor = "#F38BA8" // red
		}

		bar := lipgloss.NewStyle().Foreground(lipgloss.Color(barColor)).Render(strings.Repeat("█", progress)) +
			lipgloss.NewStyle().Foreground(lipgloss.Color("#313244")).Render(strings.Repeat("░", 10-progress))

		icon := "" // clock
		if b.State == core.StateBuilding {
			icon = "󱓞" // building
		} else if b.State == core.StateSuccess {
			icon = "check"
		}

		status := lipgloss.NewStyle().Foreground(lipgloss.Color("#6C7086")).Render(b.State.String())
		rows = append(rows, fmt.Sprintf(" %s %-18s %s %s", icon, b.Name, bar, status))
	}

	return strings.Join(rows, "\n")
}

func (d *Dashboard) Focus() {}
func (d *Dashboard) Blur()  {}
