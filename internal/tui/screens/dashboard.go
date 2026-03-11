package screens

import (
	"fmt"
	"image/color"
	"strings"

	"github.com/MemestaVedas/gobuild/internal/core"
	tea "github.com/charmbracelet/bubbletea"
	"charm.land/lipgloss/v2"
)

// Palette
var (
	dGreen   = lipgloss.Color("#A6E3A1")
	dRed     = lipgloss.Color("#F38BA8")
	dYellow  = lipgloss.Color("#F9E2AF")
	dBlue    = lipgloss.Color("#89B4FA")
	dMauve   = lipgloss.Color("#CBA6F7")
	dSky     = lipgloss.Color("#89DCEB")
	dText    = lipgloss.Color("#CDD6F4")
	dSubtext = lipgloss.Color("#A6ADC8")
	dFaint   = lipgloss.Color("#585B70")
	dSurface = lipgloss.Color("#313244")
	dActive  = lipgloss.Color("#A6E3A1")
	dInactive = lipgloss.Color("#585B70")
)

type Dashboard struct {
	width         int
	height        int
	active        int
	selectedBuild int
	bm            *core.BuildManager
}

func NewDashboard(bm *core.BuildManager) *Dashboard {
	return &Dashboard{bm: bm}
}

func (d *Dashboard) Init() tea.Cmd { return nil }

func (d *Dashboard) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		d.width = msg.Width
		d.height = msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "1":
			d.active = 0
		case "2":
			d.active = 1
		case "3":
			d.active = 2
		case "4":
			d.active = 3
		case "j", "down":
			if d.active == 0 {
				d.selectedBuild++
			}
		case "k", "up":
			if d.active == 0 {
				d.selectedBuild--
			}
		}
	}
	return d, nil
}

func (d *Dashboard) View() string {
	if d.width <= 0 || d.height <= 0 {
		return ""
	}

	leftW := d.width / 3
	rightW := d.width - leftW

	// Build each panel content
	activeBuilds := d.bm.Active()
	
	// Clamp selection
	if len(activeBuilds) == 0 {
		d.selectedBuild = 0
	} else if d.selectedBuild >= len(activeBuilds) {
		d.selectedBuild = len(activeBuilds) - 1
	} else if d.selectedBuild < 0 {
		d.selectedBuild = 0
	}

	errors := d.bm.AllErrors()

	leftTop := d.panel("Active Builds", leftW, d.height/2, d.active == 0,
		d.renderBuilds(activeBuilds, leftW-2))
	leftBot := d.panel(fmt.Sprintf("Errors (%d)", len(errors)), leftW, d.height-d.height/2, d.active == 2,
		d.renderErrors(errors, leftW-2))

	var selectedLogLines []core.LogLine
	if len(activeBuilds) > 0 {
		selectedLogLines = activeBuilds[d.selectedBuild].LogLines
	}

	rightTop := d.panel("Log Output", rightW, d.height/2, d.active == 1,
		d.renderLogs(rightW-2, selectedLogLines))
	rightBot := d.panel("Stats", rightW, d.height-d.height/2, d.active == 3,
		d.renderStats())

	left := lipgloss.JoinVertical(lipgloss.Left, leftTop, leftBot)
	right := lipgloss.JoinVertical(lipgloss.Left, rightTop, rightBot)

	return lipgloss.JoinHorizontal(lipgloss.Top, left, right)
}

