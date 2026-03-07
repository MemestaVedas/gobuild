package tui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

// StatusBarModel renders the main status bar showing Mode, Specs, and Connection.
type StatusBarModel struct {
	styles       Styles
	version      string
	appName      string
	isConnected  bool
	cpuPct       float64
	ramUsedBytes uint64
	netUp        uint64
	netDn        uint64
}

// NewStatusBarModel creates a status bar model.
func NewStatusBarModel(styles Styles) StatusBarModel {
	return StatusBarModel{
		styles:      styles,
		version:     "v1.1",
		appName:     "GoBuild",
		isConnected: false,
	}
}

// SetConnected alters the connection text.
func (s *StatusBarModel) SetConnected(connected bool) {
	s.isConnected = connected
}

func (s *StatusBarModel) UpdateStats(cpuPct float64, ramBytes, netUp, netDn uint64) {
	s.cpuPct = cpuPct
	s.ramUsedBytes = ramBytes
	s.netUp = netUp
	s.netDn = netDn
}

// View renders the bottom statusline with dynamic mode colouring.
func (s *StatusBarModel) View(mode Mode, width int) string {
	var bg lipgloss.Color
	switch mode {
	case ModeInsert:
		bg = s.styles.ColorInsert
	case ModeCommand:
		bg = s.styles.ColorCommand
	default:
		bg = s.styles.ColorNormal
	}

	appPill := s.styles.StatusMode.Copy().
		Background(bg).
		Render(fmt.Sprintf("%s %s", s.appName, s.version))

	modePill := s.styles.StatusMode.Copy().
		Background(bg).
		Render(mode.String())

	connStr := "○ Offline"
	if s.isConnected {
		connStr = "● Connected"
	}
	connPill := s.styles.StatusConnected.Copy().
		Foreground(s.styles.ColorSuccess).
		Render(connStr)

	statsStr := fmt.Sprintf("│ CPU %.0f%% │ RAM %.1fG │ ↑%d ↓%dKB",
		s.cpuPct,
		float64(s.ramUsedBytes)/1024/1024/1024,
		s.netUp/1024, s.netDn/1024)

	statsPill := s.styles.StatusStats.Render(statsStr)

	left := lipgloss.JoinHorizontal(lipgloss.Center, appPill, " │ ", modePill, " │ ", connPill)

	// Simple placement — space between if there's enough room
	w := width - lipgloss.Width(left) - lipgloss.Width(statsPill)
	if w < 0 {
		w = 0 // truncate gracefully
	}
	space := s.styles.HintsBar.Copy().Width(w).Render("")

	return lipgloss.JoinHorizontal(lipgloss.Center, left, space, statsPill)
}
