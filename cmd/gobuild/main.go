package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
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

	"charm.land/lipgloss/v2"
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
	
	proxyCmd := &cobra.Command{
		Use:   "proxy [command]",
		Short: "Proxy a command to the running gobuild daemon",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			cwd, _ := os.Getwd()
			err := ipc.SendProxyRequest(cwd, args[0], args[1:])
			if err != nil {
				fmt.Fprintf(os.Stderr, "Proxy failed: %v\n", err)
				os.Exit(1)
			}
		},
	}

	rootCmd.AddCommand(runCmd, proxyCmd)

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

	// 4.5. Daemon Server for proxied commands
	daemon := ipc.NewDaemonServer(func(cmd string, proxyArgs []string, cwd string) error {
		// Build the full command string properly
		fullCmd := cmd
		if len(proxyArgs) > 0 {
			fullCmd = cmd + " " + strings.Join(proxyArgs, " ")
		}
		
		b := &core.Build{
			ID:      fmt.Sprintf("proxy-%d", time.Now().UnixNano()),
			Name:    filepath.Base(cwd),
			Command: fullCmd,
			WorkDir: cwd,
			Tool:    core.ToolGeneric,
		}
		
		log.Printf("Daemon: starting '%s' in %s", fullCmd, cwd)
		bm.Add(b)
		return bldr.StartBuild(b)
	})
	if err := daemon.Start(); err != nil {
		log.Printf("Daemon start failed: %v", err)
	}
	defer daemon.Stop()

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
						// New build discovered externally
						buildID := fmt.Sprintf("ext-%d-%d", p.PID, time.Now().Unix())
						b := &core.Build{
							ID:        buildID,
							Name:      p.Name,
							Command:   p.CmdLine,
							PID:       p.PID,
							State:     core.StateBuilding,
							StartTime: time.Now(),
							Tool:      core.ToolGeneric, // Could try to refine
							LogLines: []core.LogLine{
								{Timestamp: time.Now(), Level: core.LogInfo, Raw: "[External Process] Output cannot be captured because this process was"},
								{Timestamp: time.Now(), Level: core.LogInfo, Raw: "started outside of the gobuild shell proxy hook."},
								{Timestamp: time.Now(), Level: core.LogInfo, Raw: "To capture output, you must integrate the shell hook by running:"},
								{Timestamp: time.Now(), Level: core.LogInfo, Raw: "    source /path/to/gobuild/scripts/gobuild-hook.sh"},
							},
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
	isDark := lipgloss.HasDarkBackground(os.Stdin, os.Stderr)
	app := tui.NewAppModel(bm, bldr, isDark)

	// The tea.WithAltScreen() ensures we restore shell history afterward
	p := tea.NewProgram(app, tea.WithAltScreen(), tea.WithMouseCellMotion())

	log.Printf("Starting GoBuild on %s...", plat.Name())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Fatal error: %v\n", err)
		os.Exit(1)
	}
}
