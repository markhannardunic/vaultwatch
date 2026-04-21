package notify

import (
	"errors"
	"testing"
	"time"
)

type callCountSender struct {
	calls    int
	failUntil int
	err      error
}

func (c *callCountSender) Send(level, message string) error {
	c.calls++
	if c.calls <= c.failUntil {
		return c.err
	}
	return nil
}

func noSleep(_ time.Duration) {}

func TestRetrySender_SucceedsFirstAttempt(t *testing.T) {
	inner := &callCountSender{failUntil: 0}
	rs := NewRetrySender(inner, DefaultRetryConfig())
	rs.sleep = noSleep

	if err := rs.Send("warning", "test"); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if inner.calls != 1 {
		t.Errorf("expected 1 call, got %d", inner.calls)
	}
}

func TestRetrySender_RetriesAndSucceeds(t *testing.T) {
	inner := &callCountSender{failUntil: 2, err: errors.New("transient")}
	rs := NewRetrySender(inner, DefaultRetryConfig())
	rs.sleep = noSleep

	if err := rs.Send("critical", "msg"); err != nil {
		t.Fatalf("expected success after retries, got %v", err)
	}
	if inner.calls != 3 {
		t.Errorf("expected 3 calls, got %d", inner.calls)
	}
}

func TestRetrySender_ExhaustsRetries(t *testing.T) {
	sentinel := errors.New("permanent failure")
	inner := &callCountSender{failUntil: 99, err: sentinel}
	cfg := RetryConfig{MaxAttempts: 3, Delay: time.Millisecond, Multiplier: 1.0}
	rs := NewRetrySender(inner, cfg)
	rs.sleep = noSleep

	err := rs.Send("critical", "msg")
	if err == nil {
		t.Fatal("expected error after exhausting retries")
	}
	if !errors.Is(err, sentinel) {
		t.Errorf("expected wrapped sentinel error, got %v", err)
	}
	if inner.calls != 3 {
		t.Errorf("expected exactly 3 attempts, got %d", inner.calls)
	}
}

func TestRetrySender_SleepCalledBetweenAttempts(t *testing.T) {
	inner := &callCountSender{failUntil: 2, err: errors.New("fail")}
	cfg := RetryConfig{MaxAttempts: 3, Delay: 100 * time.Millisecond, Multiplier: 2.0}
	rs := NewRetrySender(inner, cfg)

	var sleptDurations []time.Duration
	rs.sleep = func(d time.Duration) { sleptDurations = append(sleptDurations, d) }

	_ = rs.Send("warning", "msg")

	if len(sleptDurations) != 2 {
		t.Fatalf("expected 2 sleeps, got %d", len(sleptDurations))
	}
	if sleptDurations[0] != 100*time.Millisecond {
		t.Errorf("first sleep: expected 100ms, got %v", sleptDurations[0])
	}
	if sleptDurations[1] != 200*time.Millisecond {
		t.Errorf("second sleep: expected 200ms, got %v", sleptDurations[1])
	}
}
