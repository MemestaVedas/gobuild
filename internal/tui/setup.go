package tui

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	lipglossV1 "github.com/charmbracelet/lipgloss"
	"charm.land/lipgloss/v2"

	"github.com/MemestaVedas/gobuild/internal/config"
)

// SetupDoneMsg is sent when the setup wizard is completed.
type SetupDoneMsg struct {
	Watch config.WatchConfig
}

// SetupPage drives which page of the wizard we are on.
type SetupPage int

const (
	setupPageDir SetupPage = iota
	setupPageCmds
	setupPageReview
)

// SetupModal is the first-run configuration wizard shown as a full-screen overlay.
type SetupModal struct {
	width  int
	height int
	page   SetupPage

	// Directory input
	dirInput textinput.Model

	// Command input (one at a time, added to list)
	cmdInput textinput.Model
	cmdList  []string

	// Accumulated watched directories
	dirs []config.WatchedDir

	// Dir suggestions for autocomplete
	dirSugs []string
	sugIdx  int

	err string
}

func NewSetupModal() *SetupModal {
	mk := func(prompt, placeholder string) textinput.Model {
		t := textinput.New()
		t.Prompt = " " + prompt + " › "
		t.Placeholder = placeholder
		t.Width = 50
		t.PromptStyle = lipglossV1.NewStyle().Foreground(lipglossV1.Color("#A6E3A1")).Bold(true)
		t.TextStyle = lipglossV1.NewStyle().Foreground(lipglossV1.Color("#CDD6F4"))
		t.PlaceholderStyle = lipglossV1.NewStyle().Foreground(lipglossV1.Color("#585B70"))
		return t
	}

	home, _ := os.UserHomeDir()
	dirInput := mk("Project Directory", home+"/projects/my-app")
	dirInput.Focus()

	cmdInput := mk("Watch Command", "npm run tauri dev")

	return &SetupModal{
		dirInput: dirInput,
		cmdInput: cmdInput,
	}
}

func (s *SetupModal) Init() tea.Cmd {
	s.refreshDirSugs()
	return textinput.Blink
}

func (s *SetupModal) refreshDirSugs() {
	partial := s.dirInput.Value()
	if partial == "" {
		if u, err := user.Current(); err == nil {
			partial = u.HomeDir
		}
	}
	if strings.HasPrefix(partial, "~/") {
		home, _ := os.UserHomeDir()
		partial = filepath.Join(home, partial[2:])
	}

	dir := partial
	prefix := ""
	if !strings.HasSuffix(partial, "/") {
		dir = filepath.Dir(partial)
		prefix = filepath.Base(partial)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		s.dirSugs = nil
		return
	}
	var results []string
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		if prefix != "" && !strings.HasPrefix(strings.ToLower(e.Name()), strings.ToLower(prefix)) {
			continue
		}
		results = append(results, filepath.Join(dir, e.Name()))
		if len(results) >= 6 {
			break
		}
	}
	s.dirSugs = results
}

func (s *SetupModal) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		s.width = msg.Width
		s.height = msg.Height
		s.dirInput.Width = s.width - 20
		s.cmdInput.Width = s.width - 20

	case tea.KeyMsg:
		s.err = ""
		switch msg.String() {
		case "ctrl+c":
			return s, tea.Quit

		case "tab":
			// Autocomplete directory
			if s.page == setupPageDir && len(s.dirSugs) > 0 {
				s.sugIdx = (s.sugIdx + 1) % len(s.dirSugs)
				completed := s.dirSugs[s.sugIdx]
				s.dirInput.SetValue(completed)
				// Force cursor to end of completed text
				s.dirInput.SetCursor(len([]rune(completed)))
				s.refreshDirSugs()
				return s, nil
			}

		case "enter":
			switch s.page {
			case setupPageDir:
				dir := strings.TrimSpace(s.dirInput.Value())
				if dir == "" {
					s.err = "Directory cannot be empty"
					return s, nil
				}
				if _, err := os.Stat(dir); err != nil {
					s.err = "Directory does not exist"
					return s, nil
				}
				// Move to commands page
				s.page = setupPageCmds
				s.cmdList = nil
				s.dirInput.Blur()
				s.cmdInput.Focus()
				s.cmdInput.SetValue("")

			case setupPageCmds:
				cmd := strings.TrimSpace(s.cmdInput.Value())
				if cmd == "" && len(s.cmdList) == 0 {
					s.err = "Add at least one command"
					return s, nil
				}
				if cmd != "" {
					s.cmdList = append(s.cmdList, cmd)
					s.cmdInput.SetValue("")
					return s, nil
				}
				// Empty enter = done adding commands, move to review
				s.dirs = append(s.dirs, config.WatchedDir{
					Path:     strings.TrimSpace(s.dirInput.Value()),
					Commands: s.cmdList,
				})
				s.page = setupPageReview

			case setupPageReview:
				// Save and dismiss wizard
				watchCfg := config.WatchConfig{Directories: s.dirs}
				config.EnsureConfigDir()
				cfg := &config.Config{Watch: watchCfg}
				config.SaveWatchConfig(cfg)
				return s, func() tea.Msg { return SetupDoneMsg{Watch: watchCfg} }
			}

		case "esc":
			switch s.page {
			case setupPageCmds:
				// Back to dir page
				s.page = setupPageDir
				s.cmdInput.Blur()
				s.dirInput.Focus()
			case setupPageReview:
				s.page = setupPageCmds
				s.cmdInput.Focus()
			}

		case "ctrl+a":
			// Add another directory from review page
			if s.page == setupPageReview {
				s.page = setupPageDir
				s.dirInput.SetValue("")
				s.dirInput.Focus()
				s.cmdInput.Blur()
				s.refreshDirSugs()
			}

		case "ctrl+d":
			// Remove last command
			if s.page == setupPageCmds && len(s.cmdList) > 0 {
				s.cmdList = s.cmdList[:len(s.cmdList)-1]
				return s, nil
			}
		}

		// Route key to active input
		var cmd tea.Cmd
		switch s.page {
		case setupPageDir:
			s.dirInput, cmd = s.dirInput.Update(msg)
			s.refreshDirSugs()
		case setupPageCmds:
			s.cmdInput, cmd = s.cmdInput.Update(msg)
		}
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	return s, tea.Batch(cmds...)
}

