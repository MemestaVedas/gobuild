package builder

import (
	"bufio"
	"os/exec"
	"regexp"
	"runtime"
	"time"

	"github.com/creack/pty"

	"github.com/MemestaVedas/gobuild/internal/core"
	"github.com/MemestaVedas/gobuild/internal/parser"
)

// ansiEscape strips ANSI terminal escape sequences from a string.
var ansiEscape = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]|\x1b\[[0-9]*[a-zA-Z]|\x1b[@-Z\\-_]|\x1b\].*?(?:\x07|\x1b\\)`)

func stripANSI(s string) string {
	return ansiEscape.ReplaceAllString(s, "")
}

// Builder orchestrates the execution and monitoring of builds.
type Builder struct {
	bm *core.BuildManager
}

// New creates a new Builder instance.
func New(bm *core.BuildManager) *Builder {
	return &Builder{bm: bm}
}

// StartBuild kicks off a build process and monitors its output.
func (b *Builder) StartBuild(build *core.Build) error {
	// 1. Prepare command
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/C", build.Command)
	} else {
		cmd = exec.Command("sh", "-c", build.Command)
	}
	if build.WorkDir != "" {
		cmd.Dir = build.WorkDir
	}

	// Start in a PTY for interactive output
	ptmx, err := pty.Start(cmd)
	if err != nil {
		b.bm.Update(build.ID, func(build *core.Build) {
			build.State = core.StateFailed
			endTime := time.Now()
			build.EndTime = &endTime
		})
		return err
	}

	// 2. Initialise build state
	now := time.Now()
	b.bm.Update(build.ID, func(build *core.Build) {
		build.State = core.StateBuilding
		build.StartTime = now
		build.PID = cmd.Process.Pid
		build.PTY = ptmx
	})

	p := parser.For(build.Tool)

	// 3. Read PTY output in one goroutine (PTY merges stdout+stderr)
	go func() {
		scanner := bufio.NewScanner(ptmx)
		for scanner.Scan() {
			raw := scanner.Text()
			// Strip ANSI escape codes before storing
			clean := stripANSI(raw)
			if clean == "" {
				continue
			}

			logLine, ok := p.ParseLine(clean)
			if !ok {
				logLine = core.LogLine{
					Timestamp: time.Now(),
					Level:     core.LogInfo,
					Raw:       clean,
				}
			}

			b.bm.AppendLog(build.ID, logLine, 1000)

			if prog, ok := p.ParseProgress(clean); ok {
				b.bm.Update(build.ID, func(build *core.Build) {
					build.Progress = prog
				})
			}

			if logLine.Level == core.LogError || logLine.Level == core.LogWarning {
				if logLine.Parsed != nil {
					b.bm.AppendError(build.ID, core.BuildError{
						File:    logLine.Parsed.File,
						Line:    logLine.Parsed.Line,
						Column:  logLine.Parsed.Column,
						Code:    logLine.Parsed.Code,
						Message: logLine.Parsed.Message,
						Level:   logLine.Level,
					})
				}
			}
		}

		// 4. Scanner done (PTY closed) — wait for process exit code
		cmdErr := cmd.Wait()
		ptmx.Close()

		endTime := time.Now()
		b.bm.Update(build.ID, func(build *core.Build) {
			if cmdErr == nil {
				build.State = core.StateSuccess
			} else {
				if e, ok := cmdErr.(*exec.ExitError); ok && e.ExitCode() == 0 {
					build.State = core.StateSuccess
				} else {
					build.State = core.StateFailed
				}
			}
			build.EndTime = &endTime
			build.Duration = build.EndTime.Sub(build.StartTime)
			build.Progress = 1.0
		})
	}()

	return nil
}