// panel renders a titled box. Uses a simple top border line + content box.
func (d *Dashboard) panel(title string, w, h int, focused bool, content string) string {
	bColor := dInactive
	tColor := dFaint
	if focused {
		bColor = dActive
		tColor = dGreen
	}

	// Top line: ── Title ────────
	titleStr := lipgloss.NewStyle().Foreground(tColor).Bold(focused).Render(" " + title + " ")
	tw := lipgloss.Width(titleStr)
	remaining := w - tw - 2
	if remaining < 0 {
		remaining = 0
	}
	topLine := lipgloss.NewStyle().Foreground(bColor).Render("─") +
		titleStr +
		lipgloss.NewStyle().Foreground(bColor).Render(strings.Repeat("─", remaining))

	// Content box below the title line
	box := lipgloss.NewStyle().
		Width(w).
		Height(h - 1). // subtract top line
		BorderLeft(true).
		BorderBottom(true).
		BorderRight(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(bColor).
		Foreground(dText).
		Render(content)

	return lipgloss.JoinVertical(lipgloss.Left, topLine, box)
}

func (d *Dashboard) renderBuilds(active []*core.Build, w int) string {
	if len(active) == 0 {
		return lipgloss.NewStyle().Foreground(dFaint).Render("\n  No active builds\n  Launcher [2] → start one")
	}
	var rows []string
	for i, b := range active {
		filled := int(b.Progress * float64(w-4))
		if filled > w-4 {
			filled = w - 4
		}
		if filled < 0 {
			filled = 0
		}
		empty := (w - 4) - filled

		stateColor := dSky
		icon := "⟳"
		switch b.State {
		case core.StateSuccess:
			stateColor = dGreen; icon = "✔"
		case core.StateFailed:
			stateColor = dRed; icon = "✖"
		case core.StateQueued:
			stateColor = dMauve; icon = "◌"
		}

		name := truncate(b.Name, 20)
		bar := lipgloss.NewStyle().Foreground(stateColor).Render(strings.Repeat("█", filled)) +
			lipgloss.NewStyle().Foreground(dSurface).Render(strings.Repeat("░", empty))
		pct := int(b.Progress * 100)
		elapsed := int(b.Elapsed().Seconds())

		prefix := "  "
		nameColor := dText
		if i == d.selectedBuild && d.active == 0 {
			prefix = lipgloss.NewStyle().Foreground(dGreen).Bold(true).Render("▸ ")
			nameColor = dGreen
		} else if i == d.selectedBuild {
			prefix = lipgloss.NewStyle().Foreground(dFaint).Bold(true).Render("▸ ")
		}

		rows = append(rows,
			fmt.Sprintf("%s%s %s %s", prefix, lipgloss.NewStyle().Foreground(stateColor).Render(icon),
				lipgloss.NewStyle().Foreground(nameColor).Bold(true).Render(name),
				lipgloss.NewStyle().Foreground(dBlue).Render(b.Tool.String())),
			fmt.Sprintf("  %s %s", bar, lipgloss.NewStyle().Foreground(dFaint).Render(fmt.Sprintf("%3d%% %ds", pct, elapsed))),
			"",
		)
	}
	return strings.Join(rows, "\n")
}

func (d *Dashboard) renderErrors(errors []core.BuildError, w int) string {
	if len(errors) == 0 {
		return lipgloss.NewStyle().Foreground(dGreen).Render("\n  No errors ✔")
	}
	var rows []string
	for _, e := range errors {
		c := dRed
		if e.Level == core.LogWarning {
			c = dYellow
		}
		dot := lipgloss.NewStyle().Foreground(c).Render("●")
		loc := lipgloss.NewStyle().Foreground(dSubtext).Render(fmt.Sprintf("%s:%d", truncate(e.File, 16), e.Line))
		msg := lipgloss.NewStyle().Foreground(dText).Render(truncate(e.Message, w-24))
		rows = append(rows, fmt.Sprintf("  %s %s %s", dot, loc, msg))
	}
	return strings.Join(rows, "\n")
}

func (d *Dashboard) renderLogs(w int, logLines []core.LogLine) string {
	if len(logLines) == 0 {
		return lipgloss.NewStyle().Foreground(dFaint).Render("\n  Waiting for output...")
	}

	var lines []string
	for _, ll := range logLines {
		c := dSubtext
		switch ll.Level {
		case core.LogError:
			c = dRed
		case core.LogWarning:
			c = dYellow
		case core.LogNote:
			c = dMauve
		}
		ts := lipgloss.NewStyle().Foreground(dFaint).Render(ll.Timestamp.Format("15:04:05") + " ")
		line := ts + lipgloss.NewStyle().Foreground(c).Render(truncate(ll.Raw, w-10))
		lines = append(lines, line)
	}

	// Tail the logs to fit the box height (approx 100 lines max)
	if len(lines) > 100 {
		lines = lines[len(lines)-100:]
	}
	return strings.Join(lines, "\n")
}

func (d *Dashboard) renderStats() string {
	all := d.bm.All()
	total := len(all)
	failed := 0
	for _, b := range all {
		if b.State == core.StateFailed {
			failed++
		}
	}

	row := func(label, val string, c color.Color) string {
		l := lipgloss.NewStyle().Foreground(dSubtext).Width(18).Render("  " + label)
		v := lipgloss.NewStyle().Foreground(c).Bold(true).Render(val)
		return l + v
	}

	successStr := "—"
	if total > 0 {
		successStr = fmt.Sprintf("%.0f%%", float64(total-failed)/float64(total)*100)
	}

	return strings.Join([]string{
		"",
		row("Total builds", fmt.Sprintf("%d", total), dText),
		row("Failed", fmt.Sprintf("%d", failed), dRed),
		row("Success rate", successStr, dGreen),
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
