package parser

import (
	"regexp"
	"strings"
	"time"

	"github.com/MemestaVedas/gobuild/internal/core"
)

// GenericParser is a fallback for unknown build tools.
// It detects error/warning keywords case-insensitively.
type GenericParser struct{}

var (
	genericErrorRe = regexp.MustCompile(`(?i)(error|fatal|failed):\s+(.+)`)
	genericWarnRe  = regexp.MustCompile(`(?i)warning:\s+(.+)`)
	genericNoteRe  = regexp.MustCompile(`(?i)note:\s+(.+)`)
)

func (p *GenericParser) ToolName() string { return "generic" }

func (p *GenericParser) ParseLine(raw string) (core.LogLine, bool) {
	line := core.LogLine{
		Timestamp: time.Now(),
		Raw:       raw,
		Level:     core.LogInfo,
	}
	trimmed := strings.TrimSpace(raw)

	if m := genericErrorRe.FindStringSubmatch(trimmed); m != nil {
		line.Level = core.LogError
		line.Parsed = &core.ParsedEntry{Message: m[2]}
		return line, true
	}
	if m := genericWarnRe.FindStringSubmatch(trimmed); m != nil {
		line.Level = core.LogWarning
		line.Parsed = &core.ParsedEntry{Message: m[1]}
		return line, true
	}
	if m := genericNoteRe.FindStringSubmatch(trimmed); m != nil {
		line.Level = core.LogNote
		line.Parsed = &core.ParsedEntry{Message: m[1]}
		return line, true
	}
	return line, false
}

// GenericParser cannot determine progress.
func (p *GenericParser) ParseProgress(raw string) (float64, bool) {
	return 0, false
}
