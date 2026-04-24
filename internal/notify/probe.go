package notify

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// ProbeResult holds the outcome of a single sender probe.
type ProbeResult struct {
	Name      string
	Reachable bool
	Latency   time.Duration
	Err       error
	CheckedAt time.Time
}

// Prober runs connectivity probes against named senders.
type Prober struct {
	mu      sync.RWMutex
	senders map[string]Sender
	last    map[string]ProbeResult
	timeout time.Duration
}

// NewProber creates a Prober with the given per-probe timeout.
func NewProber(timeout time.Duration) (*Prober, error) {
	if timeout <= 0 {
		return nil, fmt.Errorf("probe timeout must be positive")
	}
	return &Prober{
		senders: make(map[string]Sender),
		last:    make(map[string]ProbeResult),
		timeout: timeout,
	}, nil
}

// Register adds a named sender to the probe set.
func (p *Prober) Register(name string, s Sender) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.senders[name] = s
}

// RunAll probes every registered sender and stores results.
func (p *Prober) RunAll(ctx context.Context) []ProbeResult {
	p.mu.Lock()
	defer p.mu.Unlock()

	results := make([]ProbeResult, 0, len(p.senders))
	for name, s := range p.senders {
		start := time.Now()
		ctx2, cancel := context.WithTimeout(ctx, p.timeout)
		err := s.Send(ctx2, "__probe__", "ping")
		cancel()
		r := ProbeResult{
			Name:      name,
			Reachable: err == nil,
			Latency:   time.Since(start),
			Err:       err,
			CheckedAt: time.Now(),
		}
		p.last[name] = r
		results = append(results, r)
	}
	return results
}

// Snapshot returns the most recent result for every registered sender.
func (p *Prober) Snapshot() map[string]ProbeResult {
	p.mu.RLock()
	defer p.mu.RUnlock()
	out := make(map[string]ProbeResult, len(p.last))
	for k, v := range p.last {
		out[k] = v
	}
	return out
}
