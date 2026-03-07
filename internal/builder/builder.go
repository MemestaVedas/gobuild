package builder

import (
	"bufio"
	"io"
	"os/exec"
	"sync"
	"time"

	"github.com/MemestaVedas/gobuild/internal/core"
	"github.com/MemestaVedas/gobuild/internal/parser"
)

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
	cmd := exec.Command("cmd", "/C", build.Command)
	if build.WorkDir != "" {
		cmd.Dir = build.WorkDir
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	// 2. Initialise build state
	now := time.Now()
	b.bm.Update(build.ID, func(build *core.Build) {
		build.State = core.StateBuilding
		build.StartTime = now
		build.PID = 0 // Will update once started
	})

	if err := cmd.Start(); err != nil {
		b.bm.Update(build.ID, func(build *core.Build) {
			build.State = core.StateFailed
			endTime := time.Now()
			build.EndTime = &endTime
		})
		return err
	}

	b.bm.Update(build.ID, func(build *core.Build) {
		build.PID = cmd.Process.Pid
	})

	// 3. Monitor output
	var wg sync.WaitGroup
	wg.Add(2)

	p := parser.For(build.Tool)

	scanFunc := func(r io.Reader, level core.LogLevel) {
		defer wg.Done()
		scanner := bufio.NewScanner(r)
		for scanner.Scan() {
			line := scanner.Text()
			logLine, ok := p.ParseLine(line)
			if !ok {
				logLine = core.LogLine{
					Timestamp: time.Now(),
					Level:     level,
					Raw:       line,
				}
			}

			b.bm.AppendLog(build.ID, logLine, 500)

			// Update progress if possible
			if prog, ok := p.ParseProgress(line); ok {
				b.bm.Update(build.ID, func(build *core.Build) {
					build.Progress = prog
				})
			}

			// If it's an error/warning level, also append to structured errors
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
	}

	go scanFunc(stdout, core.LogInfo)
	go scanFunc(stderr, core.LogError)

	// 4. Wait for completion
	go func() {
		wg.Wait()
		cmdErr := cmd.Wait()
		endTime := time.Now()

		b.bm.Update(build.ID, func(build *core.Build) {
			if cmdErr == nil {
				build.State = core.StateSuccess
			} else {
				build.State = core.StateFailed
			}
			build.EndTime = &endTime
			build.Duration = build.EndTime.Sub(build.StartTime)
			build.Progress = 1.0
		})
	}()

	return nil
}
