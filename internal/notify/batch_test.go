package notify

import (
	"strings"
	"sync"
	"testing"
	"time"
)

type captureSender struct {
	mu       sync.Mutex
	messages []string
}

func (c *captureSender) Send(path, message string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.messages = append(c.messages, message)
	return nil
}

func (c *captureSender) last() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.messages) == 0 {
		return ""
	}
	return c.messages[len(c.messages)-1]
}

func TestNewBatchSender_NilSender(t *testing.T) {
	_, err := NewBatchSender(BatchConfig{}, nil)
	if err == nil {
		t.Fatal("expected error for nil sender")
	}
}

func TestBatchSender_FlushOnMaxSize(t *testing.T) {
	cap := &captureSender{}
	bs, err := NewBatchSender(BatchConfig{MaxSize: 2, MaxWait: 10 * time.Second}, cap)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = bs.Send("secret/a", "warning")
	if len(cap.messages) != 0 {
		t.Fatal("expected no flush before max size")
	}
	_ = bs.Send("secret/b", "critical")
	if len(cap.messages) != 1 {
		t.Fatalf("expected flush after max size, got %d dispatches", len(cap.messages))
	}
	if !strings.Contains(cap.last(), "secret/a") || !strings.Contains(cap.last(), "secret/b") {
		t.Errorf("combined message missing entries: %s", cap.last())
	}
}

func TestBatchSender_ExplicitFlush(t *testing.T) {
	cap := &captureSender{}
	bs, _ := NewBatchSender(BatchConfig{MaxSize: 5, MaxWait: 10 * time.Second}, cap)
	_ = bs.Send("secret/x", "expiring soon")
	if len(cap.messages) != 0 {
		t.Fatal("expected no flush before explicit call")
	}
	_ = bs.Flush("batch/manual")
	if len(cap.messages) != 1 {
		t.Fatalf("expected one dispatch after flush, got %d", len(cap.messages))
	}
}

func TestBatchSender_FlushEmpty_NoDispatch(t *testing.T) {
	cap := &captureSender{}
	bs, _ := NewBatchSender(BatchConfig{MaxSize: 5, MaxWait: 10 * time.Second}, cap)
	_ = bs.Flush("batch/empty")
	if len(cap.messages) != 0 {
		t.Fatal("expected no dispatch when buffer is empty")
	}
}

func TestBatchSender_TimerFlush(t *testing.T) {
	cap := &captureSender{}
	bs, _ := NewBatchSender(BatchConfig{MaxSize: 10, MaxWait: 50 * time.Millisecond}, cap)
	_ = bs.Send("secret/timer", "about to expire")
	time.Sleep(150 * time.Millisecond)
	cap.mu.Lock()
	count := len(cap.messages)
	cap.mu.Unlock()
	if count == 0 {
		t.Fatal("expected timer-triggered flush")
	}
}
