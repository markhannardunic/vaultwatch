package notify

import (
	"context"
	"fmt"
	"time"
)

// TimeoutSender wraps a Sender and enforces a per-send deadline.
type TimeoutSender struct {
	sender  Sender
	timeout time.Duration
}

// NewTimeoutSender returns a TimeoutSender that cancels sends exceeding d.
// Returns an error if sender is nil or d is zero/negative.
func NewTimeoutSender(sender Sender, d time.Duration) (*TimeoutSender, error) {
	if sender == nil {
		return nil, fmt.Errorf("timeout: sender must not be nil")
	}
	if d <= 0 {
		return nil, fmt.Errorf("timeout: duration must be positive, got %s", d)
	}
	return &TimeoutSender{sender: sender, timeout: d}, nil
}

// Send calls the underlying sender within the configured timeout.
// If the deadline is exceeded, a wrapped context.DeadlineExceeded error is returned.
func (t *TimeoutSender) Send(path, message string) error {
	ctx, cancel := context.WithTimeout(context.Background(), t.timeout)
	defer cancel()

	type result struct{ err error }
	ch := make(chan result, 1)

	go func() {
		ch <- result{err: t.sender.Send(path, message)}
	}()

	select {
	case res := <-ch:
		return res.err
	case <-ctx.Done():
		return fmt.Errorf("timeout: send for %q exceeded %s: %w", path, t.timeout, context.DeadlineExceeded)
	}
}
