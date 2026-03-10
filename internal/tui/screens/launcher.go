package screens

import (
	"bufio"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"image/color"

	"github.com/MemestaVedas/gobuild/internal/builder"
	"github.com/MemestaVedas/gobuild/internal/core"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	lipglossV1 "github.com/charmbracelet/lipgloss"
	"charm.land/lipgloss/v2"
)

// ── Palette ─────────────────────────────────────────────────────────────────
var (
	lGreen   = lipgloss.Color("#A6E3A1")
	lRed     = lipgloss.Color("#F38BA8")
	lBlue    = lipgloss.Color("#89B4FA")
	lMauve   = lipgloss.Color("#CBA6F7")
	lText    = lipgloss.Color("#CDD6F4")
	lSubtext = lipgloss.Color("#A6ADC8")
	lFaint   = lipgloss.Color("#585B70")
	lSurface = lipgloss.Color("#313244")
	lBase    = lipgloss.Color("#1E1E2E")
	lOverlay = lipgloss.Color("#45475A")
	lCrust   = lipgloss.Color("#11111B")
)

// ── Common build commands ───────────────────────────────────────────────────
var commonCommands = []string{
	"go build ./...",
	"go build -o bin/app ./cmd/app",
	"go test ./...",
	"go run .",
	"go run cmd/gobuild/main.go",
	"cargo build",
	"cargo build --release",
	"cargo test",
	"cargo run",
	"npm run build",
	"npm run dev",
	"npm install",
	"npm test",
	"yarn build",
	"make",
	"make all",
	"make clean",
	"make install",
	"cmake --build build",
	"gradle build",
	"./gradlew build",
	"./gradlew assembleDebug",
	"python setup.py build",
	"docker build -t app .",
	"adb install -r BuildM-ON-android/app/build/outputs/apk/debug/app-debug.apk",
}

// ── Shell history ───────────────────────────────────────────────────────────

func readShellHistory(maxN int) []string {
	candidates := []string{os.Getenv("HISTFILE")}
	if u, err := user.Current(); err == nil {
		candidates = append(candidates,
			filepath.Join(u.HomeDir, ".zsh_history"),
			filepath.Join(u.HomeDir, ".bash_history"),
			filepath.Join(u.HomeDir, ".local/share/fish/fish_history"),
		)
	}

	buildKW := []string{
		"go build", "go run", "go test", "go install",
		"cargo", "npm run", "npm build", "yarn", "make", "cmake",
		"gradle", "gradlew", "docker build", "python setup",
	}

	seen := make(map[string]bool)
	var results []string

	for _, hf := range candidates {
		if hf == "" {
			continue
		}
		f, err := os.Open(hf)
		if err != nil {
			continue
		}
		defer f.Close()

		var lines []string
		sc := bufio.NewScanner(f)
		for sc.Scan() {
			line := sc.Text()
			if idx := strings.LastIndex(line, ";"); idx != -1 && strings.HasPrefix(line, ":") {
				line = line[idx+1:]
			}
			line = strings.TrimSpace(line)
			if line == "" || seen[line] {
				continue
			}
			for _, kw := range buildKW {
				if strings.Contains(line, kw) {
					lines = append(lines, line)
					seen[line] = true
					break
				}
			}
		}
		for i := len(lines) - 1; i >= 0 && len(results) < maxN; i-- {
			results = append(results, lines[i])
		}
	}
	return results
}

func dirSuggestions(partial string) []string {
	if partial == "" {
		home, _ := os.UserHomeDir()
		partial = home
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
		return nil
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
		if len(results) >= 8 {
			break
		}
	}
	return results
}

func cmdSuggestions(partial string, history []string) []string {
	seen := make(map[string]bool)
	var results []string

	for _, h := range history {
		if strings.Contains(strings.ToLower(h), strings.ToLower(partial)) && !seen[h] {
			results = append(results, h)
			seen[h] = true
		}
	}
	for _, c := range commonCommands {
		if strings.Contains(strings.ToLower(c), strings.ToLower(partial)) && !seen[c] {
			results = append(results, c)
			seen[c] = true
		}
	}

	sort.SliceStable(results, func(i, j int) bool { return len(results[i]) < len(results[j]) })
	if len(results) > 8 {
		results = results[:8]
	}
	return results
}

