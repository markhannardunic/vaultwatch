package notify

import (
	"sync"
	"time"
)

// ThrottleKey uniquely identifies a notification target+level combination.
type ThrottleKey struct {
	Target string
	Level  string
}

// Throttler suppresses duplicate notifications within a cooldown window.
type Throttler struct {
	mu       sync.Mutex
	cooldown time.Duration
	lastSent map[ThrottleKey]time.Time
	now      func() time.Time
}

// NewThrottler creates a Throttler with the given cooldown duration.
func NewThrottler(cooldown time.Duration) *Throttler {
	return &Throttler{
		cooldown: cooldown,
		lastSent: make(map[ThrottleKey]time.Time),
		now:      time.Now,
	}
}

// Allow returns true if a notification for the given key should be sent,
// and records the send time if so.
func (t *Throttler) Allow(key ThrottleKey) bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	now := t.now()
	if last, ok := t.lastSent[key]; ok {
		if now.Sub(last) < t.cooldown {
			return false
		}
	}
	t.lastSent[key] = now
	return true
}

// Reset clears the throttle state for a specific key.
func (t *Throttler) Reset(key ThrottleKey) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.lastSent, key)
}

// ResetAll clears all throttle state.
func (t *Throttler) ResetAll() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.lastSent = make(map[ThrottleKey]time.Time)
}
