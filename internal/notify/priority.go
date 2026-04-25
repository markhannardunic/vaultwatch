package notify

import (
	"fmt"
	"strings"
	"sync"
)

// Priority represents a numeric dispatch priority (lower = higher priority).
type Priority int

const (
	PriorityCritical Priority = 1
	PriorityWarning  Priority = 2
	PriorityInfo     Priority = 3
)

// PriorityQueue dispatches to a Sender but enforces priority ordering across
// registered levels. Messages at a higher priority always flush before lower
// priority ones when Flush is called explicitly.
type PriorityQueue struct {
	mu     sync.Mutex
	buckets map[Priority][]string
	order   []Priority
	sender  Sender
}

// NewPriorityQueue creates a PriorityQueue wrapping the given Sender.
// Returns an error if sender is nil.
func NewPriorityQueue(sender Sender) (*PriorityQueue, error) {
	if sender == nil {
		return nil, fmt.Errorf("priority queue: sender must not be nil")
	}
	return &PriorityQueue{
		buckets: make(map[Priority][]string),
		order:   []Priority{PriorityCritical, PriorityWarning, PriorityInfo},
		sender:  sender,
	}, nil
}

// Enqueue adds a message to the bucket for the given priority level.
func (pq *PriorityQueue) Enqueue(p Priority, message string) {
	pq.mu.Lock()
	defer pq.mu.Unlock()
	pq.buckets[p] = append(pq.buckets[p], message)
}

// Flush drains all buckets in priority order, sending each message via the
// underlying Sender. Returns the first error encountered; remaining messages
// in the current bucket are skipped on error.
func (pq *PriorityQueue) Flush() error {
	pq.mu.Lock()
	defer pq.mu.Unlock()

	for _, p := range pq.order {
		for _, msg := range pq.buckets[p] {
			if err := pq.sender.Send(msg); err != nil {
				return fmt.Errorf("priority queue flush at priority %d: %w", p, err)
			}
		}
		delete(pq.buckets, p)
	}
	return nil
}

// LevelToPriority maps a string alert level to a Priority constant.
func LevelToPriority(level string) Priority {
	switch strings.ToLower(level) {
	case "critical":
		return PriorityCritical
	case "warning":
		return PriorityWarning
	default:
		return PriorityInfo
	}
}
