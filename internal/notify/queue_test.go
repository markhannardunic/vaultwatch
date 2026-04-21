package notify_test

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/youorg/vaultwatch/internal/notify"
)

type countingSender struct {
	mu    sync.Mutex
	count int32
	err   error
}

func (c *countingSender) Send(_ notify.Message) error {
	atomic.AddInt32(&c.count, 1)
	return c.err
}

func TestNewQueuedSender_NilSender(t *testing.T) {
	_, err := notify.NewQueuedSender(nil, notify.QueueDefaults())
	if err == nil {
		t.Fatal("expected error for nil sender")
	}
}

func TestNewQueuedSender_InvalidMaxSize(t *testing.T) {
	cfg := notify.QueueDefaults()
	cfg.MaxSize = 0
	_, err := notify.NewQueuedSender(&countingSender{}, cfg)
	if err == nil {
		t.Fatal("expected error for zero MaxSize")
	}
}

func TestQueuedSender_Send_Enqueues(t *testing.T) {
	cs := &countingSender{}
	cfg := notify.QueueDefaults()
	cfg.Workers = 1
	qs, err := notify.NewQueuedSender(cs, cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	qs.Start()

	msg := notify.Message{Path: "secret/test", Level: "warning", Text: "expiring"}
	if err := qs.Send(msg); err != nil {
		t.Fatalf("unexpected send error: %v", err)
	}

	qs.Stop(2 * time.Second)
	if atomic.LoadInt32(&cs.count) != 1 {
		t.Errorf("expected 1 delivery, got %d", cs.count)
	}
}

func TestQueuedSender_Send_FullQueue_ReturnsError(t *testing.T) {
	cs := &countingSender{err: errors.New("fail")}
	cfg := notify.QueueConfig{MaxSize: 1, Workers: 1, DrainOnStop: false}
	qs, _ := notify.NewQueuedSender(cs, cfg)
	// Do NOT start workers so queue fills immediately.

	msg := notify.Message{Path: "secret/x", Level: "critical", Text: "soon"}
	_ = qs.Send(msg) // fills slot
	err := qs.Send(msg) // should fail
	if err == nil {
		t.Fatal("expected error when queue is full")
	}
}

func TestQueuedSender_DrainOnStop(t *testing.T) {
	cs := &countingSender{}
	cfg := notify.QueueConfig{MaxSize: 10, Workers: 1, DrainOnStop: true}
	qs, _ := notify.NewQueuedSender(cs, cfg)
	qs.Start()

	for i := 0; i < 5; i++ {
		_ = qs.Send(notify.Message{Path: "secret/drain", Level: "warning", Text: "t"})
	}
	qs.Stop(3 * time.Second)
	if atomic.LoadInt32(&cs.count) != 5 {
		t.Errorf("expected 5 deliveries after drain, got %d", cs.count)
	}
}
