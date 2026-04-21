package notify

import (
	"fmt"
	"time"
)

// RateLimitedRouter wraps a Router and enforces per-path rate limits.
type RateLimitedRouter struct {
	inner   *Router
	limiter *RateLimiter
}

// NewRateLimitedRouter creates a RateLimitedRouter with the given inner router,
// max sends per window, and window duration.
func NewRateLimitedRouter(inner *Router, max int, window time.Duration) (*RateLimitedRouter, error) {
	limiter, err := NewRateLimiter(max, window)
	if err != nil {
		return nil, fmt.Errorf("ratelimited_router: %w", err)
	}
	return &RateLimitedRouter{
		inner:   inner,
		limiter: limiter,
	}, nil
}

// Dispatch sends the alert only if the rate limit for the secret path allows it.
func (r *RateLimitedRouter) Dispatch(level, path, message string) error {
	key := path + ":" + level
	if !r.limiter.Allow(key) {
		return nil
	}
	return r.inner.Dispatch(level, path, message)
}
