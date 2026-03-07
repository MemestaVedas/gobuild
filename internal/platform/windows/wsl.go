// internal/platform/windows/wsl.go
//go:build windows

package windows

import (
	"os/exec"
	"strings"

	"github.com/MemestaVedas/gobuild/internal/core"
)

// scanWSLProcesses enumerates build processes inside running WSL2 distros.
func (w *WindowsPlatform) scanWSLProcesses() ([]core.ProcessInfo, error) {
	// Get running distros
	distrosOut, err := exec.Command("wsl.exe", "--list", "--running", "--quiet").Output()
	if err != nil {
		return nil, nil // WSL not available or no distros running
	}

	var results []core.ProcessInfo
	distros := strings.Fields(strings.ReplaceAll(string(distrosOut), "\r", ""))

	for _, distro := range distros {
		if distro == "" {
			continue
		}
		out, err := exec.Command("wsl.exe", "-d", distro, "-e", "ps", "aux", "--no-headers").Output()
		if err != nil {
			continue
		}
		for _, line := range strings.Split(string(out), "\n") {
			fields := strings.Fields(line)
			if len(fields) < 11 {
				continue
			}
			cmdline := strings.Join(fields[10:], " ")
			if !isBuildCmdLine(cmdline) {
				continue
			}
			var pid int
			if _, err := strings.NewReader(fields[1]).Read(nil); err == nil {
				// Use sscanf-style parse
				pid = parseIntFast(fields[1])
			}
			results = append(results, core.ProcessInfo{
				PID:     pid,
				Name:    fields[10],
				CmdLine: cmdline,
				CWD:     "", // not available from ps
			})
		}
	}
	return results, nil
}

func isBuildCmdLine(cmdline string) bool {
	buildPrefixes := []string{
		"cargo build", "cargo test", "cargo check",
		"npm run", "npm build",
		"make ", "make\t",
		"go build", "go test",
		"gradle", "maven",
		"cmake --build",
	}
	lower := strings.ToLower(cmdline)
	for _, prefix := range buildPrefixes {
		if strings.Contains(lower, prefix) {
			return true
		}
	}
	return false
}

func parseIntFast(s string) int {
	n := 0
	for _, c := range s {
		if c >= '0' && c <= '9' {
			n = n*10 + int(c-'0')
		}
	}
	return n
}
