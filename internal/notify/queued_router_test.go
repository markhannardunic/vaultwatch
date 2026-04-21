package notify_test

import (
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/yourusername/vaultwatch/internal/notify"
)

func makeQueuedRouter(t *testing.T, maxSize int, workers int) (*notify.QueuedRouter, *recordingSender) {
	t.Helper()
	rs := &recordingSender{}
	router, err := notify.NewQueuedRouter(rs, notify.QueueConfig{
		MaxSize:     maxSize,
		Workers:     workers,
		DrainOnStop: true,
	})
	if err != nil {
		t.Fatalf("NewQueuedRouter: %v", err)
	}
	return router, rs
}

func TestQueuedRouter_Dispatch_EnqueuesAndDelivers(t *testing.T) {
	router, rs := makeQueuedRouter(t, 10, 1)
	router.Start()
	defer router.Stop()

	if err := router.Dispatch("secret/foo", "warning", "expiring soon"); err != nil {
		t.Fatalf("Dispatch returned error: %v", err)
	}

	// Allow worker goroutine time to process
	deadline := time.Now().Add(500 * time.Millisecond)
	for time.Now().Before(deadline) {
		if rs.count() >= 1 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	if got := rs.count(); got != 1 {
		t.Errorf("expected 1 delivery, got %d", got)
	}
}

func TestQueuedRouter_Dispatch_MultipleMessages(t *testing.T) {
	const n = 5
	router, rs := makeQueuedRouter(t, 20, 2)
	router.Start()
	defer router.Stop()

	for i := 0; i < n; i++ {
		path := fmt.Sprintf("secret/item%d", i)
		if err := router.Dispatch(path, "warning", "expiring"); err != nil {
			t.Fatalf("Dispatch[%d] error: %v", i, err)
		}
	}

	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		if rs.count() >= n {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	if got := rs.count(); got != n {
		t.Errorf("expected %d deliveries, got %d", n, got)
	}
}

func TestQueuedRouter_Stop_DrainsPending(t *testing.T) {
	router, rs := makeQueuedRouter(t, 50, 1)
	router.Start()

	const n = 10
	for i := 0; i < n; i++ {
		_ = router.Dispatch(fmt.Sprintf("secret/drain%d", i), "critical", "expired")
	}

	router.Stop() // should drain before returning

	if got := rs.count(); got != n {
		t.Errorf("after drain stop: expected %d, got %d", n, got)
	}
}

func TestQueuedRouter_Dispatch_AfterStop_ReturnsError(t *testing.T) {
	router, _ := makeQueuedRouter(t, 10, 1)
	router.Start()
	router.Stop()

	err := router.Dispatch("secret/foo", "warning", "late")
	if err == nil {
		t.Error("expected error dispatching after stop, got nil")
	}
}

func TestQueuedRouter_FullQueue_ReturnsError(t *testing.T) {
	// Use 0 workers so nothing drains; queue fills immediately.
	rs := &recordingSender{delay: 50 * time.Millisecond}
	router, err := notify.NewQueuedRouter(rs, notify.QueueConfig{
		MaxSize:     1,
		Workers:     0,
		DrainOnStop: false,
	})
	if err != nil {
		t.Fatalf("NewQueuedRouter: %v", err)
	}
	router.Start()
	defer router.Stop()

	_ = router.Dispatch("secret/a", "warning", "fill the queue")
	err = router.Dispatch("secret/b", "warning", "should overflow")
	if err == nil {
		t.Error("expected overflow error, got nil")
	}
}

// recordingSender is a thread-safe Sender that records calls.
type recordingSender struct {
	mu    sync.Mutex
	calls []string
	delay time.Duration
	err   error
}

func (r *recordingSender) Send(path, level, message string) error {
	if r.delay > 0 {
		time.Sleep(r.delay)
	}
	if r.err != nil {
		return r.err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.calls = append(r.calls, fmt.Sprintf("%s|%s|%s", path, level, message))
	return nil
}

func (r *recordingSender) count() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.calls)
}

var _ = errors.New // ensure errors import is used if err field is exercised