// ── Model ───────────────────────────────────────────────────────────────────

const (
	fieldDir = iota
	fieldCmd
	fieldTags
	fieldCount
)

type Launcher struct {
	width   int
	height  int
	inputs  [fieldCount]textinput.Model
	focused int
	bm      *core.BuildManager
	bldr    *builder.Builder
	history []string

	dirSugs      []string
	cmdSugs      []string
	sugIdx       int
	detectedTool string
	statusMsg    string
}

func NewLauncher(bm *core.BuildManager, bldr *builder.Builder) *Launcher {
	l := &Launcher{
		bm:      bm,
		bldr:    bldr,
		history: readShellHistory(50),
	}

	mk := func() textinput.Model {
		t := textinput.New()
		
		// Map v1 colors specifically for textinput
		v1Green := lipglossV1.Color("#A6E3A1")
		v1Text := lipglossV1.Color("#CDD6F4")
		v1Faint := lipglossV1.Color("#585B70")

		t.Cursor.Style = lipglossV1.NewStyle().Foreground(v1Green)
		t.PromptStyle = lipglossV1.NewStyle().Foreground(v1Green).Bold(true)
		t.TextStyle = lipglossV1.NewStyle().Foreground(v1Text)
		t.PlaceholderStyle = lipglossV1.NewStyle().Foreground(v1Faint)
		t.ShowSuggestions = true
		return t
	}

	l.inputs[fieldDir] = mk()
	l.inputs[fieldDir].Prompt = " Directory › "
	l.inputs[fieldDir].Placeholder = "/home/user/projects/myapp"
	l.inputs[fieldDir].Width = 55

	l.inputs[fieldCmd] = mk()
	l.inputs[fieldCmd].Prompt = " Command   › "
	l.inputs[fieldCmd].Placeholder = "go build ./..."
	l.inputs[fieldCmd].Width = 55

	l.inputs[fieldTags] = mk()
	l.inputs[fieldTags].Prompt = " Tags      › "
	l.inputs[fieldTags].Placeholder = "#release  #debug"
	l.inputs[fieldTags].Width = 55

	l.refreshDirSuggestions()
	l.refreshCmdSuggestions()

	return l
}

func (l *Launcher) refreshDirSuggestions() {
	l.dirSugs = dirSuggestions(l.inputs[fieldDir].Value())
	l.inputs[fieldDir].SetSuggestions(l.dirSugs)
	l.sugIdx = 0
}

func (l *Launcher) refreshCmdSuggestions() {
	l.cmdSugs = cmdSuggestions(l.inputs[fieldCmd].Value(), l.history)
	l.inputs[fieldCmd].SetSuggestions(l.cmdSugs)
	l.sugIdx = 0
}

func (l *Launcher) detectTool(dir string) {
	checks := []struct {
		file, tool string
	}{
		{"Cargo.toml", "🦀 Cargo"},
		{"package.json", "📦 NPM"},
		{"go.mod", "🐹 Go"},
		{"Makefile", "⚙️  Make"},
		{"CMakeLists.txt", "🔨 CMake"},
		{"build.gradle", "🐘 Gradle"},
		{"build.gradle.kts", "🐘 Gradle"},
	}
	l.detectedTool = ""
	for _, c := range checks {
		if _, err := os.Stat(filepath.Join(dir, c.file)); err == nil {
			l.detectedTool = c.tool
			return
		}
	}
}

func (l *Launcher) Init() tea.Cmd {
	l.inputs[fieldDir].Focus()
	return textinput.Blink
}

