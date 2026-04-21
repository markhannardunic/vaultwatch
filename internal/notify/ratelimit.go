package notify

import (
	"fmt"
	"sync"
	"time"
)

// RateLimiter enforces a maximum number of notifications per window per key.
type RateLimiter struct {
	mu      sync.Mutex
	counts  map[string][]time.Time
	max     int
	window  time.Duration
	nowFunc func() time.Time
}

// NewRateLimiter creates a RateLimiter allowing at most max sends per window.
func NewRateLimiter(max int, window time.Duration) (*RateLimiter, error) {
	if max <= 0 {
		return nil, fmt.Errorf("ratelimit: max must be greater than zero")
	}
	if window <= 0 {
		return nil, fmt.Errorf("ratelimit: window must be greater than zero")
	}
	return &RateLimiter{
		counts:  make(map[string][]time.Time),
		max:     max,
		window:  window,
		nowFunc: time.Now,
	}, nil
}

// Allow returns true if the key is within the allowed rate, false if exceeded.
func (r *RateLimiter) Allow(key string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := r.nowFunc()
	cutoff := now.Add(-r.window)

	times := r.counts[key]
	filtered := times[:0]
	for _, t := range times {
		if t.After(cutoff) {
			filtered = append(filtered, t)
		}
	}

	if len(filtered) >= r.max {
		r.counts[key] = filtered
		return false
	}

	r.counts[key] = append(filtered, now)
	return true
}

// Reset clears the rate limit state for the given key.
func (r *RateLimiter) Reset(key string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.counts, key)
}
