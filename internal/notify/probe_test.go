package notify

import (
	"context"
	"errors"
	"testing"
	"time"
)

type stubSender struct {
	err error
	calls int
}

func (s *stubSender) Send(_ context.Context, _, _ string) error {
	s.calls++
	return s.err
}

func TestNewProber_InvalidTimeout(t *testing.T) {
	_, err := NewProber(0)
	if err == nil {
		t.Fatal("expected error for zero timeout")
	}
}

func TestNewProber_ValidTimeout(t *testing.T) {
	p, err := NewProber(time.Second)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p == nil {
		t.Fatal("expected non-nil prober")
	}
}

func TestProber_RunAll_ReachableSender(t *testing.T) {
	p, _ := NewProber(time.Second)
	p.Register("ok", &stubSender{})

	results := p.RunAll(context.Background())
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if !results[0].Reachable {
		t.Error("expected sender to be reachable")
	}
	if results[0].Name != "ok" {
		t.Errorf("unexpected name: %s", results[0].Name)
	}
}

func TestProber_RunAll_UnreachableSender(t *testing.T) {
	p, _ := NewProber(time.Second)
	p.Register("bad", &stubSender{err: errors.New("connection refused")})

	results := p.RunAll(context.Background())
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Reachable {
		t.Error("expected sender to be unreachable")
	}
	if results[0].Err == nil {
		t.Error("expected non-nil error")
	}
}

func TestProber_Snapshot_ReflectsLastRun(t *testing.T) {
	p, _ := NewProber(time.Second)
	p.Register("s1", &stubSender{})
	p.RunAll(context.Background())

	snap := p.Snapshot()
	if _, ok := snap["s1"]; !ok {
		t.Error("expected s1 in snapshot")
	}
}

func TestProber_Snapshot_EmptyBeforeRun(t *testing.T) {
	p, _ := NewProber(time.Second)
	p.Register("s1", &stubSender{})

	snap := p.Snapshot()
	if len(snap) != 0 {
		t.Errorf("expected empty snapshot before run, got %d entries", len(snap))
	}
}
