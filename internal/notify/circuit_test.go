package notify

import (
	"errors"
	"testing"
	"time"
)

type countingSender struct {
	calls int
	err   error
}

func (c *countingSender) Send(_, _ string) error {
	c.calls++
	return c.err
}

func TestNewCircuitBreaker_NilSender(t *testing.T) {
	_, err := NewCircuitBreaker(nil, 3, time.Second)
	if err == nil {
		t.Fatal("expected error for nil sender")
	}
}

func TestNewCircuitBreaker_InvalidThreshold(t *testing.T) {
	s := &countingSender{}
	_, err := NewCircuitBreaker(s, 0, time.Second)
	if err == nil {
		t.Fatal("expected error for zero threshold")
	}
}

func TestNewCircuitBreaker_InvalidRecovery(t *testing.T) {
	s := &countingSender{}
	_, err := NewCircuitBreaker(s, 2, 0)
	if err == nil {
		t.Fatal("expected error for zero recovery")
	}
}

func TestCircuitBreaker_ClosedOnSuccess(t *testing.T) {
	s := &countingSender{}
	cb, _ := NewCircuitBreaker(s, 2, time.Second)

	if err := cb.Send("secret/foo", "msg"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cb.State() != CircuitClosed {
		t.Errorf("expected closed, got %d", cb.State())
	}
}

func TestCircuitBreaker_OpensAfterThreshold(t *testing.T) {
	s := &countingSender{err: errors.New("boom")}
	cb, _ := NewCircuitBreaker(s, 3, time.Minute)

	for i := 0; i < 3; i++ {
		_ = cb.Send("secret/foo", "msg")
	}
	if cb.State() != CircuitOpen {
		t.Errorf("expected open, got %d", cb.State())
	}
}

func TestCircuitBreaker_RejectsWhenOpen(t *testing.T) {
	s := &countingSender{err: errors.New("boom")}
	cb, _ := NewCircuitBreaker(s, 1, time.Minute)

	_ = cb.Send("secret/foo", "msg") // trips open

	err := cb.Send("secret/foo", "msg")
	if !errors.Is(err, ErrCircuitOpen) {
		t.Errorf("expected ErrCircuitOpen, got %v", err)
	}
	if s.calls != 1 {
		t.Errorf("expected sender called once, got %d", s.calls)
	}
}

func TestCircuitBreaker_HalfOpenAfterRecovery(t *testing.T) {
	s := &countingSender{err: errors.New("boom")}
	cb, _ := NewCircuitBreaker(s, 1, time.Millisecond)

	_ = cb.Send("secret/foo", "msg") // trips open
	time.Sleep(5 * time.Millisecond)

	// Should attempt (half-open), fail again, re-open
	err := cb.Send("secret/foo", "msg")
	if err == nil || errors.Is(err, ErrCircuitOpen) {
		t.Errorf("expected sender error in half-open, got %v", err)
	}
	if cb.State() != CircuitOpen {
		t.Errorf("expected open after half-open failure, got %d", cb.State())
	}
}

func TestCircuitBreaker_Reset_AllowsSend(t *testing.T) {
	s := &countingSender{err: errors.New("boom")}
	cb, _ := NewCircuitBreaker(s, 1, time.Minute)
	_ = cb.Send("secret/foo", "msg")

	cb.Reset()
	if cb.State() != CircuitClosed {
		t.Errorf("expected closed after reset")
	}

	s.err = nil
	if err := cb.Send("secret/foo", "msg"); err != nil {
		t.Errorf("unexpected error after reset: %v", err)
	}
}
