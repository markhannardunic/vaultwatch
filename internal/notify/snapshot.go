package notify

import (
	"sync"
	"time"
)

// DispatchRecord holds the result of a single dispatch attempt.
type DispatchRecord struct {
	Path      string
	Level     string
	SentAt    time.Time
	Success   bool
	Err       error
}

// SnapshotStore records recent dispatch outcomes for observability.
type SnapshotStore struct {
	mu      sync.Mutex
	records []DispatchRecord
	maxSize int
}

// NewSnapshotStore creates a SnapshotStore that retains up to maxSize records.
func NewSnapshotStore(maxSize int) *SnapshotStore {
	if maxSize <= 0 {
		maxSize = 100
	}
	return &SnapshotStore{maxSize: maxSize}
}

// Record appends a dispatch outcome, evicting the oldest if at capacity.
func (s *SnapshotStore) Record(r DispatchRecord) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.records) >= s.maxSize {
		s.records = s.records[1:]
	}
	s.records = append(s.records, r)
}

// Snapshot returns a copy of all recorded dispatch outcomes.
func (s *SnapshotStore) Snapshot() []DispatchRecord {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]DispatchRecord, len(s.records))
	copy(out, s.records)
	return out
}

// CountBySuccess returns the number of successful and failed dispatches.
func (s *SnapshotStore) CountBySuccess() (ok int, failed int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, r := range s.records {
		if r.Success {
			ok++
		} else {
			failed++
		}
	}
	return
}

// Clear removes all stored records.
func (s *SnapshotStore) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.records = s.records[:0]
}
