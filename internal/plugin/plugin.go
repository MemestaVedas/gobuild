package plugin

import (
	"sync"

	"github.com/MemestaVedas/gobuild/internal/core"
)

// Plugin is the base interface that all plugins must implement.
type Plugin interface {
	Name() string
	Version() string
	OnLoad() error
	OnUnload() error
}

// Event handler interfaces — implement only the ones you need.
type BuildStartHandler interface {
	OnBuildStart(b *core.Build)
}

type BuildEndHandler interface {
	OnBuildEnd(b *core.Build)
}

type BuildErrorHandler interface {
	OnBuildError(b *core.Build, err core.BuildError)
}

type LogLineHandler interface {
	OnLogLine(b *core.Build, line core.LogLine)
}

// EventBus dispatches events to all loaded plugins asynchronously.
type EventBus struct {
	mu      sync.RWMutex
	plugins []Plugin
}

// NewEventBus creates a new plugin bus.
func NewEventBus() *EventBus {
	return &EventBus{
		plugins: make([]Plugin, 0),
	}
}

// Register adds an active plugin to the broadcast list.
func (eb *EventBus) Register(p Plugin) {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	eb.plugins = append(eb.plugins, p)
}

// EmitBuildStart triggers OnBuildStart asynchronously.
func (eb *EventBus) EmitBuildStart(b *core.Build) {
	eb.mu.RLock()
	defer eb.mu.RUnlock()
	for _, p := range eb.plugins {
		if h, ok := p.(BuildStartHandler); ok {
			go h.OnBuildStart(b)
		}
	}
}

// EmitBuildEnd triggers OnBuildEnd asynchronously.
func (eb *EventBus) EmitBuildEnd(b *core.Build) {
	eb.mu.RLock()
	defer eb.mu.RUnlock()
	for _, p := range eb.plugins {
		if h, ok := p.(BuildEndHandler); ok {
			go h.OnBuildEnd(b)
		}
	}
}

// EmitBuildError triggers OnBuildError asynchronously.
func (eb *EventBus) EmitBuildError(b *core.Build, err core.BuildError) {
	eb.mu.RLock()
	defer eb.mu.RUnlock()
	for _, p := range eb.plugins {
		if h, ok := p.(BuildErrorHandler); ok {
			go h.OnBuildError(b, err)
		}
	}
}

// EmitLogLine triggers OnLogLine asynchronously.
func (eb *EventBus) EmitLogLine(b *core.Build, line core.LogLine) {
	eb.mu.RLock()
	defer eb.mu.RUnlock()
	for _, p := range eb.plugins {
		if h, ok := p.(LogLineHandler); ok {
			go h.OnLogLine(b, line)
		}
	}
}

// Loaded returns a list of loaded plugin names.
func (eb *EventBus) Loaded() []string {
	eb.mu.RLock()
	defer eb.mu.RUnlock()
	var names []string
	for _, p := range eb.plugins {
		names = append(names, p.Name())
	}
	return names
}
