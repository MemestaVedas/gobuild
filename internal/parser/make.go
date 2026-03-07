package parser

import (
	"regexp"
	"strings"
	"time"

	"github.com/MemestaVedas/gobuild/internal/core"
)

// MakeParser parses Makefile build output.
type MakeParser struct{}

var (
	makeErrorRe   = regexp.MustCompile(`(?i):\s*(error|fatal error):\s+(.+)`)
	makeWarningRe = regexp.MustCompile(`(?i):\s*warning:\s+(.+)`)
	makeDirRe     = regexp.MustCompile(`^make\[\d+\]: Entering directory`)
	makeFailRe    = regexp.MustCompile(`(?i)make.*:\s+\*\*\*\s+\[.+\]\s+(Error|failed)`)
)

func (p *MakeParser) ToolName() string { return "make" }

func (p *MakeParser) ParseLine(raw string) (core.LogLine, bool) {
	line := core.LogLine{
		Timestamp: time.Now(),
		Raw:       raw,
		Level:     core.LogInfo,
	}
	trimmed := strings.TrimSpace(raw)

	if m := makeErrorRe.FindStringSubmatch(trimmed); m != nil {
		line.Level = core.LogError
		line.Parsed = &core.ParsedEntry{Message: m[2]}
		return line, true
	}
	if m := makeWarningRe.FindStringSubmatch(trimmed); m != nil {
		line.Level = core.LogWarning
		line.Parsed = &core.ParsedEntry{Message: m[1]}
		return line, true
	}
	if makeFailRe.MatchString(trimmed) {
		line.Level = core.LogError
		return line, true
	}
	if makeDirRe.MatchString(trimmed) {
		return line, true
	}
	return line, false
}

// Make doesn't expose reliable progress — always return false.
func (p *MakeParser) ParseProgress(raw string) (float64, bool) {
	return 0, false
}
