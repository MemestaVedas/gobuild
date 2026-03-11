package tui

import (
	"fmt"
	"image/color"
	"strings"
	"github.com/MemestaVedas/gobuild/internal/tui/theme"
	"charm.land/lipgloss/v2"
)

// StatusBarModel renders the bottom hints bar (lazygit-style).
type StatusBarModel struct {
	styles       theme.Styles
	version      string
	appName      string
	isConnected  bool
	cpuPct       float64
	ramUsedBytes uint64
	netUp        uint64
	netDn        uint64
	activeTab    int
	totalErrors  int
	totalWarns   int
}

func NewStatusBarModel(styles theme.Styles) StatusBarModel {
	return StatusBarModel{
		styles:  styles,
		version: "v1.1",
		appName: "GoBuild",
	}
}

func (s *StatusBarModel) SetConnected(connected bool) { s.isConnected = connected }
func (s *StatusBarModel) UpdateStats(cpuPct float64, ramBytes, netUp, netDn uint64) {
	s.cpuPct = cpuPct
	s.ramUsedBytes = ramBytes
	s.netUp = netUp
	s.netDn = netDn
}
func (s *StatusBarModel) SetErrors(errs, warns int) { s.totalErrors = errs; s.totalWarns = warns }
func (s *StatusBarModel) SetActiveTab(tab int)       { s.activeTab = tab }

// View renders a 2-line block: mode + stats, then hints bar.
func (s *StatusBarModel) View(mode Mode, width int) string {
	return lipgloss.JoinVertical(lipgloss.Left,
		s.renderModeLine(mode, width),
		s.renderHints(mode, width),
	)
}

func (s *StatusBarModel) renderModeLine(mode Mode, width int) string {
	// Mode pill
	var modeColor color.Color
	var modeText string
	switch mode {
	case ModeInsert:
		modeColor = s.styles.ColorInsert
		modeText = "INSERT"
	case ModeCommand:
		modeColor = s.styles.ColorCommand
		modeText = "COMMAND"
	default:
		modeColor = s.styles.ColorNormal
		modeText = "NORMAL"
	}

	modeSeg := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#11111B")).
		Background(modeColor).
		Bold(true).
		Padding(0, 1).
		Render(modeText)

	// Tab name
	tabNames := []string{"Dashboard", "Launcher", "History", "Plugins", "Help"}
	tab := s.appName
	if s.activeTab >= 0 && s.activeTab < len(tabNames) {
		tab = tabNames[s.activeTab]
	}
	tabSeg := lipgloss.NewStyle().
		Foreground(s.styles.ColorAccent).
		Bold(true).
		Padding(0, 1).
		Render(tab)

	// Right: CPU + RAM + errors
	var rightParts []string

	if s.totalErrors > 0 {
		rightParts = append(rightParts,
			lipgloss.NewStyle().Foreground(s.styles.ColorFailed).Bold(true).Render(
				fmt.Sprintf("✖ %d", s.totalErrors)))
	}
	if s.totalWarns > 0 {
		rightParts = append(rightParts,
			lipgloss.NewStyle().Foreground(s.styles.ColorWarning).Bold(true).Render(
				fmt.Sprintf("⚠ %d", s.totalWarns)))
	}

	ramGB := float64(s.ramUsedBytes) / 1024 / 1024 / 1024
	rightParts = append(rightParts,
		lipgloss.NewStyle().Foreground(s.styles.ColorFaint).Render(
			fmt.Sprintf(" %.0f%%   %.1fG", s.cpuPct, ramGB)))

	connDot := "●"
	connColor := s.styles.ColorFailed
	if s.isConnected {
		connColor = s.styles.ColorSuccess
	}
	rightParts = append(rightParts,
		lipgloss.NewStyle().Foreground(connColor).Render(connDot))

	rightParts = append(rightParts,
		lipgloss.NewStyle().Foreground(s.styles.ColorFaint).Render(s.version))

	left := modeSeg + " " + tabSeg
	right := strings.Join(rightParts, "  ")

	gap := width - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 0 {
		gap = 0
	}

	return left + strings.Repeat(" ", gap) + right
}

// renderHints produces a lazygit-style hints line:
// Run: r | Kill: x | Focus Log: f | Search: / | Keybindings: ?
func (s *StatusBarModel) renderHints(mode Mode, width int) string {
	if mode == ModeCommand {
		return lipgloss.NewStyle().
			Foreground(s.styles.ColorText).
			Width(width).
			Render(":")
	}

	type hint struct{ desc, key string }
	var hints []hint

	switch s.activeTab {
	case 0: // Dashboard
		hints = []hint{
			{"Run", "r"}, {"Kill", "x"}, {"Focus Log", "f"},
			{"Open Editor", "o"}, {"Details", "d"}, {"Search", "/"},
			{"Keybindings", "?"},
		}
	case 1: // Launcher
		hints = []hint{
			{"Complete", "Tab"}, {"Prev Field", "↑"},
			{"Next Field", "↓"}, {"Del Word", "Ctrl+W"},
			{"Normal", "Esc"},
		}
	case 2: // History
		hints = []hint{
			{"View Logs", "Enter"}, {"Re-run", "r"},
			{"Open", "o"}, {"Search", "/"}, {"Delete", "Del"},
		}
	case 3: // Plugins
		hints = []hint{
			{"Configure", "Enter"}, {"Toggle", "e"}, {"Remove", "Del"},
		}
	default: // Help
		hints = []hint{
			{"Scroll", "j/k"}, {"Quit", "q"},
		}
	}

	keyStyle := lipgloss.NewStyle().Foreground(s.styles.ColorHintKey)
	descStyle := lipgloss.NewStyle().Foreground(s.styles.ColorHintDesc)

	var parts []string
	for _, h := range hints {
		parts = append(parts, descStyle.Render(h.desc+": ")+keyStyle.Render(h.key))
	}

	row := strings.Join(parts, descStyle.Render(" | "))

	pad := width - lipgloss.Width(row)
	if pad < 0 {
		pad = 0
	}
	return row + strings.Repeat(" ", pad)
}