func (s *SetupModal) View() string {
	if s.width <= 0 {
		return ""
	}

	boxW := min(s.width-8, 72)
	title := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#A6E3A1")).
		Bold(true).
		Render("  goBuild — First Run Setup")

	var body strings.Builder

	switch s.page {
	case setupPageDir:
		s.renderPageDir(&body, boxW)
	case setupPageCmds:
		s.renderPageCmds(&body, boxW)
	case setupPageReview:
		s.renderPageReview(&body, boxW)
	}

	if s.err != "" {
		body.WriteString("\n  " + lipgloss.NewStyle().Foreground(lipgloss.Color("#F38BA8")).Render("⚠ "+s.err))
	}

	box := lipgloss.NewStyle().
		Width(boxW).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#A6E3A1")).
		Padding(1, 2).
		Render(lipgloss.JoinVertical(lipgloss.Left, title, "", body.String()))

	// Center on screen
	return lipgloss.Place(s.width, s.height, lipgloss.Center, lipgloss.Center, box)
}

func (s *SetupModal) renderPageDir(b *strings.Builder, w int) {
	faint := lipgloss.NewStyle().Foreground(lipgloss.Color("#585B70"))
	accent := lipgloss.NewStyle().Foreground(lipgloss.Color("#A6E3A1")).Bold(true)

	b.WriteString(accent.Render("Step 1 of 2 — Project Directory") + "\n\n")
	b.WriteString(faint.Render("  Which project directory do you want to monitor?\n\n"))
	b.WriteString(s.dirInput.View() + "\n")

	if len(s.dirSugs) > 0 {
		b.WriteString(faint.Render("\n  Suggestions (Tab to cycle):\n"))
		for i, sg := range s.dirSugs {
			prefix := "    "
			c := lipgloss.Color("#585B70")
			if i == s.sugIdx%len(s.dirSugs) {
				prefix = "  ▸ "
				c = lipgloss.Color("#A6E3A1")
			}
			b.WriteString(lipgloss.NewStyle().Foreground(c).Render(prefix+truncateS(sg, w-6)+"\n"))
		}
	}

	b.WriteString(faint.Render("\n  Enter — confirm    Tab — autocomplete"))
}

func (s *SetupModal) renderPageCmds(b *strings.Builder, w int) {
	faint := lipgloss.NewStyle().Foreground(lipgloss.Color("#585B70"))
	accent := lipgloss.NewStyle().Foreground(lipgloss.Color("#A6E3A1")).Bold(true)
	dir := lipgloss.NewStyle().Foreground(lipgloss.Color("#89B4FA")).Render(s.dirInput.Value())

	b.WriteString(accent.Render("Step 2 of 2 — Watch Commands") + "\n\n")
	b.WriteString(fmt.Sprintf("  Directory: %s\n\n", dir))

	if len(s.cmdList) > 0 {
		b.WriteString(faint.Render("  Commands to watch:\n"))
		for _, c := range s.cmdList {
			b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#A6E3A1")).Render("    ✔ "+c) + "\n")
		}
		b.WriteString("\n")
	}

	b.WriteString(s.cmdInput.View() + "\n")

	if len(s.cmdList) > 0 {
		b.WriteString(faint.Render("\n  Enter (empty) — done    Ctrl+D — remove last    Esc — back"))
	} else {
		b.WriteString(faint.Render("\n  Enter — add command    Esc — back"))
	}
}

func (s *SetupModal) renderPageReview(b *strings.Builder, w int) {
	faint := lipgloss.NewStyle().Foreground(lipgloss.Color("#585B70"))
	accent := lipgloss.NewStyle().Foreground(lipgloss.Color("#A6E3A1")).Bold(true)
	green := lipgloss.NewStyle().Foreground(lipgloss.Color("#A6E3A1"))

	b.WriteString(accent.Render("Review & Save") + "\n\n")

	for _, d := range s.dirs {
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#89B4FA")).Bold(true).Render("  📁 "+truncateS(d.Path, w-6)) + "\n")
		for _, c := range d.Commands {
			b.WriteString(green.Render("      ✔ "+c) + "\n")
		}
		b.WriteString("\n")
	}

	b.WriteString(faint.Render("  Ctrl+A — add another directory\n"))
	b.WriteString(accent.Render("  Enter — save and start goBuild"))
	b.WriteString(faint.Render("    Esc — back"))
}

func truncateS(s string, max int) string {
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	return "…" + string(r[len(r)-(max-1):])
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
