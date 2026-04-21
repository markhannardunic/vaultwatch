package notify

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

// ReplayStore holds a fixed-size window of past alerts that can be replayed
// to a Sender on demand (e.g. when a new notification channel comes online).
type ReplayStore struct {
	mu      sync.Mutex
	entries []replayEntry
	maxSize int
}

type replayEntry struct {
	path    string
	message string
	level   string
	ts      time.Time
}

// NewReplayStore creates a ReplayStore that retains up to maxSize entries.
func NewReplayStore(maxSize int) (*ReplayStore, error) {
	if maxSize <= 0 {
		return nil, errors.New("replay: maxSize must be greater than zero")
	}
	return &ReplayStore{maxSize: maxSize}, nil
}

// Record saves an alert to the store, evicting the oldest entry when full.
func (r *ReplayStore) Record(path, level, message string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if len(r.entries) >= r.maxSize {
		r.entries = r.entries[1:]
	}
	r.entries = append(r.entries, replayEntry{
		path:    path,
		message: message,
		level:   level,
		ts:      time.Now(),
	})
}

// Replay sends all stored entries to the provided Sender in chronological order.
// Errors from individual sends are collected and returned as a combined error.
func (r *ReplayStore) Replay(s Sender) error {
	r.mu.Lock()
	snap := make([]replayEntry, len(r.entries))
	copy(snap, r.entries)
	r.mu.Unlock()

	var errs []error
	for _, e := range snap {
		msg := fmt.Sprintf("[replay][%s] %s: %s", e.ts.Format(time.RFC3339), e.path, e.message)
		if err := s.Send(e.path, e.level, msg); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("replay: %d send error(s): %v", len(errs), errs)
	}
	return nil
}

// Len returns the current number of stored entries.
func (r *ReplayStore) Len() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.entries)
}

// Clear removes all stored entries.
func (r *ReplayStore) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.entries = nil
}
