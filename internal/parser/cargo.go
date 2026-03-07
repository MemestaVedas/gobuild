package parser

import (
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/MemestaVedas/gobuild/internal/core"
)

// CargoParser parses Cargo (Rust) build output.
type CargoParser struct{}

var (
	// "Compiling foo v1.0.0 (5/12)"  or  "Compiling foo v1.0.0 (/path)"
	cargoProgressRe = regexp.MustCompile(`\((\d+)/(\d+)\)`)
	// "error[E0308]: mismatched types"
	cargoErrorCodeRe = regexp.MustCompile(`^error\[([A-Z]\d+)\]:\s+(.+)`)
	// "error: mismatched types" (no code)
	cargoErrorRe = regexp.MustCompile(`^error:\s+(.+)`)
	// " --> src/main.rs:23:5"
	cargoLocationRe = regexp.MustCompile(`-->\s+(.+):(\d+):(\d+)`)
	// "warning: unused variable `x`"
	cargoWarningRe = regexp.MustCompile(`^warning:\s+(.+)`)
	// "note: ..."
	cargoNoteRe = regexp.MustCompile(`^note:\s+(.+)`)
	// "Compiling foo v1.0.0" / "Checking foo" / "Finished ..." / "error"
	compilingRe = regexp.MustCompile(`^\s*(Compiling|Checking|Finished|Running|Downloading|Updating|Fresh)\s+`)
)

func (p *CargoParser) ToolName() string { return "cargo" }

func (p *CargoParser) ParseLine(raw string) (core.LogLine, bool) {
	line := core.LogLine{
		Timestamp: time.Now(),
		Raw:       raw,
		Level:     core.LogInfo,
	}

	trimmed := strings.TrimSpace(raw)

	if m := cargoErrorCodeRe.FindStringSubmatch(trimmed); m != nil {
		line.Level = core.LogError
		line.Parsed = &core.ParsedEntry{Code: m[1], Message: m[2]}
		return line, true
	}
	if m := cargoErrorRe.FindStringSubmatch(trimmed); m != nil {
		line.Level = core.LogError
		line.Parsed = &core.ParsedEntry{Message: m[1]}
		return line, true
	}
	if m := cargoLocationRe.FindStringSubmatch(trimmed); m != nil {
		ln, _ := strconv.Atoi(m[2])
		col, _ := strconv.Atoi(m[3])
		line.Parsed = &core.ParsedEntry{File: m[1], Line: ln, Column: col}
		return line, true
	}
	if m := cargoWarningRe.FindStringSubmatch(trimmed); m != nil {
		line.Level = core.LogWarning
		line.Parsed = &core.ParsedEntry{Message: m[1]}
		return line, true
	}
	if m := cargoNoteRe.FindStringSubmatch(trimmed); m != nil {
		line.Level = core.LogNote
		line.Parsed = &core.ParsedEntry{Message: m[1]}
		return line, true
	}
	if compilingRe.MatchString(trimmed) {
		return line, true
	}
	return line, false
}

func (p *CargoParser) ParseProgress(raw string) (float64, bool) {
	m := cargoProgressRe.FindStringSubmatch(raw)
	if m == nil {
		return 0, false
	}
	cur, err1 := strconv.ParseFloat(m[1], 64)
	total, err2 := strconv.ParseFloat(m[2], 64)
	if err1 != nil || err2 != nil || total == 0 {
		return 0, false
	}
	return cur / total, true
}
