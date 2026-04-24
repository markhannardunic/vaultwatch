package notify

import (
	"errors"
	"sync"
	"testing"
	"time"
)

type stubSender struct {
	err error
}

func (s *stubSender) Send(path, message string) error { return s.err }

func TestNewObservableSender_NilSender(t *testing.T) {
	_, err := NewObservableSender(nil)
	if err == nil {
		t.Fatal("expected error for nil sender")
	}
}

func TestNewObservableSender_ValidSender(t *testing.T) {
	os, err := NewObservableSender(&stubSender{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if os == nil {
		t.Fatal("expected non-nil ObservableSender")
	}
}

func TestObservableSender_EmitsSentEvent(t *testing.T) {
	os, _ := NewObservableSender(&stubSender{})

	var got Event
	var wg sync.WaitGroup
	wg.Add(1)
	os.Register(ObserverFunc(func(e Event) {
		got = e
		wg.Done()
	}))

	_ = os.Send("secret/foo", "expiring soon")
	wg.Wait()

	if got.Type != EventSent {
		t.Errorf("expected EventSent, got %s", got.Type)
	}
	if got.Path != "secret/foo" {
		t.Errorf("expected path secret/foo, got %s", got.Path)
	}
	if got.Err != nil {
		t.Errorf("expected nil error, got %v", got.Err)
	}
	if got.Timestamp.IsZero() {
		t.Error("expected non-zero timestamp")
	}
}

func TestObservableSender_EmitsFailedEvent(t *testing.T) {
	sentErr := errors.New("connection refused")
	os, _ := NewObservableSender(&stubSender{err: sentErr})

	var got Event
	var wg sync.WaitGroup
	wg.Add(1)
	os.Register(ObserverFunc(func(e Event) {
		got = e
		wg.Done()
	}))

	_ = os.Send("secret/bar", "critical")
	wg.Wait()

	if got.Type != EventFailed {
		t.Errorf("expected EventFailed, got %s", got.Type)
	}
	if !errors.Is(got.Err, sentErr) {
		t.Errorf("expected wrapped error, got %v", got.Err)
	}
}

func TestObservableSender_MultipleObservers(t *testing.T) {
	os, _ := NewObservableSender(&stubSender{})

	var mu sync.Mutex
	var count int
	var wg sync.WaitGroup
	wg.Add(2)

	for i := 0; i < 2; i++ {
		os.Register(ObserverFunc(func(e Event) {
			mu.Lock()
			count++
			mu.Unlock()
			wg.Done()
		}))
	}

	_ = os.Send("secret/baz", "warning")

	done := make(chan struct{})
	go func() { wg.Wait(); close(done) }()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for observers")
	}

	mu.Lock()
	defer mu.Unlock()
	if count != 2 {
		t.Errorf("expected 2 observer calls, got %d", count)
	}
}
