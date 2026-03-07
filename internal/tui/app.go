package tui

import (
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/MemestaVedas/gobuild/internal/tui/screens"
)

// Messages used conceptually in the architecture.
type BuildUpdateMsg struct{}
type StatsUpdateMsg struct{}

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
}

// NewAppModel creates the root TUI model.
func NewAppModel() *AppModel {
	styles := DefaultStyles()
	return &AppModel{
		mode:      ModeNormal,
		activeTab: 0,
		styles:    styles,
		keys:      DefaultKeyMap(),
		statusBar: NewStatusBarModel(styles),
		tabs:      NewTabBarModel(styles),
		screens: []screens.Screen{
			screens.NewDashboard(),
			screens.NewLauncher(),
			screens.NewHistory(),
			screens.NewPlugins(),
			screens.NewHelp(),
		},
	}
}

func (m *AppModel) Init() tea.Cmd {
	return tea.Batch(
		m.screens[0].Init(),
		m.screens[1].Init(),
		m.screens[2].Init(),
		m.screens[3].Init(),
		m.screens[4].Init(),
	)
}

func (m *AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKey(msg)
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case BuildUpdateMsg:
		// Send to current screen if interested
		return m, nil
	case StatsUpdateMsg:
		// Re-trigger stats poll or just update status bar
		return m, nil
	}

	// Always forward uncaught messages to the active screen
	// They might be internal Bubble Tea msgs like blink commands
	var cmd tea.Cmd
	var mod tea.Model
	mod, cmd = m.screens[m.activeTab].Update(msg)
	m.screens[m.activeTab] = mod.(screens.Screen)
	return m, cmd
}

func (m *AppModel) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// 1. Hard global override: Esc always returns to Normal from anywhere
	if key.Matches(msg, m.keys.Esc) {
		m.mode = ModeNormal
		m.commandBuf = ""
		// The active screen might need to clear selection in Normal too, so dispatch it.
		// Note: The specs specify "In Normal mode, Esc clears any active selection".
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
	} else if msg.Type == tea.KeyBackspace || msg.Type == tea.KeyBackspace2 {
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

func (m *AppModel) View() string {
	if m.width == 0 {
		return "Initializing..."
	}

	statusBar := m.statusBar.View(m.mode, m.width)
	tabs := m.tabs.View(m.width)
	hints := m.hintsBar()

	// Available workspace height
	wsHeight := m.height - lipgloss.Height(statusBar) - lipgloss.Height(tabs) - lipgloss.Height(hints)
	_ = wsHeight // Later we pass this down to screens so they size properly

	activeScreen := m.screens[m.activeTab].View()

	return lipgloss.JoinVertical(
		lipgloss.Left,
		tabs,
		activeScreen,
		statusBar,
		hints,
	)
}

func (m *AppModel) hintsBar() string {
	if m.mode == ModeCommand {
		return m.styles.HintsBar.Copy().Width(m.width).Render(":" + m.commandBuf + "_")
	}

	// Show context hints
	content := "r Run   x Kill   f Focus Log   o Open Editor   / Search"
	if m.activeTab == 1 { // Launcher
		content = "Tab Complete   Esc Normal Mode   Ctrl+W Delete Word   Up/Down Move Field"
	} else if m.activeTab == 2 { // History
		content = "Enter View Logs  o Open  r Re-run  / Search  Del Delete"
	}

	return m.styles.HintsBar.Copy().Width(m.width).Render(content)
}

// Global ticks
func tickStatsUpdate() tea.Cmd {
	return tea.Tick(time.Second*1, func(time.Time) tea.Msg {
		return StatsUpdateMsg{}
	})
}
