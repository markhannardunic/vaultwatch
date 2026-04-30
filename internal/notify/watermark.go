package notify

import (
	"fmt"
	"sync"
	"time"
)

// WatermarkSender suppresses alerts whose severity has not changed since the
// last notification. A new alert is only forwarded when the level transitions
// (e.g. healthy → warning, warning → critical) or when the reset window
// expires, ensuring operators are not flooded with repeated same-level alerts.
type WatermarkSender struct {
	mu      sync.Mutex
	sender  Sender
	window  time.Duration
	marks   map[string]watermark
	nowFunc func() time.Time
}

type watermark struct {
	level     string
	recordedAt time.Time
}

// NewWatermarkSender creates a WatermarkSender wrapping sender. window is how
// long a level is remembered before the record expires and the next alert of
// the same level is forwarded again.
func NewWatermarkSender(sender Sender, window time.Duration) (*WatermarkSender, error) {
	if sender == nil {
		return nil, fmt.Errorf("watermark: sender must not be nil")
	}
	if window <= 0 {
		return nil, fmt.Errorf("watermark: window must be positive")
	}
	return &WatermarkSender{
		sender:  sender,
		window:  window,
		marks:   make(map[string]watermark),
		nowFunc: time.Now,
	}, nil
}

// Send forwards msg only when the level for msg.Path has changed or the
// watermark window has expired.
func (w *WatermarkSender) Send(msg Message) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	now := w.nowFunc()
	key := msg.Path
	if prev, ok := w.marks[key]; ok {
		if prev.level == msg.Level && now.Before(prev.recordedAt.Add(w.window)) {
			return nil // same level within window — suppress
		}
	}
	w.marks[key] = watermark{level: msg.Level, recordedAt: now}
	return w.sender.Send(msg)
}

// Purge removes expired watermarks to prevent unbounded memory growth.
func (w *WatermarkSender) Purge() {
	w.mu.Lock()
	defer w.mu.Unlock()
	now := w.nowFunc()
	for k, m := range w.marks {
		if now.After(m.recordedAt.Add(w.window)) {
			delete(w.marks, k)
		}
	}
}
