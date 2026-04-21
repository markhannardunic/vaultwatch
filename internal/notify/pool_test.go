package notify

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestNewPooledSender_NilSender(t *testing.T) {
	_, err := NewPooledSender(nil, PoolDefaults())
	if err == nil {
		t.Fatal("expected error for nil sender")
	}
}

func TestNewPooledSender_InvalidWorkers(t *testing.T) {
	cfg := PoolConfig{Workers: 0, QueueSize: 8}
	_, err := NewPooledSender(&mockSender{}, cfg)
	if err == nil {
		t.Fatal("expected error for zero workers")
	}
}

func TestNewPooledSender_InvalidQueueSize(t *testing.T) {
	cfg := PoolConfig{Workers: 2, QueueSize: 0}
	_, err := NewPooledSender(&mockSender{}, cfg)
	if err == nil {
		t.Fatal("expected error for zero queue size")
	}
}

func TestPooledSender_Send_DeliveredByWorker(t *testing.T) {
	var count int64
	ms := &funcSender{fn: func(_ Message) error {
		atomic.AddInt64(&count, 1)
		return nil
	}}

	p, err := NewPooledSender(ms, PoolConfig{Workers: 2, QueueSize: 16})
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 5; i++ {
		if err := p.Send(Message{Path: "secret/a", Level: "warning", Body: "expiring"}); err != nil {
			t.Fatalf("Send error: %v", err)
		}
	}
	p.Stop()

	if atomic.LoadInt64(&count) != 5 {
		t.Fatalf("expected 5 deliveries, got %d", count)
	}
}

func TestPooledSender_Send_AfterStop_ReturnsError(t *testing.T) {
	p, _ := NewPooledSender(&mockSender{}, PoolDefaults())
	p.Stop()
	err := p.Send(Message{Path: "secret/x", Level: "critical", Body: "gone"})
	if err == nil {
		t.Fatal("expected error sending after stop")
	}
}

func TestPooledSender_Send_FullQueue_ReturnsError(t *testing.T) {
	var mu sync.Mutex
	mu.Lock() // block workers

	ms := &funcSender{fn: func(_ Message) error {
		mu.Lock()
		mu.Unlock()
		return nil
	}}

	p, _ := NewPooledSender(ms, PoolConfig{Workers: 1, QueueSize: 2})
	defer func() {
		mu.Unlock()
		p.Stop()
	}()

	// Fill the queue
	for i := 0; i < 3; i++ {
		_ = p.Send(Message{Path: "s", Level: "warning", Body: "x"})
	}

	// Next send should fail
	err := p.Send(Message{Path: "s", Level: "warning", Body: "overflow"})
	if err == nil {
		t.Fatal("expected queue full error")
	}
}

func TestPooledSender_Stop_Idempotent(t *testing.T) {
	p, _ := NewPooledSender(&mockSender{}, PoolDefaults())
	done := make(chan struct{})
	go func() {
		p.Stop()
		p.Stop()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Stop() deadlocked")
	}
}

// funcSender is a Sender backed by an arbitrary function.
type funcSender struct {
	fn func(Message) error
}

func (f *funcSender) Send(msg Message) error {
	return f.fn(msg)
}

// mockSender is a no-op Sender used in tests.
type mockSender struct{}

func (m *mockSender) Send(_ Message) error { return nil }

// errSender always returns an error.
type errSender struct{ err error }

func (e *errSender) Send(_ Message) error { return errors.New(e.err.Error()) }
