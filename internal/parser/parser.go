package parser

import "github.com/MemestaVedas/gobuild/internal/core"

// LogParser parses raw build output for a specific tool.
type LogParser interface {
	ToolName() string
	// ParseLine parses one raw line and returns a LogLine.
	// Returns (line, true) if the line was meaningfully parsed, (line, false) otherwise.
	ParseLine(raw string) (core.LogLine, bool)
	// ParseProgress extracts a fractional progress value (0.0–1.0) from a line.
	// Returns (progress, true) if a progress value was found.
	ParseProgress(raw string) (float64, bool)
}

// Registry maps tool names to their parsers.
var Registry = map[core.BuildTool]LogParser{}

func init() {
	Registry[core.ToolCargo] = &CargoParser{}
	Registry[core.ToolNPM] = &NPMParser{}
	Registry[core.ToolMake] = &MakeParser{}
	Registry[core.ToolGeneric] = &GenericParser{}
	Registry[core.ToolGo] = &GoParser{}
	Registry[core.ToolGeneric] = &GenericParser{}
}

// For returns the parser for a given tool, falling back to Generic.
func For(tool core.BuildTool) LogParser {
	if p, ok := Registry[tool]; ok {
		return p
	}
	return Registry[core.ToolGeneric]
}
