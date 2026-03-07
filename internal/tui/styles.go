package tui

import "github.com/charmbracelet/lipgloss"

// Styles defines the application-wide color palette and UI borders.
type Styles struct {
	// Colors
	ColorNormal  lipgloss.Color
	ColorInsert  lipgloss.Color
	ColorCommand lipgloss.Color

	ColorSuccess lipgloss.Color
	ColorFailed  lipgloss.Color
	ColorRunning lipgloss.Color

	ColorText      lipgloss.Color
	ColorSubtext   lipgloss.Color
	ColorHighlight lipgloss.Color

	ColorBorderActive   lipgloss.Color
	ColorBorderInactive lipgloss.Color

	// Panels
	PanelActive   lipgloss.Style
	PanelInactive lipgloss.Style
	TitleActive   lipgloss.Style
	TitleInactive lipgloss.Style

	// Tabs
	TabActive   lipgloss.Style
	TabInactive lipgloss.Style

	// Statusline
	StatusMode      lipgloss.Style
	StatusConnected lipgloss.Style
	StatusStats     lipgloss.Style
	HintsBar        lipgloss.Style
}

// DefaultStyles returns the default "aurora-pastel" theme styles.
func DefaultStyles() Styles {
	s := Styles{
		ColorNormal:  lipgloss.Color("#89B4FA"), // Blue
		ColorInsert:  lipgloss.Color("#A6E3A1"), // Green
		ColorCommand: lipgloss.Color("#FAB387"), // Orange

		ColorSuccess: lipgloss.Color("#A6E3A1"), // Sage Green
		ColorFailed:  lipgloss.Color("#F38BA8"), // Rose Red
		ColorRunning: lipgloss.Color("#89DCEB"), // Sky Blue

		ColorText:      lipgloss.Color("#CDD6F4"),
		ColorSubtext:   lipgloss.Color("#6C7086"),
		ColorHighlight: lipgloss.Color("#313244"),

		ColorBorderActive:   lipgloss.Color("#CBA6F7"), // Mauve/Purple
		ColorBorderInactive: lipgloss.Color("#45475A"), // Dark Grey
	}

	border := lipgloss.RoundedBorder()

	basePanel := lipgloss.NewStyle().
		Border(border).
		Padding(0, 1)

	s.PanelActive = basePanel.Copy().
		BorderForeground(s.ColorBorderActive)
	s.PanelInactive = basePanel.Copy().
		BorderForeground(s.ColorBorderInactive)

	s.TitleActive = lipgloss.NewStyle().
		Foreground(s.ColorBorderActive).
		Bold(true)
	s.TitleInactive = lipgloss.NewStyle().
		Foreground(s.ColorSubtext)

	s.TabActive = lipgloss.NewStyle().
		Foreground(s.ColorNormal).
		Bold(true).
		Padding(0, 1)

	s.TabInactive = lipgloss.NewStyle().
		Foreground(s.ColorSubtext).
		Padding(0, 1)

	s.StatusMode = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#11111B")).
		Bold(true).
		Padding(0, 1)

	s.StatusConnected = lipgloss.NewStyle().
		Foreground(s.ColorSuccess)

	s.StatusStats = lipgloss.NewStyle().
		Foreground(s.ColorText)

	s.HintsBar = lipgloss.NewStyle().
		Foreground(s.ColorSubtext).
		Padding(0, 1)

	return s
}
