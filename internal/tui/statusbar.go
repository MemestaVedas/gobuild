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
	activeTab    int
	totalErrors  int
	totalWarns   int
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

func (s *StatusBarModel) SetErrors(errs, warns int) {
	s.totalErrors = errs
	s.totalWarns = warns
}

func (s *StatusBarModel) SetActiveTab(tab int) {
	s.activeTab = tab
}

const (
	separator = ""
	branch    = ""
)

// View renders the bottom statusline with dynamic mode colouring.
func (s *StatusBarModel) View(mode Mode, width int) string {
	var modeColor lipgloss.Color
	var modeText string
	switch mode {
	case ModeInsert:
		modeColor = s.styles.ColorInsert
		modeText = " INSERT "
	case ModeCommand:
		modeColor = s.styles.ColorCommand
		modeText = " COMMAND "
	default:
		modeColor = s.styles.ColorNormal
		modeText = " NORMAL "
	}

	// 1. Mode segment (Inverse)
	modeStyle := s.styles.StatusMode.Copy().Background(modeColor)
	modeSeg := modeStyle.Render(modeText)

	// 2. Project/Tab name segment
	tabNames := []string{" Dashboard ", " Launcher ", " History ", " Plugins ", " Help "}
	tabName := " GoBuild "
	if s.activeTab >= 0 && s.activeTab < len(tabNames) {
		tabName = tabNames[s.activeTab]
	}
	projSeg := s.styles.StatusSegment1.Render(tabName)

	// 3. Middle spacers / icons
	branchSeg := s.styles.StatusSegment2.Render(fmt.Sprintf("%s master", branch))

	// 4. Stats / Errors (Right side)
	errIcon := ""
	if s.totalErrors > 0 {
		errIcon = s.styles.StatusError.Render(fmt.Sprintf("  %d", s.totalErrors))
	}
	warnIcon := ""
	if s.totalWarns > 0 {
		warnIcon = s.styles.StatusWarning.Render(fmt.Sprintf("  %d", s.totalWarns))
	}

	statsText := fmt.Sprintf("  %.0f%%  %.1fG ", s.cpuPct, float64(s.ramUsedBytes)/1024/1024/1024)
	statsSeg := s.styles.StatusSegment3.Render(statsText)

	lspSeg := s.styles.StatusSegment2.Render(" build-on")

	left := lipgloss.JoinHorizontal(lipgloss.Top, modeSeg, projSeg, branchSeg)
	right := lipgloss.JoinHorizontal(lipgloss.Top, errIcon, warnIcon, statsSeg, lspSeg)

	// Join with space
	availableSpace := width - lipgloss.Width(left) - lipgloss.Width(right)
	if availableSpace < 0 {
		availableSpace = 0
	}
	middleSpace := s.styles.StatusSegment2.Copy().Width(availableSpace).Render("")

	return lipgloss.JoinHorizontal(lipgloss.Top, left, middleSpace, right)
}
