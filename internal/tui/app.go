package tui

import (
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"charm.land/lipgloss/v2"

	"github.com/MemestaVedas/gobuild/internal/builder"
	"github.com/MemestaVedas/gobuild/internal/config"
	"github.com/MemestaVedas/gobuild/internal/core"
	"github.com/MemestaVedas/gobuild/internal/tui/screens"
	"github.com/MemestaVedas/gobuild/internal/tui/theme"
)

// Messages
type BuildUpdateMsg struct{}
type StatsUpdateMsg struct{}
type TickMsg struct{}

func tick() tea.Cmd {
	return tea.Tick(time.Millisecond*250, func(t time.Time) tea.Msg {
		return TickMsg{}
	})
}

// AppModel is the root TUI component.
type AppModel struct {
	mode       Mode
	activeTab  int
	width      int
	height     int
	styles     theme.Styles
	keys       KeyMap
	statusBar  StatusBarModel
	tabs       TabBarModel
	screens    []screens.Screen
	commandBuf string

	// First-run setup wizard
	setup    *SetupModal
	showSetup bool

	bm   *core.BuildManager
	bldr *builder.Builder
}

func NewAppModel(bm *core.BuildManager, bldr *builder.Builder, isDark bool) *AppModel {
	styles := theme.DefaultStyles(isDark)

	// Detect first-run: watch.json missing or has no directories.
	cfg, _ := config.Load()
	needsSetup := cfg == nil || len(cfg.Watch.Directories) == 0

	m := &AppModel{
		mode:      ModeNormal,
		activeTab: 0,
		styles:    styles,
		keys:      DefaultKeyMap(),
		statusBar: NewStatusBarModel(styles),
		tabs:      NewTabBarModel(styles),
		bm:        bm,
		bldr:      bldr,
		showSetup: needsSetup,
		screens: []screens.Screen{
			screens.NewDashboard(bm, styles),
			screens.NewLauncher(bm, bldr, styles),
			screens.NewHistory(bm, styles),
			screens.NewPlugins(styles),
			screens.NewHelp(styles),
		},
	}
	if needsSetup {
		m.setup = NewSetupModal()
	}
	return m
}

func (m *AppModel) Init() tea.Cmd {
	var cmds []tea.Cmd
	for _, s := range m.screens {
		cmds = append(cmds, s.Init())
	}
	cmds = append(cmds, tick())
	if m.showSetup && m.setup != nil {
		cmds = append(cmds, m.setup.Init())
	}
	return tea.Batch(cmds...)
}

func (m *AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.showSetup && m.setup != nil {
		switch msg := msg.(type) {
		case SetupDoneMsg:
			m.showSetup = false
			m.setup = nil
			return m, nil
		case tea.WindowSizeMsg:
			m.width = msg.Width
			m.height = msg.Height
			mod, cmd := m.setup.Update(msg)
			m.setup = mod.(*SetupModal)
			return m, cmd
		default:
			mod, cmd := m.setup.Update(msg)
			m.setup = mod.(*SetupModal)
			return m, cmd
		}
	}

	switch msg := msg.(type) {
	case TickMsg:
		return m, tick()
	case tea.KeyMsg:
		return m.handleKey(msg)
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		innerH := m.height - 5
		if innerH < 0 { innerH = 0 }
		screenMsg := tea.WindowSizeMsg{Width: m.width, Height: innerH}
		for i := range m.screens {
			mod, _ := m.screens[i].Update(screenMsg)
			m.screens[i] = mod.(screens.Screen)
		}
		return m, nil
	case core.SwitchToDashboardMsg:
		m.activeTab = 0
		m.tabs.SetActive(0)
		m.mode = ModeNormal
		return m, nil
	case tea.MouseMsg:
		return m.handleMouse(msg)
	}

	var cmd tea.Cmd
	var mod tea.Model
	mod, cmd = m.screens[m.activeTab].Update(msg)
	m.screens[m.activeTab] = mod.(screens.Screen)

	m.statusBar.SetActiveTab(m.activeTab)
	totalErrs, totalWarns := 0, 0
	for _, b := range m.bm.All() {
		for _, e := range b.Errors {
			if e.Level == core.LogError { totalErrs++ } else { totalWarns++ }
		}
	}
	m.statusBar.SetErrors(totalErrs, totalWarns)

	return m, cmd
}

