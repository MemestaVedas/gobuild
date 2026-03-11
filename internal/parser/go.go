package parser

import (
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/MemestaVedas/gobuild/internal/core"
)

// GoParser parses standard Go toolchain output.
type GoParser struct {
	pkgCount int
}

var (
	// src/main.go:12:2: message
	goErrorFullRe = regexp.MustCompile(`^(.+\.go):(\d+):(\d+):\s*(.+)$`)
	// src/main.go:12: message
	goErrorShortRe = regexp.MustCompile(`^(.+\.go):(\d+):\s*(.+)$`)
	// # github.com/user/repo/pkg
	goPkgHeaderRe = regexp.MustCompile(`^#\s+(.+)$`)
)

func (p *GoParser) ToolName() string { return "go" }

func (p *GoParser) ParseLine(raw string) (core.LogLine, bool) {
	line := core.LogLine{
		Timestamp: time.Now(),
		Raw:       raw,
		Level:     core.LogInfo,
	}

	trimmed := strings.TrimSpace(raw)

	// Check for package headers
	if m := goPkgHeaderRe.FindStringSubmatch(trimmed); m != nil {
		p.pkgCount++
		line.Level = core.LogNote
		line.Parsed = &core.ParsedEntry{Message: "Compiling " + m[1]}
		return line, true
	}

	// Check for full error location
	if m := goErrorFullRe.FindStringSubmatch(trimmed); m != nil {
		ln, _ := strconv.Atoi(m[2])
		col, _ := strconv.Atoi(m[3])
		line.Level = core.LogError
		line.Parsed = &core.ParsedEntry{
			File:    m[1],
			Line:    ln,
			Column:  col,
			Message: m[4],
		}
		return line, true
	}

	// Check for short error location
	if m := goErrorShortRe.FindStringSubmatch(trimmed); m != nil {
		ln, _ := strconv.Atoi(m[2])
		line.Level = core.LogError
		line.Parsed = &core.ParsedEntry{
			File:    m[1],
			Line:    ln,
			Message: m[3],
		}
		return line, true
	}

	return line, false
}

func (p *GoParser) ParseProgress(raw string) (float64, bool) {
	// Go doesn't provide easy percentages, but we can fake incremental progress 
	// based on the number of package headers we've seen if we knew the total.
	// For now, we return false to let it stay at 0 or use generic heuristics.
	return 0, false
}
