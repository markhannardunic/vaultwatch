package notify

import (
	"strings"
	"sync"
	"testing"
	"time"
)

type captureDigestSender struct {
	mu       sync.Mutex
	messages []string
}

func (c *captureDigestSender) Send(_, message string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.messages = append(c.messages, message)
	return nil
}

func TestNewDigestSender_NilSender(t *testing.T) {
	_, err := NewDigestSender(nil, time.Second, "")
	if err == nil {
		t.Fatal("expected error for nil sender")
	}
}

func TestNewDigestSender_ZeroWindow(t *testing.T) {
	cap := &captureDigestSender{}
	_, err := NewDigestSender(cap, 0, "")
	if err == nil {
		t.Fatal("expected error for zero window")
	}
}

func TestDigestSender_FlushOnExplicitCall(t *testing.T) {
	cap := &captureDigestSender{}
	ds, err := NewDigestSender(cap, 10*time.Second, "Test Header")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_ = ds.Send("secret/a", "warning: expires soon")
	_ = ds.Send("secret/b", "critical: expires today")
	_ = ds.Flush()

	cap.mu.Lock()
	defer cap.mu.Unlock()
	if len(cap.messages) != 1 {
		t.Fatalf("expected 1 digest message, got %d", len(cap.messages))
	}
	body := cap.messages[0]
	if !strings.Contains(body, "Test Header") {
		t.Errorf("digest missing header, got: %s", body)
	}
	if !strings.Contains(body, "secret/a") || !strings.Contains(body, "secret/b") {
		t.Errorf("digest missing entries, got: %s", body)
	}
}

func TestDigestSender_FlushOnWindowExpiry(t *testing.T) {
	cap := &captureDigestSender{}
	ds, err := NewDigestSender(cap, 50*time.Millisecond, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_ = ds.Send("secret/x", "expiring")

	time.Sleep(150 * time.Millisecond)

	cap.mu.Lock()
	defer cap.mu.Unlock()
	if len(cap.messages) != 1 {
		t.Fatalf("expected 1 digest after window, got %d", len(cap.messages))
	}
	if !strings.Contains(cap.messages[0], "VaultWatch Digest") {
		t.Errorf("expected default header in digest, got: %s", cap.messages[0])
	}
}

func TestDigestSender_EmptyFlush_NoSend(t *testing.T) {
	cap := &captureDigestSender{}
	ds, _ := NewDigestSender(cap, time.Second, "")
	_ = ds.Flush()

	cap.mu.Lock()
	defer cap.mu.Unlock()
	if len(cap.messages) != 0 {
		t.Errorf("expected no messages on empty flush, got %d", len(cap.messages))
	}
}
