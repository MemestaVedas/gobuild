package core

import "sync"

// BuildManager owns all active builds, protected by RWMutex.
type BuildManager struct {
	mu     sync.RWMutex
	builds map[string]*Build // keyed by Build.ID
}

// NewBuildManager creates an initialised BuildManager.
func NewBuildManager() *BuildManager {
	return &BuildManager{
		builds: make(map[string]*Build),
	}
}

// Add registers a new build.
func (bm *BuildManager) Add(b *Build) {
	bm.mu.Lock()
	defer bm.mu.Unlock()
	bm.builds[b.ID] = b
}

// Get returns a build by ID.
func (bm *BuildManager) Get(id string) (*Build, bool) {
	bm.mu.RLock()
	defer bm.mu.RUnlock()
	b, ok := bm.builds[id]
	return b, ok
}

// All returns a snapshot of all builds.
func (bm *BuildManager) All() []*Build {
	bm.mu.RLock()
	defer bm.mu.RUnlock()
	out := make([]*Build, 0, len(bm.builds))
	for _, b := range bm.builds {
		out = append(out, b)
	}
	return out
}

// Active returns only builds that are running or queued.
func (bm *BuildManager) Active() []*Build {
	bm.mu.RLock()
	defer bm.mu.RUnlock()
	var out []*Build
	for _, b := range bm.builds {
		if b.IsActive() {
			out = append(out, b)
		}
	}
	return out
}

// Remove deletes a build record.
func (bm *BuildManager) Remove(id string) {
	bm.mu.Lock()
	defer bm.mu.Unlock()
	delete(bm.builds, id)
}

// Update applies a mutator function to an existing build safely.
func (bm *BuildManager) Update(id string, fn func(*Build)) bool {
	bm.mu.Lock()
	defer bm.mu.Unlock()
	b, ok := bm.builds[id]
	if !ok {
		return false
	}
	fn(b)
	return true
}

// AppendLog adds a log line to a build, trimming to maxLines if needed.
func (bm *BuildManager) AppendLog(id string, line LogLine, maxLines int) {
	bm.mu.Lock()
	defer bm.mu.Unlock()
	b, ok := bm.builds[id]
	if !ok {
		return
	}
	b.LogLines = append(b.LogLines, line)
	if maxLines > 0 && len(b.LogLines) > maxLines {
		b.LogLines = b.LogLines[len(b.LogLines)-maxLines:]
	}
}

// AppendError adds a build error to a build.
func (bm *BuildManager) AppendError(id string, err BuildError) {
	bm.mu.Lock()
	defer bm.mu.Unlock()
	b, ok := bm.builds[id]
	if !ok {
		return
	}
	b.Errors = append(b.Errors, err)
}

// Count returns the total number of tracked builds.
func (bm *BuildManager) Count() int {
	bm.mu.RLock()
	defer bm.mu.RUnlock()
	return len(bm.builds)
}
