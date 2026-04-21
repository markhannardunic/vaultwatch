package notify

import (
	"sync"
	"testing"
	"time"
)

type captureSender struct {
	mu   sync.Mutex
	msgs []Message
}

func (c *captureSender) Send(msg Message) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.msgs = append(c.msgs, msg)
	return nil
}

func (c *captureSender) count() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.msgs)
}

func TestNewBufferedSender_NilSender(t *testing.T) {
	_, err := NewBufferedSender(nil, BufferDefaults())
	if err == nil {
		t.Fatal("expected error for nil sender")
	}
}

func TestNewBufferedSender_InvalidMaxSize(t *testing.T) {
	cfg := BufferDefaults()
	cfg.MaxSize = 0
	_, err := NewBufferedSender(&captureSender{}, cfg)
	if err == nil {
		t.Fatal("expected error for MaxSize=0")
	}
}

func TestNewBufferedSender_InvalidFlushEvery(t *testing.T) {
	cfg := BufferDefaults()
	cfg.FlushEvery = 0
	_, err := NewBufferedSender(&captureSender{}, cfg)
	if err == nil {
		t.Fatal("expected error for FlushEvery=0")
	}
}

func TestBufferedSender_FlushOnMaxSize(t *testing.T) {
	cap := &captureSender{}
	cfg := BufferConfig{MaxSize: 3, FlushEvery: time.Hour, DropOnFull: false}
	b, err := NewBufferedSender(cap, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer b.Stop()

	for i := 0; i < 3; i++ {
		_ = b.Send(Message{Path: "secret/a", Level: "warning", Body: "expiring"})
	}
	// 4th send triggers flush of the full buffer first
	_ = b.Send(Message{Path: "secret/b", Level: "critical", Body: "expired"})

	if cap.count() < 3 {
		t.Errorf("expected at least 3 messages flushed, got %d", cap.count())
	}
}

func TestBufferedSender_ExplicitFlush(t *testing.T) {
	cap := &captureSender{}
	cfg := BufferConfig{MaxSize: 50, FlushEvery: time.Hour}
	b, err := NewBufferedSender(cap, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer b.Stop()

	_ = b.Send(Message{Path: "secret/x", Level: "warning", Body: "soon"})
	b.Flush()

	if cap.count() != 1 {
		t.Errorf("expected 1 message after flush, got %d", cap.count())
	}
}

func TestBufferedSender_DropOnFull(t *testing.T) {
	cap := &captureSender{}
	cfg := BufferConfig{MaxSize: 2, FlushEvery: time.Hour, DropOnFull: true}
	b, err := NewBufferedSender(cap, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer b.Stop()

	for i := 0; i < 5; i++ {
		_ = b.Send(Message{Path: "secret/z", Level: "warning", Body: "drop"})
	}
	b.Flush()

	if cap.count() > 2 {
		t.Errorf("expected at most 2 messages with DropOnFull, got %d", cap.count())
	}
}

func TestBufferedSender_FlushOnInterval(t *testing.T) {
	cap := &captureSender{}
	cfg := BufferConfig{MaxSize: 50, FlushEvery: 20 * time.Millisecond}
	b, err := NewBufferedSender(cap, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer b.Stop()

	_ = b.Send(Message{Path: "secret/t", Level: "critical", Body: "urgent"})
	time.Sleep(60 * time.Millisecond)

	if cap.count() < 1 {
		t.Error("expected message flushed by interval ticker")
	}
}
