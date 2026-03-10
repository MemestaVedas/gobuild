// internal/platform/linux/proc.go
//go:build linux

package linux

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/MemestaVedas/gobuild/internal/core"
	"github.com/fsnotify/fsnotify"
)

// LinuxPlatform implements platform.Platform for Linux using /proc and inotify.
type LinuxPlatform struct {
	mu          sync.Mutex
	prevCPU     [2]uint64 // [idle, total]
	prevNetIO   [2]uint64 // [rx_bytes, tx_bytes]
	prevNetTime time.Time
}

// New creates a new LinuxPlatform.
func New() *LinuxPlatform {
	return &LinuxPlatform{}
}

func (l *LinuxPlatform) Name() string { return "linux" }

// buildCommands is the set of executable names that indicate a build process.
var buildCommands = map[string]core.BuildTool{
	"cargo":   core.ToolCargo,
	"npm":     core.ToolNPM,
	"yarn":    core.ToolNPM,
	"make":    core.ToolMake,
	"gmake":   core.ToolMake,
	"go":      core.ToolGo,
	"gradle":  core.ToolGradle,
	"gradlew": core.ToolGradle,
	"cmake":   core.ToolCMake,
}

// buildSubcommands filters go/npm/cargo by sub-command.
var buildSubcommands = map[string][]string{
	"cargo": {"build", "test", "check", "run"},
	"go":    {"build", "test", "generate", "install", "run"},
	"npm":   {"run", "build", "install", "ci"},
}

// ScanBuildProcesses walks /proc and returns all active build processes.
func (l *LinuxPlatform) ScanBuildProcesses() ([]core.ProcessInfo, error) {
	entries, err := os.ReadDir("/proc")
	if err != nil {
		return nil, fmt.Errorf("reading /proc: %w", err)
	}

	var results []core.ProcessInfo
	for _, e := range entries {
		pid, err := strconv.Atoi(e.Name())
		if err != nil {
			continue // skip non-PID directories
		}

		cmdline, err := readProcFile(pid, "cmdline")
		if err != nil {
			continue
		}

		// /proc/<pid>/cmdline is NUL-separated
		parts := strings.Split(strings.TrimRight(cmdline, "\x00"), "\x00")
		if len(parts) == 0 {
			continue
		}

		if !isBuildCommand(parts) {
			continue
		}

		info, err := buildProcessInfo(pid, parts)
		if err == nil {
			results = append(results, info)
		}
	}
	return results, nil
}

// WatchProcess polls /proc/<pid>/stat until the process exits.
func (l *LinuxPlatform) WatchProcess(pid int, onChange func(core.ProcessInfo)) error {
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			cmdline, err := readProcFile(pid, "cmdline")
			if err != nil {
				// Process ended
				onChange(core.ProcessInfo{PID: pid})
				return
			}
			parts := strings.Split(strings.TrimRight(cmdline, "\x00"), "\x00")
			_, err = buildProcessInfo(pid, parts)
			if err != nil {
				onChange(core.ProcessInfo{PID: pid})
				return
			}
		}
	}()
	return nil
}

// WatchDirectory installs an fsnotify watcher on the given path.
func (l *LinuxPlatform) WatchDirectory(path string, onChange func(core.FileEvent)) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	if err := watcher.Add(path); err != nil {
		watcher.Close()
		return err
	}
	go func() {
		defer watcher.Close()
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				onChange(core.FileEvent{
					Path: event.Name,
					Op:   event.Op.String(),
				})
			case _, ok := <-watcher.Errors:
				if !ok {
					return
				}
			}
		}
	}()
	return nil
}

// DetectBuildTool inspects a directory for well-known build files.
func (l *LinuxPlatform) DetectBuildTool(dir string) (core.BuildTool, error) {
	checks := []struct {
		file string
		tool core.BuildTool
	}{
		{"Cargo.toml", core.ToolCargo},
		{"package.json", core.ToolNPM},
		{"Makefile", core.ToolMake},
		{"GNUmakefile", core.ToolMake},
		{"go.mod", core.ToolGo},
		{"build.gradle", core.ToolGradle},
		{"build.gradle.kts", core.ToolGradle},
		{"CMakeLists.txt", core.ToolCMake},
	}
	for _, c := range checks {
		if _, err := os.Stat(filepath.Join(dir, c.file)); err == nil {
			return c.tool, nil
		}
	}
	return core.ToolGeneric, nil
}

// SendNotification sends a desktop notification via notify-send.
func (l *LinuxPlatform) SendNotification(title, body string) error {
	return exec.Command(
		"notify-send",
		"--app-name=GoBuild",
		"--icon=utilities-terminal",
		title,
		body,
	).Run()
}

// PlaySound plays an audio file via paplay (PulseAudio) or aplay (ALSA).
func (l *LinuxPlatform) PlaySound(soundType core.SoundType) error {
	file := soundFile(soundType)
	if err := exec.Command("paplay", file).Run(); err != nil {
		return exec.Command("aplay", file).Run()
	}
	return nil
}

func soundFile(soundType core.SoundType) string {
	switch soundType {
	case core.SoundSuccess:
		return "/usr/share/sounds/freedesktop/stereo/complete.oga"
	case core.SoundFailure:
		return "/usr/share/sounds/freedesktop/stereo/dialog-error.oga"
	default:
		return "/usr/share/sounds/freedesktop/stereo/message.oga"
	}
}

