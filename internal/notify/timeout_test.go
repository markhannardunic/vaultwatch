package notify

import (
	"errors"
	"testing"
	"time"
)

type slowSender struct {
	delay time.Duration
	err   error
}

func (s *slowSender) Send(_, _ string) error {
	time.Sleep(s.delay)
	return s.err
}

func TestNewTimeoutSender_NilSender(t *testing.T) {
	_, err := NewTimeoutSender(nil, time.Second)
	if err == nil {
		t.Fatal("expected error for nil sender")
	}
}

func TestNewTimeoutSender_ZeroDuration(t *testing.T) {
	s := &slowSender{}
	_, err := NewTimeoutSender(s, 0)
	if err == nil {
		t.Fatal("expected error for zero duration")
	}
}

func TestNewTimeoutSender_NegativeDuration(t *testing.T) {
	s := &slowSender{}
	_, err := NewTimeoutSender(s, -time.Second)
	if err == nil {
		t.Fatal("expected error for negative duration")
	}
}

func TestTimeoutSender_Send_Success(t *testing.T) {
	s := &slowSender{delay: 1 * time.Millisecond}
	ts, err := NewTimeoutSender(s, 100*time.Millisecond)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := ts.Send("secret/foo", "expiring soon"); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestTimeoutSender_Send_Timeout(t *testing.T) {
	s := &slowSender{delay: 200 * time.Millisecond}
	ts, err := NewTimeoutSender(s, 20*time.Millisecond)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	err = ts.Send("secret/bar", "expiring soon")
	if err == nil {
		t.Fatal("expected timeout error")
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected DeadlineExceeded, got: %v", err)
	}
}

func TestTimeoutSender_Send_PropagatesError(t *testing.T) {
	expected := errors.New("downstream failure")
	s := &slowSender{delay: 1 * time.Millisecond, err: expected}
	ts, err := NewTimeoutSender(s, 100*time.Millisecond)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := ts.Send("secret/baz", "msg"); !errors.Is(got, expected) {
		t.Fatalf("expected %v, got %v", expected, got)
	}
}
