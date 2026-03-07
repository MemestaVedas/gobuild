package parser

import (
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/MemestaVedas/gobuild/internal/core"
)

// NPMParser parses NPM / Webpack / TypeScript build output.
type NPMParser struct{}

var (
	// "10% building 2/3 modules"
	webpackProgressRe = regexp.MustCompile(`^(\d+)%\s+\w+`)
	// "ERROR in ./src/index.js"
	npmErrorRe = regexp.MustCompile(`^ERROR in (.+)`)
	// "TS2345: Argument of type..."
	tsErrorRe = regexp.MustCompile(`TS(\d+):\s+(.+)`)
	// "Module not found: Error: ..."
	moduleNotFoundRe = regexp.MustCompile(`Module not found: (.+)`)
	// "warning  ..."
	npmWarnRe = regexp.MustCompile(`(?i)^(warning|warn)\s+(.+)`)
)

func (p *NPMParser) ToolName() string { return "npm" }

func (p *NPMParser) ParseLine(raw string) (core.LogLine, bool) {
	line := core.LogLine{
		Timestamp: time.Now(),
		Raw:       raw,
		Level:     core.LogInfo,
	}
	trimmed := strings.TrimSpace(raw)

	if m := npmErrorRe.FindStringSubmatch(trimmed); m != nil {
		line.Level = core.LogError
		line.Parsed = &core.ParsedEntry{File: m[1]}
		return line, true
	}
	if m := tsErrorRe.FindStringSubmatch(trimmed); m != nil {
		line.Level = core.LogError
		line.Parsed = &core.ParsedEntry{Code: "TS" + m[1], Message: m[2]}
		return line, true
	}
	if m := moduleNotFoundRe.FindStringSubmatch(trimmed); m != nil {
		line.Level = core.LogError
		line.Parsed = &core.ParsedEntry{Message: m[1]}
		return line, true
	}
	if m := npmWarnRe.FindStringSubmatch(trimmed); m != nil {
		line.Level = core.LogWarning
		line.Parsed = &core.ParsedEntry{Message: m[2]}
		return line, true
	}
	if webpackProgressRe.MatchString(trimmed) {
		return line, true
	}
	return line, false
}

func (p *NPMParser) ParseProgress(raw string) (float64, bool) {
	m := webpackProgressRe.FindStringSubmatch(strings.TrimSpace(raw))
	if m == nil {
		return 0, false
	}
	pct, err := strconv.ParseFloat(m[1], 64)
	if err != nil {
		return 0, false
	}
	return pct / 100.0, true
}