func (m *AppModel) handleMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	if msg.Type == tea.MouseLeft && msg.Y == 0 {
		tabW := m.width / len(m.screens)
		if tabW > 0 {
			tabIdx := msg.X / tabW
			if tabIdx >= 0 && tabIdx < len(m.screens) {
				m.activeTab = tabIdx
				m.tabs.SetActive(tabIdx)
				return m, nil
			}
		}
	}
	mod, cmd := m.screens[m.activeTab].Update(msg)
	m.screens[m.activeTab] = mod.(screens.Screen)
	return m, cmd
}

func (m *AppModel) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if key.Matches(msg, m.keys.Esc) {
		m.mode = ModeNormal
		m.commandBuf = ""
		m.screens[m.activeTab].Blur()
		mod, cmd := m.screens[m.activeTab].Update(msg)
		m.screens[m.activeTab] = mod.(screens.Screen)
		return m, cmd
	}

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
	if key.Matches(msg, m.keys.Command) {
		m.mode = ModeCommand
		return m, nil
	}
	if key.Matches(msg, m.keys.Quit) {
		return m, tea.Quit
	}

	switchTab := func(idx int) (tea.Model, tea.Cmd) {
		m.screens[m.activeTab].Blur()
		m.activeTab = idx
		m.tabs.SetActive(idx)
		m.screens[m.activeTab].Focus()
		if idx == 1 { m.mode = ModeInsert } else { m.mode = ModeNormal }
		return m, nil
	}

	if key.Matches(msg, m.keys.Tab1) { return switchTab(0) }
	if key.Matches(msg, m.keys.Tab2) { return switchTab(1) }
	if key.Matches(msg, m.keys.Tab3) { return switchTab(2) }
	if key.Matches(msg, m.keys.Tab4) { return switchTab(3) }
	if key.Matches(msg, m.keys.Help, m.keys.Tab5) { return switchTab(4) }

	if msg.String() == "i" || msg.String() == "a" || msg.String() == "enter" {
		m.mode = ModeInsert
		m.screens[m.activeTab].Focus()
		if msg.String() == "enter" {
			mod, cmd := m.screens[m.activeTab].Update(msg)
			m.screens[m.activeTab] = mod.(screens.Screen)
			return m, cmd
		}
		return m, nil
	}

	mod, cmd := m.screens[m.activeTab].Update(msg)
	m.screens[m.activeTab] = mod.(screens.Screen)
	return m, cmd
}

func (m *AppModel) handleInsertKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	mod, cmd := m.screens[m.activeTab].Update(msg)
	m.screens[m.activeTab] = mod.(screens.Screen)
	return m, cmd
}

func (m *AppModel) handleCommandKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if key.Matches(msg, m.keys.Enter) {
		return m.executeCommand(m.commandBuf)
	}
	if msg.Type == tea.KeyRunes || msg.Type == tea.KeySpace {
		m.commandBuf += string(msg.Runes)
	} else if msg.Type == tea.KeyBackspace {
		if len(m.commandBuf) > 0 { m.commandBuf = m.commandBuf[:len(m.commandBuf)-1] }
	}
	return m, nil
}

func (m *AppModel) executeCommand(cmdStr string) (tea.Model, tea.Cmd) {
	if strings.TrimSpace(cmdStr) == "q" || strings.TrimSpace(cmdStr) == "quit" {
		return m, tea.Quit
	}
	m.mode = ModeNormal
	m.commandBuf = ""
	return m, nil
}

func (m *AppModel) View() string {
	if m.width <= 0 || m.height <= 0 { return "Initializing..." }
	if m.showSetup && m.setup != nil { return m.setup.View() }

	tabRow := m.tabs.View(m.width)
	screen := m.screens[m.activeTab].View()
	statusRow := m.statusBar.View(m.mode, m.width)

	sep := lipgloss.NewStyle().Foreground(lipgloss.Color("#313244")).Render(strings.Repeat("─", m.width))

	return lipgloss.JoinVertical(lipgloss.Left,
		tabRow,
		sep,
		screen,
		sep,
		statusRow,
	)
}

func tickStatsUpdate() tea.Cmd {
	return tea.Tick(time.Second*1, func(time.Time) tea.Msg {
		return StatsUpdateMsg{}
	})
}
