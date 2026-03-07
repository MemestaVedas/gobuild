package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// TabBarModel manages the display of the top 5 tabs.
type TabBarModel struct {
	styles    Styles
	activeTab int
	tabs      []string
}

// NewTabBarModel creates a tab bar model.
func NewTabBarModel(styles Styles) TabBarModel {
	return TabBarModel{
		styles: styles,
		tabs: []string{
			"1 Dashboard",
			"2 Launcher",
			"3 History",
			"4 Plugins",
			"5 Help",
		},
	}
}

// SetActive sets the current active tab index (0-4).
func (t *TabBarModel) SetActive(idx int) {
	if idx >= 0 && idx < len(t.tabs) {
		t.activeTab = idx
	}
}

// View renders the tab row with Neovim buffering style.
func (t *TabBarModel) View(width int) string {
	var rendered []string

	for i, name := range t.tabs {
		var s lipgloss.Style
		if i == t.activeTab {
			s = t.styles.TabActive.Copy().Background(t.styles.ColorHighlight)
		} else {
			s = t.styles.TabInactive
		}
		rendered = append(rendered, s.Render(fmt.Sprintf(" %s ", name)))
	}

	row := strings.Join(rendered, " ")
	padding := width - lipgloss.Width(row)
	if padding < 0 {
		padding = 0
	}
	space := strings.Repeat(" ", padding)

	return lipgloss.JoinHorizontal(lipgloss.Top, row, space)
}
