package tui

import "github.com/charmbracelet/lipgloss"

// Styles defines the application-wide color palette and UI elements.
type Styles struct {
	// ── Colors ───────────────────────────────────────────────────────────
	ColorNormal  lipgloss.Color
	ColorInsert  lipgloss.Color
	ColorCommand lipgloss.Color

	ColorSuccess lipgloss.Color
	ColorWarning lipgloss.Color
	ColorFailed  lipgloss.Color
	ColorRunning lipgloss.Color
	ColorQueued  lipgloss.Color

	ColorText      lipgloss.Color
	ColorSubtext   lipgloss.Color
	ColorFaint     lipgloss.Color
	ColorHighlight lipgloss.Color
	ColorSurface   lipgloss.Color
	ColorBase      lipgloss.Color
	ColorCrust     lipgloss.Color

	ColorBorderActive   lipgloss.Color
	ColorBorderInactive lipgloss.Color
	ColorBorderDim      lipgloss.Color

	ColorAccent lipgloss.Color // Primary green accent (lazygit-style)

	// ── Panels ───────────────────────────────────────────────────────────
	PanelActive   lipgloss.Style
	PanelInactive lipgloss.Style
	TitleActive   lipgloss.Style
	TitleInactive lipgloss.Style

	// ── Tabs ─────────────────────────────────────────────────────────────
	TabActive   lipgloss.Style
	TabInactive lipgloss.Style

	// ── Status bar ───────────────────────────────────────────────────────
	StatusMode     lipgloss.Style
	StatusSegment1 lipgloss.Style
	StatusSegment2 lipgloss.Style
	StatusSegment3 lipgloss.Style
	StatusError    lipgloss.Style
	StatusWarning  lipgloss.Style
	HintsBar       lipgloss.Style
	HintKey        lipgloss.Style
	HintDesc       lipgloss.Style

	// ── Inputs ───────────────────────────────────────────────────────────
	InputFocused   lipgloss.Style
	InputUnfocused lipgloss.Style
	Suggestion     lipgloss.Style
	GhostText      lipgloss.Style
}

// DefaultStyles returns the lazygit-inspired + aurora-pastel dark theme.
func DefaultStyles() Styles {
	s := Styles{
		// Mode indicators — colourful!
		ColorNormal:  lipgloss.Color("#89B4FA"), // Soft blue
		ColorInsert:  lipgloss.Color("#A6E3A1"), // Green
		ColorCommand: lipgloss.Color("#FAB387"), // Peach

		// Build state — vivid pastels
		ColorSuccess: lipgloss.Color("#A6E3A1"),
		ColorWarning: lipgloss.Color("#F9E2AF"),
		ColorFailed:  lipgloss.Color("#F38BA8"),
		ColorRunning: lipgloss.Color("#89DCEB"),
		ColorQueued:  lipgloss.Color("#CBA6F7"),

		// Text ramp
		ColorText:    lipgloss.Color("#CDD6F4"),
		ColorSubtext: lipgloss.Color("#A6ADC8"),
		ColorFaint:   lipgloss.Color("#585B70"),

		// Surface ramp (dark)
		ColorCrust:     lipgloss.Color("#11111B"),
		ColorBase:      lipgloss.Color("#1E1E2E"),
		ColorSurface:   lipgloss.Color("#313244"),
		ColorHighlight: lipgloss.Color("#45475A"),

		// Borders & accent — green primary like lazygit
		ColorAccent:         lipgloss.Color("#A6E3A1"),
		ColorBorderActive:   lipgloss.Color("#A6E3A1"), // Green for focused panel
		ColorBorderInactive: lipgloss.Color("#585B70"), // Subtle grey
		ColorBorderDim:      lipgloss.Color("#313244"),
	}

	border := lipgloss.NormalBorder()
	basePanel := lipgloss.NewStyle().Border(border).Padding(0, 1)

	s.PanelActive = basePanel.Copy().BorderForeground(s.ColorBorderActive)
	s.PanelInactive = basePanel.Copy().BorderForeground(s.ColorBorderInactive)

	s.TitleActive = lipgloss.NewStyle().Foreground(s.ColorAccent).Bold(true)
	s.TitleInactive = lipgloss.NewStyle().Foreground(s.ColorFaint)

	// Tab bar — clean, no backgrounds, just color change
	s.TabActive = lipgloss.NewStyle().
		Foreground(s.ColorAccent).
		Bold(true).
		Padding(0, 1)

	s.TabInactive = lipgloss.NewStyle().
		Foreground(s.ColorFaint).
		Padding(0, 1)

	// Status / mode pill
	s.StatusMode = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#11111B")).
		Bold(true).
		Padding(0, 1)

	s.StatusSegment1 = lipgloss.NewStyle().
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
		Bold(true).
		Padding(0, 1)

	s.StatusWarning = lipgloss.NewStyle().
		Foreground(s.ColorWarning).
		Bold(true).
		Padding(0, 1)

	// Hints bar — lazygit style: green keys, faint descriptions, pipe separators
	s.HintsBar = lipgloss.NewStyle().
		Foreground(s.ColorFaint).
		Height(1).
		Padding(0, 0)

	s.HintKey = lipgloss.NewStyle().
		Foreground(s.ColorAccent).
		Bold(true)

	s.HintDesc = lipgloss.NewStyle().
		Foreground(s.ColorFaint)

	// Inputs
	s.InputFocused = lipgloss.NewStyle().
		Foreground(s.ColorText).
		Padding(0, 1).
		Border(lipgloss.NormalBorder(), false, false, false, true).
		BorderForeground(s.ColorAccent)

	s.InputUnfocused = lipgloss.NewStyle().
		Foreground(s.ColorSubtext).
		Padding(0, 1).
		Border(lipgloss.NormalBorder(), false, false, false, true).
		BorderForeground(s.ColorBorderInactive)

	s.Suggestion = lipgloss.NewStyle().
		Foreground(s.ColorText).
		Background(s.ColorSurface)

	s.GhostText = lipgloss.NewStyle().
		Foreground(s.ColorFaint)

	return s
}
