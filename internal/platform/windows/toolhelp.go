// internal/platform/windows/toolhelp.go
//go:build windows

package windows

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
	"unsafe"

	"github.com/MemestaVedas/gobuild/internal/core"
	"github.com/fsnotify/fsnotify"
	"golang.org/x/sys/windows"
)

// WindowsPlatform implements platform.Platform for Windows.
type WindowsPlatform struct {
	mu          sync.Mutex
	prevCPU     [2]uint64
	prevNetIO   [2]uint64
	prevNetTime time.Time
}

// New creates a new WindowsPlatform.
func New() *WindowsPlatform {
	return &WindowsPlatform{}
}

func (w *WindowsPlatform) Name() string { return "windows" }

// buildExeNames maps known build executable names to BuildTool.
var buildExeNames = map[string]core.BuildTool{
	"cargo.exe":        core.ToolCargo,
	"npm.cmd":          core.ToolNPM,
	"npm.exe":          core.ToolNPM,
	"yarn.cmd":         core.ToolNPM,
	"make.exe":         core.ToolMake,
	"mingw32-make.exe": core.ToolMake,
	"go.exe":           core.ToolGo,
	"gradle.bat":       core.ToolGradle,
	"gradlew.bat":      core.ToolGradle,
	"cmake.exe":        core.ToolCMake,
	"msbuild.exe":      core.ToolMSBuild,
	"dotnet.exe":       core.ToolMSBuild,
}

const (
	TH32CS_SNAPPROCESS                = 0x00000002
	PROCESS_QUERY_LIMITED_INFORMATION = 0x1000
	PROCESS_VM_READ                   = 0x0010
)

type processEntry32 struct {
	Size            uint32
	Usage           uint32
	ProcessID       uint32
	DefaultHeapID   uintptr
	ModuleID        uint32
	Threads         uint32
	ParentProcessID uint32
	PriClassBase    int32
	Flags           uint32
	ExeFile         [windows.MAX_PATH]uint16
}

var (
	modKernel32                  = windows.NewLazySystemDLL("kernel32.dll")
	procCreateToolhelp32Snapshot = modKernel32.NewProc("CreateToolhelp32Snapshot")
	procProcess32First           = modKernel32.NewProc("Process32FirstW")
	procProcess32Next            = modKernel32.NewProc("Process32NextW")
)

// ScanBuildProcesses uses ToolHelp32 to enumerate all processes.
func (w *WindowsPlatform) ScanBuildProcesses() ([]core.ProcessInfo, error) {
	snapshot, _, err := procCreateToolhelp32Snapshot.Call(TH32CS_SNAPPROCESS, 0)
	if snapshot == uintptr(windows.InvalidHandle) {
		return nil, fmt.Errorf("CreateToolhelp32Snapshot: %w", err)
	}
	defer windows.CloseHandle(windows.Handle(snapshot))

	var entry processEntry32
	entry.Size = uint32(unsafe.Sizeof(entry))

	ret, _, _ := procProcess32First.Call(snapshot, uintptr(unsafe.Pointer(&entry)))
	if ret == 0 {
		return nil, fmt.Errorf("Process32First failed")
	}

	var results []core.ProcessInfo
	for {
		exeName := windows.UTF16ToString(entry.ExeFile[:])
		base := strings.ToLower(filepath.Base(exeName))
		if tool, ok := buildExeNames[base]; ok {
			info := core.ProcessInfo{
				PID:     int(entry.ProcessID),
				PPID:    int(entry.ParentProcessID),
				Name:    base,
				CmdLine: exeName,
			}
			_ = tool
			results = append(results, info)
		}

		ret, _, _ = procProcess32Next.Call(snapshot, uintptr(unsafe.Pointer(&entry)))
		if ret == 0 {
			break
		}
	}
	return results, nil
}

// WatchProcess polls for the given PID until it exits.
func (w *WindowsPlatform) WatchProcess(pid int, onChange func(core.ProcessInfo)) error {
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		handle, err := windows.OpenProcess(PROCESS_QUERY_LIMITED_INFORMATION, false, uint32(pid))
		if err != nil {
			return
		}
		defer windows.CloseHandle(handle)
		for range ticker.C {
			var exitCode uint32
			if err := windows.GetExitCodeProcess(handle, &exitCode); err != nil || exitCode != 259 {
				// 259 = STILL_ACTIVE
				return
			}
			onChange(core.ProcessInfo{PID: pid})
		}
	}()
	return nil
}

