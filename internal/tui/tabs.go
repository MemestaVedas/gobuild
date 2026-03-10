package tui

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
)

// TabBarModel manages the top tab bar display.
type TabBarModel struct {
	styles    Styles
	activeTab int
	tabs      []string
}

func NewTabBarModel(styles Styles) TabBarModel {
	return TabBarModel{
		styles: styles,
		tabs: []string{
			"Dashboard",
			"Launcher",
			"History",
			"Plugins",
			"Help",
		},
	}
}

func (t *TabBarModel) SetActive(idx int) {
	if idx >= 0 && idx < len(t.tabs) {
		t.activeTab = idx
	}
}

// View renders a lazygit-inspired tab row: ─[1]Dashboard─[2]Launcher─...
func (t *TabBarModel) View(width int) string {
	var parts []string

	for i, name := range t.tabs {
		label := fmt.Sprintf("[%d]%s", i+1, name)
		if i == t.activeTab {
			parts = append(parts, lipgloss.NewStyle().
				Foreground(t.styles.ColorAccent).
				Bold(true).
				Render(label))
		} else {
			parts = append(parts, lipgloss.NewStyle().
				Foreground(t.styles.ColorFaint).
				Render(label))
		}
	}

	sep := lipgloss.NewStyle().Foreground(t.styles.ColorBorderInactive).Render("   ")
	row := " " + strings.Join(parts, sep) + " "

	// Fill remaining width
	pad := width - lipgloss.Width(row)
	if pad < 0 {
		pad = 0
	}
	filler := lipgloss.NewStyle().Foreground(t.styles.ColorBorderInactive).Render(strings.Repeat(" ", pad))
	return row + filler
}
