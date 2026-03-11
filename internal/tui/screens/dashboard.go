package screens

import (
	"fmt"
	"strings"

	"github.com/MemestaVedas/gobuild/internal/core"
	"github.com/MemestaVedas/gobuild/internal/tui/theme"
	tea "github.com/charmbracelet/bubbletea"
	"charm.land/lipgloss/v2"
)

type Dashboard struct {
	bm            *core.BuildManager
	width, height int
	active        int
	selectedBuild int
	styles        theme.Styles
}

func NewDashboard(bm *core.BuildManager, styles theme.Styles) *Dashboard {
	return &Dashboard{
		bm:     bm,
		styles: styles,
	}
}

func (d *Dashboard) Init() tea.Cmd { return nil }

func (d *Dashboard) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		d.width = msg.Width
		d.height = msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "tab":
			d.active = (d.active + 1) % 4
		case "up", "k":
			if d.active == 0 && d.selectedBuild > 0 {
				d.selectedBuild--
			}
		case "down", "j":
			if d.active == 0 && len(d.bm.Active()) > 0 && d.selectedBuild < len(d.bm.Active())-1 {
				d.selectedBuild++
			}
		}
	}
	return d, nil
}

func (d *Dashboard) View() string {
	if d.width == 0 || d.height == 0 {
		return "Loading..."
	}

	leftW := (d.width * 2) / 5
	rightW := d.width - leftW

	activeBuilds := d.bm.Active()
	errors := d.bm.AllErrors()

	leftTop := d.panel(fmt.Sprintf("Active (%d)", len(activeBuilds)), leftW, d.height/2, d.active == 0,
		d.renderBuilds(activeBuilds, leftW-2))
	leftBot := d.panel(fmt.Sprintf("Errors (%d)", len(errors)), leftW, d.height-d.height/2, d.active == 2,
		d.renderErrors(errors, leftW-2))

	var selectedBuild *core.Build
	var selectedLogLines []core.LogLine
	if len(activeBuilds) > 0 {
		if d.selectedBuild >= len(activeBuilds) {
			d.selectedBuild = len(activeBuilds) - 1
		}
		selectedBuild = activeBuilds[d.selectedBuild]
		selectedLogLines = selectedBuild.LogLines
	}

	rightTop := d.panel("Log Output", rightW, d.height/2, d.active == 1,
		d.renderLogs(rightW-2, selectedLogLines))
	rightBot := d.panel("Details", rightW, d.height-d.height/2, d.active == 3,
		d.renderDetails(selectedBuild, rightW-2))

	left := lipgloss.JoinVertical(lipgloss.Left, leftTop, leftBot)
	right := lipgloss.JoinVertical(lipgloss.Left, rightTop, rightBot)

	return lipgloss.JoinHorizontal(lipgloss.Top, left, right)
}

func (d *Dashboard) panel(title string, w, h int, active bool, content string) string {
	bColor := d.styles.ColorBorderInactive
	if active {
		bColor = d.styles.ColorBorderActive
	}

	titleStyled := lipgloss.NewStyle().
		Foreground(d.styles.ColorText).
		Bold(active).
		Render(" " + title + " ")

	titleBarWidth := w - 2
	if titleBarWidth < 0 {
		titleBarWidth = 0
	}

	topLine := lipgloss.NewStyle().
		Foreground(bColor).
		Render("╭") +
		titleStyled +
		lipgloss.NewStyle().Foreground(bColor).Render(strings.Repeat("─", titleBarWidth-lipgloss.Width(titleStyled))) +
		lipgloss.NewStyle().Foreground(bColor).Render("╮")

	box := lipgloss.NewStyle().
		Width(w - 2).
		MaxWidth(w - 2).
		Height(h - 2).
		MaxHeight(h - 2).
		BorderLeft(true).
		BorderBottom(true).
		BorderRight(true).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(bColor).
		Foreground(d.styles.ColorText).
		Render(content)

	return lipgloss.JoinVertical(lipgloss.Left, topLine, box)
}

func (d *Dashboard) renderDetails(b *core.Build, w int) string {
	if b == nil {
		return lipgloss.NewStyle().Foreground(d.styles.ColorFaint).Render("\n  Select a build to see details")
	}
	return lipgloss.NewStyle().Foreground(d.styles.ColorText).Render(fmt.Sprintf("\n  ID: %s\n  Tool: %s\n  Cmd: %s", b.ID, b.Tool.String(), b.Command))
}