func (l *Launcher) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		l.width = msg.Width
		l.height = msg.Height
		for i := range l.inputs {
			l.inputs[i].Width = l.width - 25
			if l.inputs[i].Width < 20 {
				l.inputs[i].Width = 20
			}
		}

	case tea.KeyMsg:
		switch msg.String() {
		case "up":
			l.inputs[l.focused].Blur()
			l.focused = (l.focused - 1 + fieldCount) % fieldCount
			l.inputs[l.focused].Focus()
			l.sugIdx = 0

		case "down":
			l.inputs[l.focused].Blur()
			l.focused = (l.focused + 1) % fieldCount
			l.inputs[l.focused].Focus()
			l.sugIdx = 0

		case "tab":
			switch l.focused {
			case fieldDir:
				if len(l.dirSugs) > 0 {
					l.sugIdx = (l.sugIdx + 1) % len(l.dirSugs)
					l.inputs[fieldDir].SetValue(l.dirSugs[l.sugIdx])
					l.refreshDirSuggestions()
					l.detectTool(l.inputs[fieldDir].Value())
				} else {
					l.inputs[l.focused].Blur()
					l.focused = (l.focused + 1) % fieldCount
					l.inputs[l.focused].Focus()
				}
			case fieldCmd:
				if len(l.cmdSugs) > 0 {
					l.sugIdx = (l.sugIdx + 1) % len(l.cmdSugs)
					l.inputs[fieldCmd].SetValue(l.cmdSugs[l.sugIdx])
					l.refreshCmdSuggestions()
				} else {
					l.inputs[l.focused].Blur()
					l.focused = (l.focused + 1) % fieldCount
					l.inputs[l.focused].Focus()
				}
			default:
				l.inputs[l.focused].Blur()
				l.focused = (l.focused + 1) % fieldCount
				l.inputs[l.focused].Focus()
			}
			return l, nil

		case "shift+tab":
			switch l.focused {
			case fieldDir:
				if len(l.dirSugs) > 0 {
					l.sugIdx = (l.sugIdx - 1 + len(l.dirSugs)) % len(l.dirSugs)
					l.inputs[fieldDir].SetValue(l.dirSugs[l.sugIdx])
					l.detectTool(l.inputs[fieldDir].Value())
				}
			case fieldCmd:
				if len(l.cmdSugs) > 0 {
					l.sugIdx = (l.sugIdx - 1 + len(l.cmdSugs)) % len(l.cmdSugs)
					l.inputs[fieldCmd].SetValue(l.cmdSugs[l.sugIdx])
				}
			}
			return l, nil

		case "enter":
			if l.focused == fieldTags || l.focused == fieldCmd {
				dir := l.inputs[fieldDir].Value()
				cmdStr := l.inputs[fieldCmd].Value()
				if cmdStr != "" {
					tool := core.ToolGeneric
					switch {
					case strings.HasPrefix(cmdStr, "cargo"):
						tool = core.ToolCargo
					case strings.HasPrefix(cmdStr, "go "):
						tool = core.ToolGo
					case strings.HasPrefix(cmdStr, "npm"), strings.HasPrefix(cmdStr, "yarn"):
						tool = core.ToolNPM
					case strings.HasPrefix(cmdStr, "make"):
						tool = core.ToolMake
					case strings.HasPrefix(cmdStr, "gradle"), strings.HasPrefix(cmdStr, "./gradlew"):
						tool = core.ToolGradle
					case strings.HasPrefix(cmdStr, "cmake"):
						tool = core.ToolCMake
					}
					b := &core.Build{
						ID:        fmt.Sprintf("%d", time.Now().UnixNano()),
						Name:      filepath.Base(cmdStr),
						Command:   cmdStr,
						WorkDir:   dir,
						Tool:      tool,
						StartTime: time.Now(),
					}
					l.bm.Add(b)
					l.bldr.StartBuild(b)
					l.statusMsg = fmt.Sprintf("✔ Launched: %s", cmdStr)
					return l, func() tea.Msg { return core.SwitchToDashboardMsg{} }
				}
				return l, nil
			}
			l.inputs[l.focused].Blur()
			l.focused = (l.focused + 1) % fieldCount
			l.inputs[l.focused].Focus()
		}
	}

	for i := range l.inputs {
		var cmd tea.Cmd
		l.inputs[i], cmd = l.inputs[i].Update(msg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}

	if _, ok := msg.(tea.KeyMsg); ok {
		switch l.focused {
		case fieldDir:
			l.refreshDirSuggestions()
			l.detectTool(l.inputs[fieldDir].Value())
		case fieldCmd:
			l.refreshCmdSuggestions()
		}
	}

	return l, tea.Batch(cmds...)
}

