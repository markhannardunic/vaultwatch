package notify

import (
	"errors"
	"sync"
	"time"
)

// CircuitState represents the state of a circuit breaker.
type CircuitState int

const (
	CircuitClosed CircuitState = iota
	CircuitOpen
	CircuitHalfOpen
)

// ErrCircuitOpen is returned when a send is attempted while the circuit is open.
var ErrCircuitOpen = errors.New("circuit breaker is open")

// CircuitBreaker wraps a Sender and trips open after a threshold of consecutive failures.
type CircuitBreaker struct {
	sender     Sender
	threshold  int
	recovery   time.Duration

	mu         sync.Mutex
	failures   int
	state      CircuitState
	openedAt   time.Time
}

// NewCircuitBreaker creates a CircuitBreaker that opens after threshold consecutive
// failures and attempts recovery after the given recovery duration.
func NewCircuitBreaker(sender Sender, threshold int, recovery time.Duration) (*CircuitBreaker, error) {
	if sender == nil {
		return nil, errors.New("circuit breaker: sender must not be nil")
	}
	if threshold < 1 {
		return nil, errors.New("circuit breaker: threshold must be at least 1")
	}
	if recovery <= 0 {
		return nil, errors.New("circuit breaker: recovery duration must be positive")
	}
	return &CircuitBreaker{
		sender:    sender,
		threshold: threshold,
		recovery:  recovery,
		state:     CircuitClosed,
	}, nil
}

// Send forwards the message if the circuit is closed or half-open.
// A successful send resets the breaker; a failure may trip it open.
func (cb *CircuitBreaker) Send(path, message string) error {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case CircuitOpen:
		if time.Since(cb.openedAt) >= cb.recovery {
			cb.state = CircuitHalfOpen
		} else {
			return ErrCircuitOpen
		}
	case CircuitClosed, CircuitHalfOpen:
		// proceed
	}

	err := cb.sender.Send(path, message)
	if err != nil {
		cb.failures++
		if cb.failures >= cb.threshold || cb.state == CircuitHalfOpen {
			cb.state = CircuitOpen
			cb.openedAt = time.Now()
		}
		return err
	}

	cb.failures = 0
	cb.state = CircuitClosed
	return nil
}

// State returns the current circuit state.
func (cb *CircuitBreaker) State() CircuitState {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.state
}

// Reset forces the circuit back to closed with zero failures.
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.failures = 0
	cb.state = CircuitClosed
}
