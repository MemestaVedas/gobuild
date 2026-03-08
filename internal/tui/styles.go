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
	StatusMode     lipgloss.Style
	StatusSegment1 lipgloss.Style // e.g. Branch/Project
	StatusSegment2 lipgloss.Style // e.g. Stats/LSP
	StatusSegment3 lipgloss.Style // e.g. Language
	StatusError    lipgloss.Style
	StatusWarning  lipgloss.Style
	HintsBar       lipgloss.Style
}

// DefaultStyles returns the default "aurora-pastel" theme styles.
func DefaultStyles() Styles {
	s := Styles{
		ColorNormal:  lipgloss.Color("#89B4FA"), // Blue (Normal)
		ColorInsert:  lipgloss.Color("#A6E3A1"), // Green (Insert)
		ColorCommand: lipgloss.Color("#FAB387"), // Orange (Command)

		ColorSuccess: lipgloss.Color("#A6E3A1"),
		ColorFailed:  lipgloss.Color("#F38BA8"),
		ColorRunning: lipgloss.Color("#89DCEB"),

		ColorText:      lipgloss.Color("#CDD6F4"),
		ColorSubtext:   lipgloss.Color("#6C7086"),
		ColorHighlight: lipgloss.Color("#313244"),

		ColorBorderActive:   lipgloss.Color("#CBA6F7"), // Mauve
		ColorBorderInactive: lipgloss.Color("#1E1E2E"), // Crust/Base
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
		Background(s.ColorNormal).
		Foreground(lipgloss.Color("#11111B")).
		Bold(true).
		Padding(0, 1)

	s.StatusSegment1 = lipgloss.NewStyle().
		Background(lipgloss.Color("#313244")).
		Foreground(s.ColorText).
		Padding(0, 1)

	s.StatusSegment2 = lipgloss.NewStyle().
		Foreground(s.ColorSubtext).
		Padding(0, 1)

	s.StatusSegment3 = lipgloss.NewStyle().
		Foreground(s.ColorSubtext).
		Padding(0, 1)

	s.StatusError = lipgloss.NewStyle().
		Foreground(s.ColorFailed).
		Bold(true)

	s.StatusWarning = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#F9E2AF")).
		Bold(true)

	s.HintsBar = lipgloss.NewStyle().
		Background(lipgloss.Color("#181825")).
		Foreground(s.ColorSubtext).
		Height(1).
		Padding(0, 1)

	return s
}
