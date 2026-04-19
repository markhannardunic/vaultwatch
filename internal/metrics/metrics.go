package metrics

import (
	"sync"
	"time"
)

// Level represents a secret health level.
type Level string

const (
	LevelHealthy  Level = "healthy"
	LevelWarning  Level = "warning"
	LevelCritical Level = "critical"
)

// Counter holds counts of secrets checked per level.
type Counter struct {
	mu        sync.Mutex
	counts    map[Level]int
	LastRun   time.Time
	TotalRuns int
}

// New returns an initialized Counter.
func New() *Counter {
	return &Counter{
		counts: map[Level]int{
			LevelHealthy:  0,
			LevelWarning:  0,
			LevelCritical: 0,
		},
	}
}

// Record increments the counter for the given level.
func (c *Counter) Record(level Level) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.counts[level]++
}

// Snapshot returns a copy of current counts and marks a run.
func (c *Counter) Snapshot() map[Level]int {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.LastRun = time.Now()
	c.TotalRuns++
	copy := make(map[Level]int, len(c.counts))
	for k, v := range c.counts {
		copy[k] = v
	}
	return copy
}

// Reset clears all counters.
func (c *Counter) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()
	for k := range c.counts {
		c.counts[k] = 0
	}
}
