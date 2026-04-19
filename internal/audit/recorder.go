package audit

import (
	"sync"
	"time"
)

// Record holds an in-memory snapshot of audit entries for the current run.
type Record struct {
	mu      sync.Mutex
	entries []Entry
}

// NewRecord creates an empty Record.
func NewRecord() *Record {
	return &Record{}
}

// Add appends an entry, stamping the timestamp if absent.
func (r *Record) Add(e Entry) {
	if e.Timestamp.IsZero() {
		e.Timestamp = time.Now().UTC()
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.entries = append(r.entries, e)
}

// Entries returns a copy of all recorded entries.
func (r *Record) Entries() []Entry {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]Entry, len(r.entries))
	copy(out, r.entries)
	return out
}

// CountByLevel returns a map of level -> count.
func (r *Record) CountByLevel() map[string]int {
	r.mu.Lock()
	defer r.mu.Unlock()
	counts := make(map[string]int)
	for _, e := range r.entries {
		counts[e.Level]++
	}
	return counts
}
