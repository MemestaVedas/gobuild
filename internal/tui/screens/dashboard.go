package screens

import (
	"fmt"
	"strings"

	"github.com/MemestaVedas/gobuild/internal/core"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ── Palette (lazygit-inspired + aurora pastels) ─────────────────────────────
var (
	dBorderActive   = lipgloss.Color("#A6E3A1") // Green — focused panel
	dBorderInactive = lipgloss.Color("#585B70") // Grey — unfocused
	dText           = lipgloss.Color("#CDD6F4")
	dSubtext        = lipgloss.Color("#A6ADC8")
	dFaint          = lipgloss.Color("#585B70")
	dSurface        = lipgloss.Color("#313244")
	dGreen          = lipgloss.Color("#A6E3A1")
	dRed            = lipgloss.Color("#F38BA8")
	dYellow         = lipgloss.Color("#F9E2AF")
	dBlue           = lipgloss.Color("#89B4FA")
	dMauve          = lipgloss.Color("#CBA6F7")
	dSky            = lipgloss.Color("#89DCEB")
)

type Dashboard struct {
	width  int
	height int
	active int // 0-4: which panel is focused
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
		d.height = msg.Height - 3
	case tea.KeyMsg:
		switch msg.String() {
		case "tab":
			d.active = (d.active + 1) % 5
		case "shift+tab":
			d.active = (d.active - 1 + 5) % 5
		case "ctrl+h":
			if d.active == 1 || d.active == 2 || d.active == 4 {
				d.active = 0
			}
		case "ctrl+l":
			if d.active == 0 || d.active == 3 {
				d.active = 2
			}
		case "ctrl+k":
			switch d.active {
			case 3:
				d.active = 0
			case 2:
				d.active = 1
			case 4:
				d.active = 2
			}
		case "ctrl+j":
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

	leftW := 46
	rightW := d.width - leftW

	activeCount := len(d.bm.Active())
	errors := d.bm.AllErrors()

	// ── Left column ─────────────────────────────────────────────────────
	buildsPanel := d.renderPanel(
		fmt.Sprintf("[1]-Active Builds (%d)", activeCount),
		leftW, d.height/2, d.active == 0,
		d.renderActiveBuilds(leftW-4))

	errPanel := d.renderPanel(
		fmt.Sprintf("[3]-Errors (%d)", len(errors)),
		leftW, d.height-(d.height/2), d.active == 3,
		d.renderErrors(errors, leftW-4))

	// ── Right column ────────────────────────────────────────────────────
	flamechart := d.renderPanel(
		"[2]-Flamechart",
		rightW, d.height/4, d.active == 1,
		d.renderFlame(rightW-4))

	logPanel := d.renderPanel(
		"[4]-Log Output",
		rightW, d.height/2, d.active == 2,
		d.renderLogs(rightW-4))

	statsPanel := d.renderPanel(
		"[5]-System Stats",
		rightW, d.height-(d.height/4)-(d.height/2), d.active == 4,
		d.renderStats(rightW-4))

	leftCol := lipgloss.JoinVertical(lipgloss.Left, buildsPanel, errPanel)
	rightCol := lipgloss.JoinVertical(lipgloss.Left, flamechart, logPanel, statsPanel)

	return lipgloss.JoinHorizontal(lipgloss.Top, leftCol, rightCol)
}

// renderPanel draws a lazygit-style panel: ─[N]-Title───────────────
func (d *Dashboard) renderPanel(title string, width, height int, focused bool, content string) string {
	borderColor := dBorderInactive
	titleColor := dFaint
	if focused {
		borderColor = dBorderActive
		titleColor = dGreen
	}

	bc := lipgloss.NewStyle().Foreground(borderColor)
	tc := lipgloss.NewStyle().Foreground(titleColor).Bold(focused)

	// Top border: ─[N]-Title──────────
	titleStr := tc.Render(title)
	titleW := lipgloss.Width(title) + 2 // account for ─ before and after
	fillW := width - titleW - 1
	if fillW < 0 {
		fillW = 0
	}
	topLine := bc.Render("─") + titleStr + bc.Render(strings.Repeat("─", fillW))

	// Content box (left, right, bottom borders — no top, the title line IS the top)
	b := lipgloss.NormalBorder()
	contentStyle := lipgloss.NewStyle().
		Width(width - 2).
		Height(height - 2).
		Foreground(dText).
		Border(lipgloss.Border{
			Left: b.Left, Right: b.Right, Bottom: b.Bottom,
			BottomLeft: b.BottomLeft, BottomRight: b.BottomRight,
		}).
		BorderForeground(borderColor)

	return lipgloss.JoinVertical(lipgloss.Left, topLine, contentStyle.Render(content))
}

func (d *Dashboard) renderActiveBuilds(width int) string {
	active := d.bm.Active()
	if len(active) == 0 {
		return lipgloss.NewStyle().Foreground(dFaint).Render(
			"\n  No active builds.\n  Use the Launcher [2] to start one\n  or wait for auto-discovery.")
	}

	var rows []string
	for _, b := range active {
		filled := int(b.Progress * 20)
		if filled > 20 {
			filled = 20
		}
		empty := 20 - filled

		barColor := dGreen
		stateIcon := "●"
		switch b.State {
		case core.StateFailed:
			barColor = dRed
			stateIcon = "✖"
		case core.StateQueued:
			barColor = dMauve
			stateIcon = "◌"
		case core.StateBuilding:
			barColor = dSky
			stateIcon = "⟳"
		case core.StateSuccess:
			barColor = dGreen
			stateIcon = "✔"
		}

		bar := lipgloss.NewStyle().Foreground(barColor).Render(strings.Repeat("█", filled)) +
			lipgloss.NewStyle().Foreground(dSurface).Render(strings.Repeat("░", empty))

		pct := int(b.Progress * 100)
		elapsed := int(b.Elapsed().Seconds())

		icon := lipgloss.NewStyle().Foreground(barColor).Render(stateIcon)
		name := lipgloss.NewStyle().Foreground(dText).Bold(true).Width(16).Render(truncate(b.Name, 16))
		tool := lipgloss.NewStyle().Foreground(dBlue).Render(b.Tool.String())
		stat := lipgloss.NewStyle().Foreground(dFaint).Render(fmt.Sprintf("%3d%% %ds", pct, elapsed))

		rows = append(rows, fmt.Sprintf("  %s %s %s", icon, name, tool))
		rows = append(rows, fmt.Sprintf("    %s %s", bar, stat))
		rows = append(rows, "")
	}
	return strings.Join(rows, "\n")
}

func (d *Dashboard) renderErrors(errors []core.BuildError, width int) string {
	if len(errors) == 0 {
		return lipgloss.NewStyle().Foreground(dGreen).Render("\n  No errors ✔")
	}
	var rows []string
	for _, e := range errors {
		c := dRed
		if e.Level == core.LogWarning {
			c = dYellow
		}
		dot := lipgloss.NewStyle().Foreground(c).Bold(true).Render("●")
		loc := lipgloss.NewStyle().Foreground(dSubtext).Render(
			fmt.Sprintf("%s:%d", truncate(e.File, 18), e.Line))
		msg := lipgloss.NewStyle().Foreground(dText).Render(
			truncate(e.Message, width-28))
		rows = append(rows, fmt.Sprintf("  %s %s  %s", dot, loc, msg))
	}
	return strings.Join(rows, "\n")
}

func (d *Dashboard) renderFlame(width int) string {
	return lipgloss.NewStyle().Foreground(dFaint).
		Render("\n  No stage data available.\n  Supported: Cargo, Webpack")
}

func (d *Dashboard) renderLogs(width int) string {
	all := d.bm.All()
	var lines []string
	for _, b := range all {
		for _, ll := range b.LogLines {
			var c lipgloss.Color
			switch ll.Level {
			case core.LogError:
				c = dRed
			case core.LogWarning:
				c = dYellow
			case core.LogNote:
				c = dMauve
			default:
				c = dSubtext
			}
			ts := lipgloss.NewStyle().Foreground(dFaint).Render(ll.Timestamp.Format("15:04:05") + " ")
			line := ts + lipgloss.NewStyle().Foreground(c).Render(truncate(ll.Raw, width-12))
			lines = append(lines, line)
		}
	}
	if len(lines) == 0 {
		return lipgloss.NewStyle().Foreground(dFaint).Render("\n  Waiting for log output...")
	}
	maxLines := 12
	if len(lines) > maxLines {
		lines = lines[len(lines)-maxLines:]
	}
	return strings.Join(lines, "\n")
}

func (d *Dashboard) renderStats(width int) string {
	row := func(label, val string, valColor lipgloss.Color) string {
		l := lipgloss.NewStyle().Foreground(dSubtext).Width(20).Render("  " + label)
		v := lipgloss.NewStyle().Foreground(valColor).Bold(true).Render(val)
		return l + v
	}

	all := d.bm.All()
	total := len(all)
	failed := 0
	for _, b := range all {
		if b.State == core.StateFailed {
			failed++
		}
	}

	return strings.Join([]string{
		"",
		row("Total builds", fmt.Sprintf("%d", total), dText),
		row("Failed", fmt.Sprintf("%d", failed), dRed),
		row("Success rate", func() string {
			if total == 0 {
				return "—"
			}
			return fmt.Sprintf("%.0f%%", float64(total-failed)/float64(total)*100)
		}(), dGreen),
	}, "\n")
}

func (d *Dashboard) Focus() {}
func (d *Dashboard) Blur()  {}

func truncate(s string, max int) string {
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	return string(r[:max-1]) + "…"
}
