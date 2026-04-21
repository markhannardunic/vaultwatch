package notify

import (
	"fmt"
	"time"
)

// RetryConfig controls retry behaviour for notification senders.
type RetryConfig struct {
	MaxAttempts int
	Delay       time.Duration
	Multiplier  float64
}

// DefaultRetryConfig returns a sensible default retry configuration.
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts: 3,
		Delay:       2 * time.Second,
		Multiplier:  2.0,
	}
}

// RetrySender wraps a Sender and retries on failure using exponential backoff.
type RetrySender struct {
	inner  Sender
	config RetryConfig
	sleep  func(time.Duration)
}

// NewRetrySender creates a RetrySender wrapping the given Sender.
func NewRetrySender(inner Sender, cfg RetryConfig) *RetrySender {
	return &RetrySender{
		inner:  inner,
		config: cfg,
		sleep:  time.Sleep,
	}
}

// Send attempts to deliver the message, retrying up to MaxAttempts times.
func (r *RetrySender) Send(level, message string) error {
	delay := r.config.Delay
	var lastErr error

	for attempt := 1; attempt <= r.config.MaxAttempts; attempt++ {
		if err := r.inner.Send(level, message); err == nil {
			return nil
		} else {
			lastErr = err
		}

		if attempt < r.config.MaxAttempts {
			r.sleep(delay)
			delay = time.Duration(float64(delay) * r.config.Multiplier)
		}
	}

	return fmt.Errorf("send failed after %d attempts: %w", r.config.MaxAttempts, lastErr)
}
