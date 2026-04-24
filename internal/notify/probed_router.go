package notify

import (
	"context"
	"fmt"
	"time"
)

// ProbedRouter wraps a Router and gates dispatch behind a connectivity probe.
// If the most recent probe result for the target sender is unreachable the
// message is rejected immediately rather than timing out inside the sender.
type ProbedRouter struct {
	router  *Router
	prober  *Prober
	timeout time.Duration
}

// NewProbedRouter creates a ProbedRouter.
// probeTimeout is used when RunAll is called inside Dispatch.
func NewProbedRouter(r *Router, p *Prober, probeTimeout time.Duration) (*ProbedRouter, error) {
	if r == nil {
		return nil, fmt.Errorf("router must not be nil")
	}
	if p == nil {
		return nil, fmt.Errorf("prober must not be nil")
	}
	if probeTimeout <= 0 {
		return nil, fmt.Errorf("probeTimeout must be positive")
	}
	return &ProbedRouter{router: r, prober: p, timeout: probeTimeout}, nil
}

// Dispatch probes the named sender, then forwards to the underlying Router.
// If no probe result exists the message is forwarded unconditionally.
func (pr *ProbedRouter) Dispatch(ctx context.Context, level, path, message string) error {
	snap := pr.prober.Snapshot()
	if result, ok := snap[path]; ok && !result.Reachable {
		return fmt.Errorf("sender %q is unreachable (last checked %s): %w",
			path, result.CheckedAt.Format(time.RFC3339), result.Err)
	}
	return pr.router.Dispatch(ctx, level, path, message)
}

// Probe triggers an immediate probe run and returns the results.
func (pr *ProbedRouter) Probe(ctx context.Context) []ProbeResult {
	return pr.prober.RunAll(ctx)
}