func (d *Dashboard) renderBuilds(active []*core.Build, w int) string {
	if len(active) == 0 {
		return lipgloss.NewStyle().Foreground(d.styles.ColorFaint).Render("\n  No active builds\n  Launcher [2] → start one")
	}

	var rows []string
	for i, b := range active {
		barW := w - 18
		if barW < 4 {
			barW = 4
		}
		filled := int(b.Progress * float64(barW))
		if filled > barW {
			filled = barW
		}
		if filled < 0 {
			filled = 0
		}

		barColor := d.styles.ColorRunning
		bIcon := "󰦖"
		switch b.State {
		case core.StateSuccess:
			barColor = d.styles.ColorSuccess
			bIcon = b.StatusIcon()
		case core.StateFailed:
			barColor = d.styles.ColorFailed
			bIcon = b.StatusIcon()
		case core.StateQueued:
			barColor = d.styles.ColorQueued
			bIcon = "󱞙"
		}

		bar := lipgloss.NewStyle().Foreground(barColor).Render(strings.Repeat("█", filled)) +
			lipgloss.NewStyle().Foreground(d.styles.ColorSurface).Render(strings.Repeat("░", barW-filled))

		prefix := "  "
		if i == d.selectedBuild {
			prefix = lipgloss.NewStyle().Foreground(d.styles.ColorAccent).Render("➔ ")
		}

		label := lipgloss.NewStyle().Foreground(d.styles.ColorText).Render(truncate(b.ID, 16))
		pct := int(b.Progress * 100)
		elapsed := b.Elapsed().Seconds()

		statusIconStyled := lipgloss.NewStyle().
			Foreground(barColor).
			Render(bIcon)

		timeTextStyled := lipgloss.NewStyle().
			Foreground(d.styles.ColorFaint).
			Render(fmt.Sprintf("%ds", int(elapsed)))

		rows = append(rows,
			fmt.Sprintf("%s%s %s %s", prefix, statusIconStyled, label, lipgloss.NewStyle().Foreground(d.styles.ColorNormal).Render(b.Tool.String())),
			fmt.Sprintf("  %s %s", bar, lipgloss.NewStyle().Foreground(d.styles.ColorFaint).Render(fmt.Sprintf("%3d%% %s", pct, timeTextStyled))),
			"",
		)
	}
	return strings.Join(rows, "\n")
}

func (d *Dashboard) renderErrors(errors []core.BuildError, w int) string {
	if len(errors) == 0 {
		return lipgloss.NewStyle().Foreground(d.styles.ColorSuccess).Render("\n  No errors ✔")
	}
	var rows []string
	for _, e := range errors {
		c := d.styles.ColorFailed
		if e.Level == core.LogWarning {
			c = d.styles.ColorWarning
		}
		dot := lipgloss.NewStyle().Foreground(c).Render("●")
		loc := lipgloss.NewStyle().Foreground(d.styles.ColorSubtext).Render(fmt.Sprintf("%s:%d", truncate(e.File, 16), e.Line))
		msg := lipgloss.NewStyle().Foreground(d.styles.ColorText).Render(truncate(e.Message, w-24))
		rows = append(rows, fmt.Sprintf("  %s %s %s", dot, loc, msg))
	}
	return strings.Join(rows, "\n")
}

func (d *Dashboard) renderLogs(w int, logLines []core.LogLine) string {
	if len(logLines) == 0 {
		return lipgloss.NewStyle().Foreground(d.styles.ColorFaint).Render("\n  Waiting for output...")
	}

	var lines []string
	for _, ll := range logLines {
		c := d.styles.ColorSubtext
		switch ll.Level {
		case core.LogError:
			c = d.styles.ColorFailed
		case core.LogWarning:
			c = d.styles.ColorWarning
		case core.LogNote:
			c = d.styles.ColorQueued
		}
		ts := lipgloss.NewStyle().Foreground(d.styles.ColorFaint).Render(ll.Timestamp.Format("15:04:05") + " ")
		line := ts + lipgloss.NewStyle().Foreground(c).Render(truncate(ll.Raw, w-10))
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}

func (d *Dashboard) Focus() { d.active = 0 }
func (d *Dashboard) Blur()  { d.active = -1 }

func truncate(s string, l int) string {
	if len(s) <= l {
		return s
	}
	return s[:l-1] + "…"
}
