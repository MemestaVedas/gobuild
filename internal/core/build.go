package core

import (
	"os"
	"time"
)

// BuildState represents the lifecycle state of a build.
type BuildState int

const (
	StateIdle BuildState = iota
	StateQueued
	StateBuilding
	StateSuccess
	StateFailed
	StateCancelled
)

// Global messages
type SwitchToDashboardMsg struct{}

func (s BuildState) String() string {
	switch s {
	case StateIdle:
		return "idle"
	case StateQueued:
		return "queued"
	case StateBuilding:
		return "building"
	case StateSuccess:
		return "success"
	case StateFailed:
		return "failed"
	case StateCancelled:
		return "cancelled"
	default:
		return "unknown"
	}
}

// BuildTool identifies the build system being used.
type BuildTool int

const (
	ToolGeneric BuildTool = iota
	ToolCargo
	ToolNPM
	ToolMake
	ToolMSBuild
	ToolGo
	ToolGradle
	ToolCMake
)

func (t BuildTool) String() string {
	switch t {
	case ToolCargo:
		return "Cargo"
	case ToolNPM:
		return "NPM"
	case ToolMake:
		return "Make"
	case ToolMSBuild:
		return "MSBuild"
	case ToolGo:
		return "Go"
	case ToolGradle:
		return "Gradle"
	case ToolCMake:
		return "CMake"
	default:
		return "Generic"
	}
}

// LogLevel classifies a log line's severity.
type LogLevel int

const (
	LogInfo LogLevel = iota
	LogWarning
	LogError
	LogNote
)

func (l LogLevel) String() string {
	switch l {
	case LogWarning:
		return "warning"
	case LogError:
		return "error"
	case LogNote:
		return "note"
	default:
		return "info"
	}
}

// ParsedEntry holds structured data extracted from a raw log line.
type ParsedEntry struct {
	File    string
	Line    int
	Column  int
	Code    string
	Message string
}

// LogLine is a single line of build output with metadata.
type LogLine struct {
	Timestamp time.Time
	Level     LogLevel
	Raw       string
	Parsed    *ParsedEntry // nil if line could not be parsed
}

// BuildError represents a structured compiler/builder error.
type BuildError struct {
	File    string
	Line    int
	Column  int
	Code    string // e.g. "E0308" for Rust, "TS2345" for TypeScript
	Message string
	Level   LogLevel
}

// Build represents a single build invocation.
type Build struct {
	ID          string
	Name        string
	Tool        BuildTool
	Command     string
	WorkDir     string
	State       BuildState
	Progress    float64 // 0.0 – 1.0
	StartTime   time.Time
	EndTime     *time.Time // nil if still running
	Duration    time.Duration
	PID         int
	PTY         *os.File // Handle to the pseudoterminal for interactive io
	LogLines    []LogLine
	Errors      []BuildError
	Tags        []string
	ProfileName string
}

// Elapsed returns the current elapsed time for the build.
func (b *Build) Elapsed() time.Duration {
	if b.EndTime != nil {
		return b.Duration
	}
	if b.StartTime.IsZero() {
		return 0
	}
	return time.Since(b.StartTime)
}

// IsActive returns true if the build is currently running or queued.
func (b *Build) IsActive() bool {
	return b.State == StateBuilding || b.State == StateQueued
}

// StatusIcon returns a representative icon for the build's current state.
func (b *Build) StatusIcon() string {
	switch b.State {
	case StateSuccess:
		return "󰄬"
	case StateFailed:
		return "󰅖"
	case StateBuilding:
		return "󰦖"
	case StateQueued:
		return "󱞙"
	case StateCancelled:
		return "󰜺"
	default:
		return "󰔚"
	}
}

// ProcessInfo holds data about a discovered OS process.
type ProcessInfo struct {
	PID      int
	PPID     int
	Name     string
	CmdLine  string
	CWD      string
	CPUPct   float64
	MemBytes uint64
}

// FileEvent represents a filesystem change.
type FileEvent struct {
	Path string
	Op   string // "create", "write", "remove", "rename"
}

// SoundType for desktop audio alerts.
type SoundType int

const (
	SoundSuccess SoundType = iota
	SoundFailure
	SoundNotification
)
