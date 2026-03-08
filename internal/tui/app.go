package tui

import (
	// Added fmt import
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/MemestaVedas/gobuild/internal/builder"
	"github.com/MemestaVedas/gobuild/internal/core"
	"github.com/MemestaVedas/gobuild/internal/tui/screens"
)

// Messages used conceptually in the architecture.
type BuildUpdateMsg struct{}
type StatsUpdateMsg struct{}
type TickMsg struct{}

func tick() tea.Cmd {
	return tea.Tick(time.Millisecond*250, func(t time.Time) tea.Msg {
		return TickMsg{}
	})
}

// AppModel is the root component of the TUI.
type AppModel struct {
	mode       Mode
	activeTab  int
	width      int
	height     int
	styles     Styles
	keys       KeyMap
	statusBar  StatusBarModel
	tabs       TabBarModel
	screens    []screens.Screen
	commandBuf string // current command input when in ModeCommand

	// Services
	bm   *core.BuildManager
	bldr *builder.Builder
}

// NewAppModel creates the root TUI model.
func NewAppModel(bm *core.BuildManager, bldr *builder.Builder) *AppModel {
	styles := DefaultStyles()
	m := &AppModel{
		mode:      ModeNormal,
		activeTab: 0,
		styles:    styles,
		keys:      DefaultKeyMap(),
		statusBar: NewStatusBarModel(styles),
		tabs:      NewTabBarModel(styles),
		bm:        bm,
		bldr:      bldr,
		screens: []screens.Screen{
			screens.NewDashboard(bm),
			screens.NewLauncher(bm, bldr),
			screens.NewHistory(bm),
			screens.NewPlugins(),
			screens.NewHelp(),
		},
	}
	return m
}

func (m *AppModel) Init() tea.Cmd {
	var cmds []tea.Cmd
	for _, s := range m.screens {
		cmds = append(cmds, s.Init())
	}
	cmds = append(cmds, tick())
	return tea.Batch(cmds...)
}

func (m *AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case TickMsg:
		return m, tick()
	case tea.KeyMsg:
		return m.handleKey(msg)
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		// The workspace height = total - top(1) - tabs(1) - sep1(1) - sep2(1) - status(1) - sep3(1) - hints(1) - bot(1)
		wsHeight := m.height - 8
		if wsHeight < 0 {
			wsHeight = 0
		}
		wsWidth := m.width - 2
		if wsWidth < 0 {
			wsWidth = 0
		}

		// Broadcast the correctly reduced inner size to all child screens
		screenMsg := tea.WindowSizeMsg{Width: wsWidth, Height: wsHeight}
		for i := range m.screens {
			mod, _ := m.screens[i].Update(screenMsg)
			m.screens[i] = mod.(screens.Screen)
		}
		return m, nil
	case BuildUpdateMsg:
		// Send to current screen if interested
		return m, nil
	case StatsUpdateMsg:
		// Re-trigger stats poll or just update status bar
		return m, nil
	case core.SwitchToDashboardMsg:
		m.activeTab = 0
		m.tabs.SetActive(0)
		m.mode = ModeNormal
		return m, nil
	}

	// Always forward uncaught messages to the active screen
	// They might be internal Bubble Tea msgs like blink commands
	var cmd tea.Cmd
	var mod tea.Model
	mod, cmd = m.screens[m.activeTab].Update(msg)
	m.screens[m.activeTab] = mod.(screens.Screen)

	// Sync status bar
	m.statusBar.SetActiveTab(m.activeTab)
	totalErrs, totalWarns := 0, 0
	for _, b := range m.bm.All() {
		for _, e := range b.Errors {
			if e.Level == core.LogError {
				totalErrs++
			} else {
				totalWarns++
			}
		}
	}
	m.statusBar.SetErrors(totalErrs, totalWarns)

	return m, cmd
}

func (m *AppModel) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// 1. Hard global override: Esc always returns to Normal from anywhere
	if key.Matches(msg, m.keys.Esc) {
		m.mode = ModeNormal
		m.commandBuf = ""
		// The active screen might need to clear selection in Normal too, so dispatch it.
		// Note: The specs specify "In Normal mode, Esc clears any active selection".
		m.screens[m.activeTab].Blur()
		mod, cmd := m.screens[m.activeTab].Update(msg)
		m.screens[m.activeTab] = mod.(screens.Screen)
		return m, cmd
	}

	// Route based on current mode
	switch m.mode {
	case ModeNormal:
		return m.handleNormalKey(msg)
	case ModeInsert:
		return m.handleInsertKey(msg)
	case ModeCommand:
		return m.handleCommandKey(msg)
	}

	return m, nil
}

