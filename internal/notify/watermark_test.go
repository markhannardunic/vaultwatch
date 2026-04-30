package notify

import (
	"errors"
	"testing"
	"time"
)

func TestNewWatermarkSender_NilSender(t *testing.T) {
	_, err := NewWatermarkSender(nil, time.Minute)
	if err == nil {
		t.Fatal("expected error for nil sender")
	}
}

func TestNewWatermarkSender_ZeroWindow(t *testing.T) {
	_, err := NewWatermarkSender(&captureSender{}, 0)
	if err == nil {
		t.Fatal("expected error for zero window")
	}
}

func TestWatermarkSender_FirstSend_Forwarded(t *testing.T) {
	cap := &captureSender{}
	w, _ := NewWatermarkSender(cap, time.Minute)

	if err := w.Send(Message{Path: "secret/a", Level: "warning", Body: "expiring"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cap.messages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(cap.messages))
	}
}

func TestWatermarkSender_SameLevel_WithinWindow_Suppressed(t *testing.T) {
	cap := &captureSender{}
	w, _ := NewWatermarkSender(cap, time.Minute)
	now := time.Now()
	w.nowFunc = func() time.Time { return now }

	msg := Message{Path: "secret/a", Level: "warning", Body: "expiring"}
	_ = w.Send(msg)
	_ = w.Send(msg) // same level, same window

	if len(cap.messages) != 1 {
		t.Fatalf("expected 1 forwarded message, got %d", len(cap.messages))
	}
}

func TestWatermarkSender_LevelTransition_Forwarded(t *testing.T) {
	cap := &captureSender{}
	w, _ := NewWatermarkSender(cap, time.Minute)
	now := time.Now()
	w.nowFunc = func() time.Time { return now }

	_ = w.Send(Message{Path: "secret/a", Level: "warning", Body: "warn"})
	_ = w.Send(Message{Path: "secret/a", Level: "critical", Body: "crit"}) // level changed

	if len(cap.messages) != 2 {
		t.Fatalf("expected 2 forwarded messages, got %d", len(cap.messages))
	}
}

func TestWatermarkSender_SameLevel_AfterWindowExpiry_Forwarded(t *testing.T) {
	cap := &captureSender{}
	w, _ := NewWatermarkSender(cap, time.Minute)
	now := time.Now()
	w.nowFunc = func() time.Time { return now }
	_ = w.Send(Message{Path: "secret/a", Level: "warning", Body: "first"})

	w.nowFunc = func() time.Time { return now.Add(2 * time.Minute) } // past window
	_ = w.Send(Message{Path: "secret/a", Level: "warning", Body: "second"})

	if len(cap.messages) != 2 {
		t.Fatalf("expected 2 forwarded messages, got %d", len(cap.messages))
	}
}

func TestWatermarkSender_PropagatesError(t *testing.T) {
	sentinel := errors.New("send failed")
	w, _ := NewWatermarkSender(&errorSender{err: sentinel}, time.Minute)
	err := w.Send(Message{Path: "secret/a", Level: "warning", Body: "x"})
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel error, got %v", err)
	}
}

func TestWatermarkSender_Purge_RemovesExpired(t *testing.T) {
	cap := &captureSender{}
	w, _ := NewWatermarkSender(cap, time.Minute)
	now := time.Now()
	w.nowFunc = func() time.Time { return now }
	_ = w.Send(Message{Path: "secret/a", Level: "warning", Body: "x"})

	w.nowFunc = func() time.Time { return now.Add(2 * time.Minute) }
	w.Purge()

	w.mu.Lock()
	n := len(w.marks)
	w.mu.Unlock()
	if n != 0 {
		t.Fatalf("expected marks to be empty after purge, got %d", n)
	}
}
