package screens

import (
	"fmt"
	"strings"
	"time"

	"github.com/MemestaVedas/gobuild/internal/builder"
	"github.com/MemestaVedas/gobuild/internal/core"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Launcher struct {
	width   int
	height  int
	inputs  []textinput.Model
	focused int
	bm      *core.BuildManager
	bldr    *builder.Builder
}

func NewLauncher(bm *core.BuildManager, bldr *builder.Builder) *Launcher {
	l := &Launcher{
		inputs: make([]textinput.Model, 3),
		bm:     bm,
		bldr:   bldr,
	}

	for i := range l.inputs {
		t := textinput.New()
		t.Cursor.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#A6E3A1"))
		t.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#89B4FA"))
		t.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#CDD6F4"))
		t.ShowSuggestions = true
		switch i {
		case 0:
			t.Prompt = "  Directory:  "
			t.Placeholder = "/home/user/projects/myapp"
		case 1:
			t.Prompt = "  Command:    "
			t.Placeholder = "cargo build --release"
		case 2:
			t.Prompt = "  Tags:       "
			t.Placeholder = "#release"
		}
		l.inputs[i] = t
	}

	return l
}

func (l *Launcher) Init() tea.Cmd {
	return textinput.Blink
}

func (l *Launcher) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		l.width = msg.Width
		l.height = msg.Height - 3
	case tea.KeyMsg:
		switch msg.String() {
		case "up":
			l.inputs[l.focused].Blur()
			l.focused--
			if l.focused < 0 {
				l.focused = len(l.inputs) - 1
			}
			l.inputs[l.focused].Focus()
		case "down":
			l.inputs[l.focused].Blur()
			l.focused = (l.focused + 1) % len(l.inputs)
			l.inputs[l.focused].Focus()
		case "enter":
			if l.focused == len(l.inputs)-1 {
				// Execute Launch
				dir := l.inputs[0].Value()
				cmdStr := l.inputs[1].Value()
				if cmdStr != "" {
					b := &core.Build{
						ID:        fmt.Sprintf("%d", time.Now().UnixNano()),
						Name:      cmdStr,
						Command:   cmdStr,
						WorkDir:   dir,
						Tool:      core.ToolGeneric, // Logic for auto-detecting later
						StartTime: time.Now(),
					}
					l.bm.Add(b)
					l.bldr.StartBuild(b)
					// Switch away from launcher
					return l, func() tea.Msg {
						return core.SwitchToDashboardMsg{} // We need to import core or return the app's msg
					}
				}
				return l, nil
			}
			l.inputs[l.focused].Blur()
			l.focused++
			l.inputs[l.focused].Focus()
		}
	}

	// Route to textinput
	for i := range l.inputs {
		var cmd tea.Cmd
		l.inputs[i], cmd = l.inputs[i].Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	return l, tea.Batch(cmds...)
}

func (l *Launcher) View() string {
	titleColor := lipgloss.Color("#CBA6F7")
	title := "  NEW BUILD "

	titleRow := lipgloss.NewStyle().Foreground(titleColor).Bold(true).Render(title)

	var formRows []string
	formRows = append(formRows, "")
	for i := range l.inputs {
		formRows = append(formRows, l.inputs[i].View())
		formRows = append(formRows, "")
	}

	formRows = append(formRows, lipgloss.NewStyle().Foreground(lipgloss.Color("#A6E3A1")).Render("  Detected:   🔧 Unknown tool"))
	formRows = append(formRows, "")
	formRows = append(formRows, "  Profiles:   [frontend]  [backend]  [+ Save]")
	formRows = append(formRows, "")

	buttons := lipgloss.JoinHorizontal(lipgloss.Center,
		lipgloss.NewStyle().Foreground(lipgloss.Color("#A6E3A1")).Padding(0, 2).Render("[ ▶ Run Build ]"),
		lipgloss.NewStyle().Foreground(lipgloss.Color("#F38BA8")).Padding(0, 2).Render("[ ✕ Cancel ]"),
	)
	formRows = append(formRows, "  "+buttons)

	content := strings.Join(formRows, "\n")
	contentStyle := lipgloss.NewStyle().
		Width(l.width).
		Height(l.height)

	return lipgloss.JoinVertical(lipgloss.Left, titleRow, contentStyle.Render(content))
}

func (l *Launcher) Focus() {
	if len(l.inputs) > 0 {
		l.inputs[l.focused].Focus()
	}
}

func (l *Launcher) Blur() {
	for i := range l.inputs {
		l.inputs[i].Blur()
	}
}
