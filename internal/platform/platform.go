package platform

import "github.com/MemestaVedas/gobuild/internal/core"

// Platform abstracts all OS-specific operations.
// Implementations are selected at compile time via build tags.
type Platform interface {
	// Process discovery
	ScanBuildProcesses() ([]core.ProcessInfo, error)
	WatchProcess(pid int, onChange func(core.ProcessInfo)) error

	// File system
	WatchDirectory(path string, onChange func(core.FileEvent)) error
	DetectBuildTool(dir string) (core.BuildTool, error)

	// Notifications
	SendNotification(title, body string) error
	PlaySound(soundType core.SoundType) error

	// System stats
	GetCPUPercent() (float64, error)
	GetRAMUsage() (used, total uint64, err error)
	GetNetworkIO() (up, down uint64, err error) // bytes/s

	// Platform info
	Name() string // "linux" or "windows"
}
