package notify

import (
	"fmt"
	"time"
)

// ThrottledRouter wraps a Router with per-target, per-level send throttling.
type ThrottledRouter struct {
	router    *Router
	throttler *Throttler
}

// NewThrottledRouter creates a ThrottledRouter with the given cooldown.
func NewThrottledRouter(router *Router, cooldown time.Duration) *ThrottledRouter {
	return &ThrottledRouter{
		router:    router,
		throttler: NewThrottler(cooldown),
	}
}

// Dispatch sends an alert through the underlying router only if the
// target+level combination has not been sent within the cooldown window.
// It returns a map of target names to any errors encountered.
func (tr *ThrottledRouter) Dispatch(level, message string) map[string]error {
	errs := make(map[string]error)

	for target, sender := range tr.router.senders {
		if !tr.router.matches(target, level) {
			continue
		}
		key := ThrottleKey{Target: target, Level: level}
		if !tr.throttler.Allow(key) {
			continue
		}
		if err := sender.Send(level, message); err != nil {
			tr.throttler.Reset(key)
			errs[target] = fmt.Errorf("send to %s failed: %w", target, err)
		}
	}

	return errs
}

// ResetThrottle clears all throttle state, allowing immediate resend.
func (tr *ThrottledRouter) ResetThrottle() {
	tr.throttler.ResetAll()
}
