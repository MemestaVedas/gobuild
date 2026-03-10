package tui

import (
	// Added fmt import
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"charm.land/lipgloss/v2"

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

		// Chrome = TopBorder(1) + Tabs(1) + StatusMode(1) + Hints(1) + BottomBorder(1) = 5
		wsHeight := m.height - 5
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
	case tea.MouseMsg:
		return m.handleMouse(msg)
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

func (m *AppModel) handleMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	if msg.Type != tea.MouseLeft {
		// Forward scroll/movement events to active screen
		mod, cmd := m.screens[m.activeTab].Update(msg)
		m.screens[m.activeTab] = mod.(screens.Screen)
		return m, cmd
	}

	// Y = 1 is the tabs row assuming Y=0 is the top border
	if msg.Y == 1 {
		x := msg.X
		newTab := -1
		if x >= 2 && x <= 13 {
			newTab = 0
		} else if x >= 15 && x <= 25 {
			newTab = 1
		} else if x >= 27 && x <= 36 {
			newTab = 2
		} else if x >= 38 && x <= 47 {
			newTab = 3
		} else if x >= 49 && x <= 55 {
			newTab = 4
		}

		if newTab != -1 && newTab != m.activeTab {
			m.screens[m.activeTab].Blur()
			m.activeTab = newTab
			m.tabs.SetActive(newTab)
			m.mode = ModeNormal
			m.commandBuf = ""
			
			// If moving to Launcher, automatically focus it
			if newTab == 1 {
				m.screens[m.activeTab].Focus()
			}
			return m, nil
		}
	}

	// Forward other clicks to active screen
	mod, cmd := m.screens[m.activeTab].Update(msg)
	m.screens[m.activeTab] = mod.(screens.Screen)
	return m, cmd
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
	// We removed manual top, bot, and separators. Now using a single Lipgloss box!
	tabsContent := lipgloss.PlaceHorizontal(innerW, lipgloss.Left, m.tabs.View(innerW))
	activeScreen := m.screens[m.activeTab].View()
	statusContent := m.statusBar.View(m.mode, innerW)

	// Combine inside the box
	content := lipgloss.JoinVertical(lipgloss.Left,
		tabsContent,
		activeScreen,
		statusContent,
	)

	// Apply gradient border around the whole application
	appStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForegroundBlend(m.styles.ColorAccent, lipgloss.Color("#89B4FA")).
		Width(innerW).
		Height(m.height - 2)

	return appStyle.Render(content)
}

// hintsBar removed — hints are now part of StatusBarModel.View()

// Global ticks
func tickStatsUpdate() tea.Cmd {
	return tea.Tick(time.Second*1, func(time.Time) tea.Msg {
		return StatsUpdateMsg{}
	})
}
