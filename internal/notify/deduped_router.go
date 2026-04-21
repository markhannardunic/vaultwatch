package notify

import (
	"fmt"
	"time"
)

// DedupedRouter wraps a Router and suppresses duplicate alerts within a time window.
type DedupedRouter struct {
	router *Router
	store  *DedupeStore
}

// NewDedupedRouter creates a DedupedRouter with the given underlying router and window.
func NewDedupedRouter(router *Router, window time.Duration) *DedupedRouter {
	return &DedupedRouter{
		router: router,
		store:  NewDedupeStore(window),
	}
}

// Dispatch sends the alert only if it has not been sent within the deduplication window.
// Returns an error if the alert is a duplicate or if the underlying dispatch fails.
func (d *DedupedRouter) Dispatch(level, path, message string) error {
	key := Fingerprint(path, level)
	if d.store.IsDuplicate(key) {
		return fmt.Errorf("dedupe: suppressed duplicate alert for path=%s level=%s", path, level)
	}
	return d.router.Dispatch(level, message)
}

// PurgeExpired removes stale deduplication entries to prevent unbounded memory growth.
func (d *DedupedRouter) PurgeExpired() {
	d.store.Purge()
}
