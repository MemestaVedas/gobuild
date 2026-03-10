package main

import (
	"fmt"
	"log"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/MemestaVedas/gobuild/internal/builder"
	"github.com/MemestaVedas/gobuild/internal/config"
	"github.com/MemestaVedas/gobuild/internal/core"
	"github.com/MemestaVedas/gobuild/internal/ipc"
	"github.com/MemestaVedas/gobuild/internal/platform"
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
	plat := platform.New()

	// 3. Builder
	bldr := builder.New(bm)

	// 4. IPC Server
	ipcServer := ipc.NewServer(cfg.Server.WSPort, bm, func(profile string) {
		// Android initiated build
		// For now, if profile is empty, we might need a default or lookup
		b := &core.Build{
			ID:      fmt.Sprintf("%d", time.Now().UnixNano()),
			Name:    "Android Triggered",
			Command: "echo 'Build started from Android'", // Placeholder
			Tool:    core.ToolGeneric,
		}
		bm.Add(b)
		bldr.StartBuild(b)
	}, nil)
	log.Printf("IPC WebSocket server listening on 0.0.0.0:%d", cfg.Server.WSPort)
	if err := ipcServer.Start(); err != nil {
		log.Printf("IPC Server start failed: %v", err)
	}

	broadcaster := ipc.NewBroadcaster(cfg.Server.UDPPort, cfg.Server.WSPort)
	broadcaster.Start()
	defer broadcaster.Stop()

	// 5. Background stats polling and build discovery
	go func() {
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			// Stats
			cpu, _ := plat.GetCPUPercent()
			_, _, _ = plat.GetRAMUsage()
			_, _, _ = plat.GetNetworkIO()

			// Android expects a unified "update" message with builds list
			activeBuilds := bm.Active()
			builds := make([]*ipc.HelloBuild, len(activeBuilds))
			for i, b := range activeBuilds {
				builds[i] = &ipc.HelloBuild{
					Project:         b.Name,
					Tool:            b.Tool.String(),
					Status:          b.State.String(),
					Progress:        b.Progress,
					PID:             b.PID,
					DurationSeconds: int(b.Elapsed().Seconds()),
				}
			}

			ipcServer.Broadcast(ipc.BuildUpdateMessage{
				Type:        ipc.MsgBuildUpdate,
				CPU:         cpu,
				Builds:      builds,
				ActiveCount: len(builds),
				Timestamp:   time.Now().Unix(),
			})

			// Build Discovery
			procs, err := plat.ScanBuildProcesses()
			if err == nil {
				for _, p := range procs {
					if _, ok := bm.FindByPID(p.PID); !ok {
						// New build discovered
						buildID := fmt.Sprintf("ext-%d-%d", p.PID, time.Now().Unix())
						b := &core.Build{
							ID:        buildID,
							Name:      p.Name,
							Command:   p.CmdLine,
							PID:       p.PID,
							State:     core.StateBuilding,
							StartTime: time.Now(),
							Tool:      core.ToolGeneric, // Could try to refine
						}
						bm.Add(b)

						// Watch for exit
						_ = plat.WatchProcess(p.PID, func(info core.ProcessInfo) {
							bm.Update(buildID, func(build *core.Build) {
								build.State = core.StateSuccess // Default for external
								now := time.Now()
								build.EndTime = &now
								build.Duration = build.EndTime.Sub(build.StartTime)
								build.Progress = 1.0
							})
						})
					}
				}
			}
		}
	}()

	// 6. Start TUI
	app := tui.NewAppModel(bm, bldr)

	// The tea.WithAltScreen() ensures we restore shell history afterward
	p := tea.NewProgram(app, tea.WithAltScreen(), tea.WithMouseCellMotion())

	log.Printf("Starting GoBuild on %s...", plat.Name())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Fatal error: %v\n", err)
		os.Exit(1)
	}
}