// ── View ────────────────────────────────────────────────────────────────────

func (l *Launcher) View() string {
	if l.width == 0 {
		return ""
	}

	innerW := l.width - 4

	// ── Panel title
	title := lipgloss.NewStyle().Foreground(lGreen).Bold(true).Render("New Build")
	topLine := lipgloss.NewStyle().Padding(1, 1).Render(title)

	// ── Fields ────────────────────────────────────────────────────────
	var rows []string
	fieldLabels := []string{"Directory", "Command", "Tags"}

	for i := range l.inputs {
		focused := i == l.focused
		var accent color.Color
		if focused {
			accent = lGreen
		} else {
			accent = lFaint
		}

		// Field with left-bar accent
		line := lipgloss.NewStyle().
			PaddingLeft(1).
			BorderLeft(true).
			BorderStyle(lipgloss.ThickBorder()).
			BorderForeground(accent).
			Width(innerW - 2).
			Render(l.inputs[i].View())

		rows = append(rows, line)

		// Dropdown
		if focused && i == fieldDir && len(l.dirSugs) > 0 {
			rows = append(rows, l.renderDropdown(l.dirSugs, innerW, fieldLabels[i]))
		}
		if focused && i == fieldCmd && len(l.cmdSugs) > 0 {
			rows = append(rows, l.renderDropdown(l.cmdSugs, innerW, fieldLabels[i]))
		}
		rows = append(rows, "")
	}

	// ── Detected tool ─────────────────────────────────────────────────
	if l.detectedTool != "" {
		rows = append(rows,
			lipgloss.NewStyle().Foreground(lGreen).PaddingLeft(2).Render(
				"Detected: "+l.detectedTool), "")
	}

	// ── Status msg ────────────────────────────────────────────────────
	if l.statusMsg != "" {
		rows = append(rows,
			lipgloss.NewStyle().Foreground(lGreen).PaddingLeft(2).Render(l.statusMsg), "")
	}

	// ── Buttons ───────────────────────────────────────────────────────
	runBtn := lipgloss.NewStyle().
		Foreground(lCrust).
		Background(lGreen).
		Bold(true).
		Padding(0, 2).
		Render("▶ Run Build")

	cancelBtn := lipgloss.NewStyle().
		Foreground(lRed).
		Bold(true).
		Padding(0, 2).
		Render("✕ Cancel")

	rows = append(rows, "  "+runBtn+"   "+cancelBtn)

	// ── Assemble ──────────────────────────────────────────────────────
	body := strings.Join(rows, "\n")

	contentStyle := lipgloss.NewStyle().
		Width(innerW).
		Padding(1, 1)

	return lipgloss.JoinVertical(lipgloss.Left, topLine, contentStyle.Render(body))
}

func (l *Launcher) renderDropdown(items []string, width int, label string) string {
	maxShow := 5
	if len(items) < maxShow {
		maxShow = len(items)
	}

	var lines []string
	for i, item := range items[:maxShow] {
		if i == l.sugIdx%maxShow {
			lines = append(lines,
				lipgloss.NewStyle().
					Foreground(lGreen).Bold(true).
					Width(width-8).PaddingLeft(1).
					Render("▸ "+item))
		} else {
			lines = append(lines,
				lipgloss.NewStyle().
					Foreground(lSubtext).
					Width(width-8).PaddingLeft(1).
					Render("  "+item))
		}
	}

	return lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lOverlay).
		MarginLeft(3).
		Render(strings.Join(lines, "\n"))
}

func (l *Launcher) Focus() { l.inputs[l.focused].Focus() }
func (l *Launcher) Blur() {
	for i := range l.inputs {
		l.inputs[i].Blur()
	}
}