// WatchDirectory uses fsnotify (wraps ReadDirectoryChangesW).
func (w *WindowsPlatform) WatchDirectory(path string, onChange func(core.FileEvent)) error {
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
				onChange(core.FileEvent{Path: event.Name, Op: event.Op.String()})
			case _, ok := <-watcher.Errors:
				if !ok {
					return
				}
			}
		}
	}()
	return nil
}

// DetectBuildTool inspects the directory for build marker files.
func (w *WindowsPlatform) DetectBuildTool(dir string) (core.BuildTool, error) {
	checks := []struct {
		file string
		tool core.BuildTool
	}{
		{"Cargo.toml", core.ToolCargo},
		{"package.json", core.ToolNPM},
		{"Makefile", core.ToolMake},
		{"go.mod", core.ToolGo},
		{"build.gradle", core.ToolGradle},
		{"build.gradle.kts", core.ToolGradle},
		{"CMakeLists.txt", core.ToolCMake},
	}
	// MSBuild: look for .sln or .vcxproj
	entries, _ := os.ReadDir(dir)
	for _, e := range entries {
		ext := strings.ToLower(filepath.Ext(e.Name()))
		if ext == ".sln" || ext == ".vcxproj" || ext == ".csproj" {
			return core.ToolMSBuild, nil
		}
	}
	for _, c := range checks {
		if _, err := os.Stat(filepath.Join(dir, c.file)); err == nil {
			return c.tool, nil
		}
	}
	return core.ToolGeneric, nil
}

// SendNotification shows a Windows toast notification via go-toast.
func (w *WindowsPlatform) SendNotification(title, body string) error {
	// Use PowerShell as a portable fallback that works on all Windows versions.
	script := fmt.Sprintf(`
[Windows.UI.Notifications.ToastNotificationManager, Windows.UI.Notifications, ContentType = WindowsRuntime] | Out-Null
$template = [Windows.UI.Notifications.ToastNotificationManager]::GetTemplateContent([Windows.UI.Notifications.ToastTemplateType]::ToastText02)
$template.SelectSingleNode('//text[@id=1]').InnerText = '%s'
$template.SelectSingleNode('//text[@id=2]').InnerText = '%s'
[Windows.UI.Notifications.ToastNotificationManager]::CreateToastNotifier('GoBuild').Show([Windows.UI.Notifications.ToastNotification]::new($template))
`, title, body)
	return exec.Command("powershell", "-NoProfile", "-Command", script).Run()
}

// PlaySound plays a system sound.
func (w *WindowsPlatform) PlaySound(soundType core.SoundType) error {
	var sound string
	switch soundType {
	case core.SoundSuccess:
		sound = `[System.Media.SystemSounds]::Asterisk.Play()`
	case core.SoundFailure:
		sound = `[System.Media.SystemSounds]::Hand.Play()`
	default:
		sound = `[System.Media.SystemSounds]::Beep.Play()`
	}
	return exec.Command("powershell", "-NoProfile", "-Command", sound).Run()
}

// GetCPUPercent uses WMI via PowerShell (simple, no CGO).
func (w *WindowsPlatform) GetCPUPercent() (float64, error) {
	out, err := exec.Command(
		"powershell", "-NoProfile", "-Command",
		"(Get-WmiObject -Class Win32_Processor | Measure-Object -Property LoadPercentage -Average).Average",
	).Output()
	if err != nil {
		return 0, err
	}
	var pct float64
	fmt.Sscanf(strings.TrimSpace(string(out)), "%f", &pct)
	return pct, nil
}

// GetRAMUsage reads GlobalMemoryStatusEx via PowerShell.
func (w *WindowsPlatform) GetRAMUsage() (used, total uint64, err error) {
	out, e := exec.Command(
		"powershell", "-NoProfile", "-Command",
		`$m=(Get-WmiObject -Class Win32_OperatingSystem); "$($m.TotalVisibleMemorySize) $($m.FreePhysicalMemory)"`,
	).Output()
	if e != nil {
		return 0, 0, e
	}
	parts := strings.Fields(strings.TrimSpace(string(out)))
	if len(parts) < 2 {
		return 0, 0, fmt.Errorf("unexpected output")
	}
	fmt.Sscanf(parts[0], "%d", &total)
	var free uint64
	fmt.Sscanf(parts[1], "%d", &free)
	total *= 1024
	free *= 1024
	used = total - free
	return used, total, nil
}

// GetNetworkIO returns 0 values — full implementation requires ETW or PDH (v2).
func (w *WindowsPlatform) GetNetworkIO() (up, down uint64, err error) {
	return 0, 0, nil
}
