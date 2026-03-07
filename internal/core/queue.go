package core

import "sync"

// BuildQueue manages a dependency-ordered queue of pending builds.
type BuildQueue struct {
	mu    sync.Mutex
	items []*Build
}

// NewBuildQueue creates a new empty queue.
func NewBuildQueue() *BuildQueue {
	return &BuildQueue{}
}

// Enqueue adds a build to the end of the queue.
func (q *BuildQueue) Enqueue(b *Build) {
	q.mu.Lock()
	defer q.mu.Unlock()
	b.State = StateQueued
	q.items = append(q.items, b)
}

// Dequeue pops the next build from the front of the queue.
func (q *BuildQueue) Dequeue() (*Build, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if len(q.items) == 0 {
		return nil, false
	}
	b := q.items[0]
	q.items = q.items[1:]
	return b, true
}

// Len returns the number of queued builds.
func (q *BuildQueue) Len() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return len(q.items)
}

// Remove removes a build by ID from the queue.
func (q *BuildQueue) Remove(id string) bool {
	q.mu.Lock()
	defer q.mu.Unlock()
	for i, b := range q.items {
		if b.ID == id {
			q.items = append(q.items[:i], q.items[i+1:]...)
			return true
		}
	}
	return false
}

// Snapshot returns a copy of the current queue.
func (q *BuildQueue) Snapshot() []*Build {
	q.mu.Lock()
	defer q.mu.Unlock()
	out := make([]*Build, len(q.items))
	copy(out, q.items)
	return out
}
