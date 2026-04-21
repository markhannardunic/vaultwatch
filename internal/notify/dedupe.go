package notify

import (
	"crypto/sha256"
	"fmt"
	"sync"
	"time"
)

// DedupeStore tracks recently sent alert fingerprints to suppress duplicates.
type DedupeStore struct {
	mu      sync.Mutex
	seen    map[string]time.Time
	window  time.Duration
}

// NewDedupeStore creates a DedupeStore with the given deduplication window.
func NewDedupeStore(window time.Duration) *DedupeStore {
	if window <= 0 {
		window = 1 * time.Hour
	}
	return &DedupeStore{
		seen:   make(map[string]time.Time),
		window: window,
	}
}

// IsDuplicate returns true if the key was seen within the deduplication window.
// If not a duplicate, the key is recorded.
func (d *DedupeStore) IsDuplicate(key string) bool {
	d.mu.Lock()
	defer d.mu.Unlock()

	now := time.Now()
	if t, ok := d.seen[key]; ok && now.Sub(t) < d.window {
		return true
	}
	d.seen[key] = now
	return false
}

// Fingerprint generates a stable hash key from path and level.
func Fingerprint(path, level string) string {
	h := sha256.Sum256([]byte(fmt.Sprintf("%s|%s", path, level)))
	return fmt.Sprintf("%x", h[:8])
}

// Purge removes entries older than the deduplication window.
func (d *DedupeStore) Purge() {
	d.mu.Lock()
	defer d.mu.Unlock()

	now := time.Now()
	for k, t := range d.seen {
		if now.Sub(t) >= d.window {
			delete(d.seen, k)
		}
	}
}
