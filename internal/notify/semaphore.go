package notify

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// Semaphore limits the number of concurrent Send calls to an underlying Sender.
type Semaphore struct {
	sender  Sender
	tokens  chan struct{}
	timeout time.Duration
}

// NewSemaphore wraps sender and restricts concurrent sends to maxConcurrent.
// An optional acquireTimeout controls how long to wait for a slot before
// returning an error; pass 0 for no timeout.
func NewSemaphore(sender Sender, maxConcurrent int, acquireTimeout time.Duration) (*Semaphore, error) {
	if sender == nil {
		return nil, errors.New("semaphore: sender must not be nil")
	}
	if maxConcurrent <= 0 {
		return nil, fmt.Errorf("semaphore: maxConcurrent must be > 0, got %d", maxConcurrent)
	}
	tokens := make(chan struct{}, maxConcurrent)
	for i := 0; i < maxConcurrent; i++ {
		tokens <- struct{}{}
	}
	return &Semaphore{
		sender:  sender,
		tokens:  tokens,
		timeout: acquireTimeout,
	}, nil
}

// Send acquires a concurrency slot, delegates to the wrapped sender, then
// releases the slot. If acquireTimeout > 0 and no slot is available within
// that duration, Send returns an error without calling the underlying sender.
func (s *Semaphore) Send(path, message string) error {
	if s.timeout > 0 {
		ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
		defer cancel()
		select {
		case <-s.tokens:
		case <-ctx.Done():
			return fmt.Errorf("semaphore: timed out waiting for concurrency slot for %q", path)
		}
	} else {
		<-s.tokens
	}
	defer func() { s.tokens <- struct{}{} }()
	return s.sender.Send(path, message)
}
