package notify

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type slowSender struct {
	delay   time.Duration
	called  int32
	peakCon int32
	current int32
}

func (s *slowSender) Send(_, _ string) error {
	new := atomic.AddInt32(&s.current, 1)
	for {
		peak := atomic.LoadInt32(&s.peakCon)
		if new <= peak || atomic.CompareAndSwapInt32(&s.peakCon, peak, new) {
			break
		}
	}
	time.Sleep(s.delay)
	atomic.AddInt32(&s.current, -1)
	atomic.AddInt32(&s.called, 1)
	return nil
}

func TestNewSemaphore_NilSender(t *testing.T) {
	_, err := NewSemaphore(nil, 2, 0)
	if err == nil {
		t.Fatal("expected error for nil sender")
	}
}

func TestNewSemaphore_InvalidMax(t *testing.T) {
	snd := &slowSender{}
	_, err := NewSemaphore(snd, 0, 0)
	if err == nil {
		t.Fatal("expected error for maxConcurrent=0")
	}
}

func TestSemaphore_LimitsConcurrency(t *testing.T) {
	snd := &slowSender{delay: 30 * time.Millisecond}
	sem, err := NewSemaphore(snd, 2, time.Second)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var wg sync.WaitGroup
	for i := 0; i < 6; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = sem.Send("secret/foo", "msg")
		}()
	}
	wg.Wait()

	if peak := atomic.LoadInt32(&snd.peakCon); peak > 2 {
		t.Errorf("peak concurrency %d exceeded limit of 2", peak)
	}
	if got := atomic.LoadInt32(&snd.called); got != 6 {
		t.Errorf("expected 6 calls, got %d", got)
	}
}

func TestSemaphore_TimeoutReturnsError(t *testing.T) {
	snd := &slowSender{delay: 200 * time.Millisecond}
	sem, err := NewSemaphore(snd, 1, 50*time.Millisecond)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// occupy the single slot
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		_ = sem.Send("secret/slow", "blocking")
	}()
	time.Sleep(10 * time.Millisecond) // let goroutine acquire the slot

	err = sem.Send("secret/fast", "should timeout")
	if err == nil {
		t.Error("expected timeout error but got nil")
	}
	wg.Wait()
}

func TestSemaphore_PropagatesSenderError(t *testing.T) {
	errSender := senderFunc(func(_, _ string) error {
		return errors.New("downstream failure")
	})
	sem, _ := NewSemaphore(errSender, 1, 0)
	if err := sem.Send("secret/x", "msg"); err == nil {
		t.Error("expected error from underlying sender")
	}
}

// TestSemaphore_ZeroTimeoutNoBlock verifies that a zero timeout causes an
// immediate error when all slots are occupied, rather than blocking.
func TestSemaphore_ZeroTimeoutNoBlock(t *testing.T) {
	snd := &slowSender{delay: 200 * time.Millisecond}
	sem, err := NewSemaphore(snd, 1, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// occupy the single slot
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		_ = sem.Send("secret/slow", "blocking")
	}()
	time.Sleep(10 * time.Millisecond) // let goroutine acquire the slot

	start := time.Now()
	err = sem.Send("secret/instant", "no wait")
	elapsed := time.Since(start)

	if err == nil {
		t.Error("expected immediate error with zero timeout but got nil")
	}
	if elapsed > 20*time.Millisecond {
		t.Errorf("zero timeout blocked for %v, expected near-instant return", elapsed)
	}
	wg.Wait()
}

type senderFunc func(path, msg string) error

func (f senderFunc) Send(path, msg string) error { return f(path, msg) }
