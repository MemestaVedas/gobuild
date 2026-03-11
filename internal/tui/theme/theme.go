package theme

import (
	"image/color"

	"charm.land/lipgloss/v2"
)

// Styles defines the application-wide color palette and UI elements.
type Styles struct {
	// ── Colors ───────────────────────────────────────────────────────────
	ColorNormal  color.Color
	ColorInsert  color.Color
	ColorCommand color.Color

	ColorSuccess color.Color
	ColorWarning color.Color
	ColorFailed  color.Color
	ColorRunning color.Color
	ColorQueued  color.Color

	ColorText      color.Color
	ColorSubtext   color.Color
	ColorFaint     color.Color
	ColorHighlight color.Color
	ColorSurface   color.Color
	ColorBase      color.Color
	ColorCrust     color.Color

	ColorBorderActive   color.Color
	ColorBorderInactive color.Color
	ColorBorderDim      color.Color

	ColorAccent color.Color // Primary green accent (lazygit-style)

	ColorHintKey  color.Color
	ColorHintDesc color.Color
}

// DefaultStyles returns the lazygit-inspired + aurora-pastel dark theme.
func DefaultStyles(isDark bool) Styles {
	ld := lipgloss.LightDark(isDark)
	return Styles{
		// Mode indicators — colourful!
		ColorNormal:  ld(lipgloss.Color("#5588FF"), lipgloss.Color("#89B4FA")),
		ColorInsert:  ld(lipgloss.Color("#228822"), lipgloss.Color("#A6E3A1")),
		ColorCommand: ld(lipgloss.Color("#EE6600"), lipgloss.Color("#FAB387")),

		// Build state — vivid pastels
		ColorSuccess: ld(lipgloss.Color("#228822"), lipgloss.Color("#A6E3A1")),
		ColorWarning: ld(lipgloss.Color("#DD9900"), lipgloss.Color("#F9E2AF")),
		ColorFailed:  ld(lipgloss.Color("#CC2222"), lipgloss.Color("#F38BA8")),
		ColorRunning: ld(lipgloss.Color("#0088AA"), lipgloss.Color("#89DCEB")),
		ColorQueued:  ld(lipgloss.Color("#8822AA"), lipgloss.Color("#CBA6F7")),

		// Text ramp
		ColorText:    ld(lipgloss.Color("#11111B"), lipgloss.Color("#CDD6F4")),
		ColorSubtext: ld(lipgloss.Color("#313244"), lipgloss.Color("#CDD6F4")), // Brightened
		ColorFaint:   ld(lipgloss.Color("#45475A"), lipgloss.Color("#A6ADC8")), // Brightened

		// Surface ramp
		ColorCrust:     ld(lipgloss.Color("#EEEEEE"), lipgloss.Color("#11111B")),
		ColorBase:      ld(lipgloss.Color("#FFFFFF"), lipgloss.Color("#1E1E2E")),
		ColorSurface:   ld(lipgloss.Color("#DDDDDD"), lipgloss.Color("#313244")),
		ColorHighlight: ld(lipgloss.Color("#CCCCCC"), lipgloss.Color("#45475A")),

		// Borders & accent
		ColorAccent:         ld(lipgloss.Color("#228822"), lipgloss.Color("#A6E3A1")),
		ColorBorderActive:   ld(lipgloss.Color("#228822"), lipgloss.Color("#A6E3A1")),
		ColorBorderInactive: ld(lipgloss.Color("#888888"), lipgloss.Color("#6C7086")), // Brighter border
		ColorBorderDim:      ld(lipgloss.Color("#CCCCCC"), lipgloss.Color("#313244")),

		ColorHintKey:  ld(lipgloss.Color("#228822"), lipgloss.Color("#A6E3A1")),
		ColorHintDesc: ld(lipgloss.Color("#444444"), lipgloss.Color("#89ADC8")),
	}
}
