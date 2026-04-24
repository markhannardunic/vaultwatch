package notify

import (
	"fmt"
	"sync"
	"time"
)

// DeadLetterEntry holds a message that failed all delivery attempts.
type DeadLetterEntry struct {
	Path      string
	Message   string
	Level     string
	Err       error
	FailedAt  time.Time
	Attempts  int
}

// DeadLetterStore records messages that could not be delivered.
type DeadLetterStore struct {
	mu      sync.Mutex
	entries []DeadLetterEntry
	maxSize int
}

// NewDeadLetterStore creates a DeadLetterStore with the given capacity.
// maxSize must be greater than zero.
func NewDeadLetterStore(maxSize int) (*DeadLetterStore, error) {
	if maxSize <= 0 {
		return nil, fmt.Errorf("deadletter: maxSize must be greater than zero")
	}
	return &DeadLetterStore{maxSize: maxSize}, nil
}

// Record adds a failed message to the store. If the store is full the
// oldest entry is evicted to make room.
func (d *DeadLetterStore) Record(path, message, level string, attempts int, err error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if len(d.entries) >= d.maxSize {
		d.entries = d.entries[1:]
	}
	d.entries = append(d.entries, DeadLetterEntry{
		Path:     path,
		Message:  message,
		Level:    level,
		Err:      err,
		FailedAt: time.Now(),
		Attempts: attempts,
	})
}

// Snapshot returns a copy of all dead-letter entries.
func (d *DeadLetterStore) Snapshot() []DeadLetterEntry {
	d.mu.Lock()
	defer d.mu.Unlock()
	out := make([]DeadLetterEntry, len(d.entries))
	copy(out, d.entries)
	return out
}

// Len returns the current number of entries.
func (d *DeadLetterStore) Len() int {
	d.mu.Lock()
	defer d.mu.Unlock()
	return len(d.entries)
}

// Clear removes all entries from the store.
func (d *DeadLetterStore) Clear() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.entries = d.entries[:0]
}
