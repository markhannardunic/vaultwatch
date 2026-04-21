package notify

import (
	"errors"
	"testing"
	"time"
)

func TestNewHealthChecker_InvalidInterval(t *testing.T) {
	_, err := NewHealthChecker(0)
	if err == nil {
		t.Fatal("expected error for zero interval")
	}
}

func TestNewHealthChecker_ValidInterval(t *testing.T) {
	hc, err := NewHealthChecker(time.Second)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hc == nil {
		t.Fatal("expected non-nil HealthChecker")
	}
}

func TestHealthChecker_Register_AppearsInSnapshot(t *testing.T) {
	hc, _ := NewHealthChecker(time.Minute)
	hc.Register("slack", func() error { return nil })

	snap := hc.Snapshot()
	if len(snap) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(snap))
	}
	if snap[0].Name != "slack" {
		t.Errorf("expected name 'slack', got %q", snap[0].Name)
	}
	if !snap[0].Healthy {
		t.Error("expected initial state to be healthy")
	}
}

func TestHealthChecker_RunAll_MarksUnhealthy(t *testing.T) {
	hc, _ := NewHealthChecker(time.Minute)
	probeErr := errors.New("connection refused")
	hc.Register("pagerduty", func() error { return probeErr })

	hc.runAll()

	snap := hc.Snapshot()
	if len(snap) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(snap))
	}
	if snap[0].Healthy {
		t.Error("expected sender to be marked unhealthy")
	}
	if snap[0].LastError == nil {
		t.Error("expected LastError to be set")
	}
	if snap[0].LastCheck.IsZero() {
		t.Error("expected LastCheck to be stamped")
	}
}

func TestHealthChecker_RunAll_MarksHealthy(t *testing.T) {
	hc, _ := NewHealthChecker(time.Minute)
	hc.Register("webhook", func() error { return nil })
	hc.runAll()

	snap := hc.Snapshot()
	if !snap[0].Healthy {
		t.Error("expected sender to be marked healthy")
	}
	if snap[0].LastError != nil {
		t.Errorf("expected no error, got %v", snap[0].LastError)
	}
}

func TestHealthChecker_MultipleProbes(t *testing.T) {
	hc, _ := NewHealthChecker(time.Minute)
	hc.Register("ok", func() error { return nil })
	hc.Register("bad", func() error { return errors.New("fail") })
	hc.runAll()

	snap := hc.Snapshot()
	if len(snap) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(snap))
	}
	healthy := 0
	for _, s := range snap {
		if s.Healthy {
			healthy++
		}
	}
	if healthy != 1 {
		t.Errorf("expected 1 healthy sender, got %d", healthy)
	}
}
