package main

import (
	"fmt"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/MemestaVedas/gobuild/internal/config"
	"github.com/MemestaVedas/gobuild/internal/core"
	"github.com/MemestaVedas/gobuild/internal/ipc"
	"github.com/MemestaVedas/gobuild/internal/platform/windows" // using windows directly for now or conditionally loaded
	"github.com/MemestaVedas/gobuild/internal/plugin"
	"github.com/MemestaVedas/gobuild/internal/tui"
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "gobuild",
		Short: "Developer-centric build monitor",
		Run:   runApp,
	}

	runCmd := &cobra.Command{
		Use:   "run [command]",
		Short: "Run a build command and monitor it",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			// Stub for wrapper mode
			fmt.Printf("Starting wrapped build: %v\n", args)
		},
	}
	rootCmd.AddCommand(runCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func runApp(cmd *cobra.Command, args []string) {
	// Initialize logging
	logFile, err := os.OpenFile("gobuild.log", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if err == nil {
		log.SetOutput(logFile)
		defer logFile.Close()
	}

	// 1. Config
	if err := config.EnsureConfigDir(); err != nil {
		log.Fatalf("Config dir error: %v", err)
	}
	cfg, err := config.Load()
	if err != nil {
		log.Printf("Using default config, load error: %v", err)
		cfg = config.DefaultConfig()
	}

	// 2. Core services
	bm := core.NewBuildManager()
	eb := plugin.NewEventBus()
	plugin.LoadBuiltins(eb, cfg.Plugins)

	// Platform abstraction (hardcoded Windows for this environment example)
	plat := windows.New()

	// 3. IPC Server
	ipcServer := ipc.NewServer(cfg.Server.WSPort, bm, nil, nil)
	if err := ipcServer.Start(); err != nil {
		log.Printf("IPC Server start failed: %v", err)
	}

	broadcaster := ipc.NewBroadcaster(cfg.Server.UDPPort, cfg.Server.WSPort)
	broadcaster.Start()
	defer broadcaster.Stop()

	// 4. Start TUI
	app := tui.NewAppModel()

	// The tea.WithAltScreen() ensures we restore shell history afterward
	p := tea.NewProgram(app, tea.WithAltScreen(), tea.WithMouseCellMotion())

	log.Printf("Starting GoBuild on %s...", plat.Name())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Fatal error: %v\n", err)
		os.Exit(1)
	}
}