func (m *AppModel) handleNormalKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Global Mode shifts
	if key.Matches(msg, m.keys.Command) {
		m.mode = ModeCommand
		return m, nil
	}
	// Note: 'i' or 'Enter' normally transition to Insert, but typically
	// only if a specific panel (like an input field on Planner) handles it.
	// We'll let the active screen intercept `Enter` and tell the parent,
	// or we just dispatch. For now, directly check tab switching and quit.

	if key.Matches(msg, m.keys.Quit) {
		return m, tea.Quit
	}

	// Tier 1 - Tab Switching
	if key.Matches(msg, m.keys.Tab1) {
		m.activeTab = 0
		m.tabs.SetActive(0)
		return m, nil
	}
	if key.Matches(msg, m.keys.Tab2) {
		m.activeTab = 1
		m.tabs.SetActive(1)
		return m, nil
	}
	if key.Matches(msg, m.keys.Tab3) {
		m.activeTab = 2
		m.tabs.SetActive(2)
		return m, nil
	}
	if key.Matches(msg, m.keys.Tab4) {
		m.activeTab = 3
		m.tabs.SetActive(3)
		return m, nil
	}
	if key.Matches(msg, m.keys.Help, m.keys.Tab5) {
		m.activeTab = 4
		m.tabs.SetActive(4)
		return m, nil
	}

	// Entering Insert Mode
	if msg.String() == "i" || msg.String() == "a" || msg.String() == "enter" {
		m.mode = ModeInsert
		m.screens[m.activeTab].Focus()
		// Let the screen process the key if necessary (e.g., enter to select an input)
		// But usually we just consume it to enter the mode
		if msg.String() == "enter" {
			mod, cmd := m.screens[m.activeTab].Update(msg)
			m.screens[m.activeTab] = mod.(screens.Screen)
			return m, cmd
		}
		return m, nil
	}

	// Dispatch other Normal keys to the active screen
	mod, cmd := m.screens[m.activeTab].Update(msg)
	m.screens[m.activeTab] = mod.(screens.Screen)
	return m, cmd
}

func (m *AppModel) handleInsertKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Entirely driven by active screen (e.g. Launcher inputs)
	mod, cmd := m.screens[m.activeTab].Update(msg)
	m.screens[m.activeTab] = mod.(screens.Screen)
	return m, cmd
}

func (m *AppModel) handleCommandKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if key.Matches(msg, m.keys.Enter) {
		// Execute commandBuf and return to normal mode
		return m.executeCommand(m.commandBuf)
	}
	if msg.Type == tea.KeyRunes || msg.Type == tea.KeySpace {
		m.commandBuf += string(msg.Runes)
	} else if msg.Type == tea.KeyBackspace {
		if len(m.commandBuf) > 0 {
			m.commandBuf = m.commandBuf[:len(m.commandBuf)-1]
		}
	}
	return m, nil
}

func (m *AppModel) executeCommand(cmdStr string) (tea.Model, tea.Cmd) {
	// Stub for executing :run, :kill, :q commands
	if strings.TrimSpace(cmdStr) == "q" || strings.TrimSpace(cmdStr) == "quit" {
		return m, tea.Quit
	}
	m.mode = ModeNormal
	m.commandBuf = ""
	return m, nil
}

func safeRepeat(s string, count int) string {
	if count <= 0 {
		return ""
	}
	return strings.Repeat(s, count)
}

func (m *AppModel) View() string {
	if m.width <= 0 || m.height <= 0 {
		return "Initializing..."
	}

	innerW := m.width - 2
	if innerW < 0 {
		innerW = 0
	}
	wsHeight := m.height - 8
	if wsHeight < 0 {
		wsHeight = 0
	}

	top := "┌" + safeRepeat("─", innerW) + "┐"

	// Ensure these components render to exact width
	tabsContent := lipgloss.PlaceHorizontal(innerW, lipgloss.Left, m.tabs.View(innerW))
	tabs := "│" + tabsContent + "│"

	sep1 := "├" + safeRepeat("─", innerW) + "┤"

	activeScreen := m.screens[m.activeTab].View()
	workspaceLines := strings.Split(activeScreen, "\n")

	var wsRendered []string
	for i := 0; i < wsHeight; i++ {
		line := ""
		if i < len(workspaceLines) {
			line = workspaceLines[i]
		}
		renderedLine := lipgloss.PlaceHorizontal(innerW, lipgloss.Left, line)
		wsRendered = append(wsRendered, "│"+renderedLine+"│")
	}

	sep2 := "├" + safeRepeat("─", innerW) + "┤"

	statusContent := lipgloss.PlaceHorizontal(innerW, lipgloss.Left, m.statusBar.View(m.mode, innerW))
	status := "│" + statusContent + "│"

	bot := "└" + safeRepeat("─", innerW) + "┘"

	return lipgloss.JoinVertical(lipgloss.Left,
		lipgloss.NewStyle().Foreground(m.styles.ColorBorderInactive).Render(top),
		tabs,
		lipgloss.NewStyle().Foreground(m.styles.ColorBorderInactive).Render(sep1),
		strings.Join(wsRendered, "\n"),
		lipgloss.NewStyle().Foreground(m.styles.ColorBorderInactive).Render(sep2),
		status,
		lipgloss.NewStyle().Foreground(m.styles.ColorBorderInactive).Render(bot),
	)
}

func (m *AppModel) hintsBar(width int) string {
	var content string
	if m.mode == ModeCommand {
		content = ":" + m.commandBuf + "_"
	} else {
		content = "  r Run   x Kill   f Focus Log   o Open Editor   / Search"
		if m.activeTab == 1 { // Launcher
			content = "  Tab Complete   Esc Normal Mode   Up/Down Move Field"
		} else if m.activeTab == 2 { // History
			content = "  Enter View Logs  o Open  r Re-run  / Search  Del Delete"
		}
	}
	return m.styles.HintsBar.Copy().Width(width).Render(content)
}

// Global ticks
func tickStatsUpdate() tea.Cmd {
	return tea.Tick(time.Second*1, func(time.Time) tea.Msg {
		return StatsUpdateMsg{}
	})
}
