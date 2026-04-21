package notify

import (
	"fmt"
	"sync"
	"time"
)

// SenderHealth tracks the health state of a named sender.
type SenderHealth struct {
	Name      string
	Healthy   bool
	LastCheck time.Time
	LastError error
}

// HealthChecker periodically probes registered senders and records their health.
type HealthChecker struct {
	mu      sync.RWMutex
	states  map[string]*SenderHealth
	probes  map[string]func() error
	interval time.Duration
	stop    chan struct{}
}

// NewHealthChecker creates a HealthChecker with the given probe interval.
func NewHealthChecker(interval time.Duration) (*HealthChecker, error) {
	if interval <= 0 {
		return nil, fmt.Errorf("healthcheck: interval must be positive, got %v", interval)
	}
	return &HealthChecker{
		states:   make(map[string]*SenderHealth),
		probes:   make(map[string]func() error),
		interval: interval,
		stop:     make(chan struct{}),
	}, nil
}

// Register adds a named probe function to the checker.
func (h *HealthChecker) Register(name string, probe func() error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.probes[name] = probe
	h.states[name] = &SenderHealth{Name: name, Healthy: true}
}

// Start begins periodic health checks in a background goroutine.
func (h *HealthChecker) Start() {
	go func() {
		ticker := time.NewTicker(h.interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				h.runAll()
			case <-h.stop:
				return
			}
		}
	}()
}

// Stop halts background health checks.
func (h *HealthChecker) Stop() {
	close(h.stop)
}

// Snapshot returns a copy of all current health states.
func (h *HealthChecker) Snapshot() []SenderHealth {
	h.mu.RLock()
	defer h.mu.RUnlock()
	out := make([]SenderHealth, 0, len(h.states))
	for _, s := range h.states {
		out = append(out, *s)
	}
	return out
}

func (h *HealthChecker) runAll() {
	h.mu.Lock()
	defer h.mu.Unlock()
	for name, probe := range h.probes {
		err := probe()
		s := h.states[name]
		s.LastCheck = time.Now()
		s.LastError = err
		s.Healthy = err == nil
	}
}
