package notify

import (
	"fmt"
	"sync"
	"time"
)

// AckStatus represents the acknowledgement state of a dispatched alert.
type AckStatus int

const (
	AckPending AckStatus = iota
	AckAcknowledged
	AckExpired
)

// AckRecord holds the state of a single acknowledgeable alert.
type AckRecord struct {
	Path      string
	Level     string
	IssuedAt  time.Time
	AckedAt   *time.Time
	ExpiresAt time.Time
	Status    AckStatus
}

// AckStore tracks acknowledgements for dispatched alerts.
type AckStore struct {
	mu      sync.Mutex
	records map[string]*AckRecord
	ttl     time.Duration
}

// NewAckStore creates a new AckStore with the given TTL for pending acks.
func NewAckStore(ttl time.Duration) (*AckStore, error) {
	if ttl <= 0 {
		return nil, fmt.Errorf("ack store: ttl must be positive, got %s", ttl)
	}
	return &AckStore{
		records: make(map[string]*AckRecord),
		ttl:     ttl,
	}, nil
}

// Track registers a new alert path as pending acknowledgement.
func (a *AckStore) Track(path, level string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	now := time.Now()
	a.records[path] = &AckRecord{
		Path:      path,
		Level:     level,
		IssuedAt:  now,
		ExpiresAt: now.Add(a.ttl),
		Status:    AckPending,
	}
}

// Acknowledge marks an alert as acknowledged. Returns an error if not found or expired.
func (a *AckStore) Acknowledge(path string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	rec, ok := a.records[path]
	if !ok {
		return fmt.Errorf("ack store: no record for path %q", path)
	}
	if time.Now().After(rec.ExpiresAt) {
		rec.Status = AckExpired
		return fmt.Errorf("ack store: ack window expired for path %q", path)
	}
	now := time.Now()
	rec.AckedAt = &now
	rec.Status = AckAcknowledged
	return nil
}

// Snapshot returns a copy of all current records.
func (a *AckStore) Snapshot() []AckRecord {
	a.mu.Lock()
	defer a.mu.Unlock()
	out := make([]AckRecord, 0, len(a.records))
	for _, r := range a.records {
		out = append(out, *r)
	}
	return out
}

// Purge removes expired records from the store.
func (a *AckStore) Purge() {
	a.mu.Lock()
	defer a.mu.Unlock()
	now := time.Now()
	for k, r := range a.records {
		if r.Status == AckPending && now.After(r.ExpiresAt) {
			r.Status = AckExpired
		}
		if r.Status == AckExpired || r.Status == AckAcknowledged {
			delete(a.records, k)
		}
	}
}
