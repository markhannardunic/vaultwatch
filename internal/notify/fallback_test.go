package notify

import (
	"errors"
	"testing"
)

type mockFallbackSender struct {
	called bool
	err    error
}

func (m *mockFallbackSender) Send(_, _, _ string) error {
	m.called = true
	return m.err
}

func TestNewFallbackRouter_NilPrimary(t *testing.T) {
	_, err := NewFallbackRouter(nil, &mockFallbackSender{})
	if err == nil {
		t.Fatal("expected error for nil primary")
	}
}

func TestNewFallbackRouter_NoFallbacks(t *testing.T) {
	_, err := NewFallbackRouter(&mockFallbackSender{})
	if err == nil {
		t.Fatal("expected error when no fallbacks provided")
	}
}

func TestNewFallbackRouter_NilFallback(t *testing.T) {
	_, err := NewFallbackRouter(&mockFallbackSender{}, nil)
	if err == nil {
		t.Fatal("expected error for nil fallback sender")
	}
}

func TestFallbackRouter_PrimarySucceeds_FallbackNotCalled(t *testing.T) {
	primary := &mockFallbackSender{}
	fb := &mockFallbackSender{}

	r, err := NewFallbackRouter(primary, fb)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := r.Dispatch("warning", "secret/foo", "expires soon"); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if !primary.called {
		t.Error("expected primary to be called")
	}
	if fb.called {
		t.Error("expected fallback NOT to be called")
	}
}

func TestFallbackRouter_PrimaryFails_FallbackCalled(t *testing.T) {
	primary := &mockFallbackSender{err: errors.New("primary down")}
	fb := &mockFallbackSender{}

	r, _ := NewFallbackRouter(primary, fb)

	if err := r.Dispatch("critical", "secret/bar", "expired"); err != nil {
		t.Fatalf("expected fallback to succeed, got: %v", err)
	}
	if !fb.called {
		t.Error("expected fallback to be called")
	}
}

func TestFallbackRouter_AllFail_ReturnsError(t *testing.T) {
	primary := &mockFallbackSender{err: errors.New("primary down")}
	fb1 := &mockFallbackSender{err: errors.New("fb1 down")}
	fb2 := &mockFallbackSender{err: errors.New("fb2 down")}

	r, _ := NewFallbackRouter(primary, fb1, fb2)

	err := r.Dispatch("critical", "secret/baz", "expired")
	if err == nil {
		t.Fatal("expected error when all senders fail")
	}
	if !fb1.called || !fb2.called {
		t.Error("expected all fallbacks to be attempted")
	}
}