// GetCPUPercent reads /proc/stat to compute CPU usage since last call.
func (l *LinuxPlatform) GetCPUPercent() (float64, error) {
	data, err := os.ReadFile("/proc/stat")
	if err != nil {
		return 0, err
	}
	lines := strings.Split(string(data), "\n")
	if len(lines) == 0 {
		return 0, fmt.Errorf("empty /proc/stat")
	}
	fields := strings.Fields(lines[0])
	if len(fields) < 5 || fields[0] != "cpu" {
		return 0, fmt.Errorf("unexpected /proc/stat format")
	}
	var vals [10]uint64
	for i := 1; i < len(fields) && i <= 10; i++ {
		vals[i-1], _ = strconv.ParseUint(fields[i], 10, 64)
	}
	idle := vals[3] + vals[4] // idle + iowait
	total := uint64(0)
	for _, v := range vals {
		total += v
	}

	l.mu.Lock()
	prevIdle := l.prevCPU[0]
	prevTotal := l.prevCPU[1]
	l.prevCPU[0] = idle
	l.prevCPU[1] = total
	l.mu.Unlock()

	diffIdle := float64(idle - prevIdle)
	diffTotal := float64(total - prevTotal)
	if diffTotal == 0 {
		return 0, nil
	}
	return (1 - diffIdle/diffTotal) * 100, nil
}

// GetRAMUsage reads /proc/meminfo for used/total memory.
func (l *LinuxPlatform) GetRAMUsage() (used, total uint64, err error) {
	data, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		return 0, 0, err
	}
	m := make(map[string]uint64)
	for _, line := range strings.Split(string(data), "\n") {
		fields := strings.Fields(line)
		if len(fields) >= 2 {
			key := strings.TrimSuffix(fields[0], ":")
			val, _ := strconv.ParseUint(fields[1], 10, 64)
			m[key] = val * 1024 // kB → bytes
		}
	}
	total = m["MemTotal"]
	free := m["MemFree"] + m["Buffers"] + m["Cached"] + m["SReclaimable"]
	used = total - free
	return used, total, nil
}

// GetNetworkIO reads /proc/net/dev and returns bytes/s since last call.
func (l *LinuxPlatform) GetNetworkIO() (up, down uint64, err error) {
	data, err := os.ReadFile("/proc/net/dev")
	if err != nil {
		return 0, 0, err
	}
	var rx, tx uint64
	for _, line := range strings.Split(string(data), "\n")[2:] {
		fields := strings.Fields(line)
		if len(fields) < 10 {
			continue
		}
		iface := strings.TrimSuffix(fields[0], ":")
		if iface == "lo" {
			continue
		}
		r, _ := strconv.ParseUint(fields[1], 10, 64)
		t, _ := strconv.ParseUint(fields[9], 10, 64)
		rx += r
		tx += t
	}

	l.mu.Lock()
	prevRX := l.prevNetIO[0]
	prevTX := l.prevNetIO[1]
	prevTime := l.prevNetTime
	l.prevNetIO[0] = rx
	l.prevNetIO[1] = tx
	l.prevNetTime = time.Now()
	l.mu.Unlock()

	if prevTime.IsZero() {
		return 0, 0, nil
	}
	elapsed := time.Since(prevTime).Seconds()
	if elapsed == 0 {
		return 0, 0, nil
	}
	down = uint64(float64(rx-prevRX) / elapsed)
	up = uint64(float64(tx-prevTX) / elapsed)
	return up, down, nil
}

// --- helpers ---

func readProcFile(pid int, name string) (string, error) {
	data, err := os.ReadFile(fmt.Sprintf("/proc/%d/%s", pid, name))
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func isBuildCommand(parts []string) bool {
	if len(parts) == 0 {
		return false
	}
	
	// Skip common interpreters to find the actual compiled/script binary
	idx := 0
	for idx < len(parts) {
		base := filepath.Base(parts[idx])
		if base == "bash" || base == "sh" || base == "zsh" || base == "env" || base == "node" || base == "python" || base == "python3" {
			idx++
		} else {
			break
		}
	}
	
	if idx >= len(parts) {
		return false
	}

	base := filepath.Base(parts[idx])
	tool, ok := buildCommands[base]
	if !ok {
		// Android Gradle Daemon fallback
		if base == "java" || base == "java.exe" {
			for _, p := range parts {
				if strings.Contains(p, "gradle") {
					return true
				}
			}
		}
		return false
	}
	// For tools with sub-command filtering, check the next argument
	subs, hasSubs := buildSubcommands[base]
	if !hasSubs {
		return true
	}
	if idx+1 >= len(parts) {
		return false
	}
	for _, s := range subs {
		if parts[idx+1] == s {
			_ = tool
			return true
		}
	}
	return false
}

func buildProcessInfo(pid int, parts []string) (core.ProcessInfo, error) {
	statData, err := readProcFile(pid, "stat")
	if err != nil {
		return core.ProcessInfo{}, err
	}
	statFields := strings.Fields(statData)
	ppid := 0
	if len(statFields) > 3 {
		ppid, _ = strconv.Atoi(statFields[3])
	}

	cwd, _ := os.Readlink(fmt.Sprintf("/proc/%d/cwd", pid))

	// Skip interpreters for Name detection
	idx := 0
	for idx < len(parts) {
		base := filepath.Base(parts[idx])
		if base == "bash" || base == "sh" || base == "zsh" || base == "env" || base == "node" || base == "python" || base == "python3" {
			idx++
		} else {
			break
		}
	}
	
	name := "unknown"
	if idx < len(parts) {
		name = filepath.Base(parts[idx])
		if name == "java" || name == "java.exe" {
			for _, p := range parts {
				if strings.Contains(p, "gradle") {
					name = "gradle [daemon]"
					break
				}
			}
		}
	}

	return core.ProcessInfo{
		PID:     pid,
		PPID:    ppid,
		Name:    name,
		CmdLine: strings.Join(parts, " "), // Keep the whole cmdline for display/debug
		CWD:     cwd,
	}, nil
}
