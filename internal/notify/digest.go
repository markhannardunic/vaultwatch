package notify

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

// DigestSender accumulates alerts over a time window and sends them as a
// single batched digest message via the wrapped Sender.
type DigestSender struct {
	mu       sync.Mutex
	sender   Sender
	window   time.Duration
	entries  []string
	timer    *time.Timer
	header   string
}

// NewDigestSender creates a DigestSender that flushes accumulated alerts
// through sender after window duration. header is prepended to the digest.
func NewDigestSender(sender Sender, window time.Duration, header string) (*DigestSender, error) {
	if sender == nil {
		return nil, fmt.Errorf("digest: sender must not be nil")
	}
	if window <= 0 {
		return nil, fmt.Errorf("digest: window must be positive, got %s", window)
	}
	h := header
	if h == "" {
		h = "VaultWatch Digest"
	}
	return &DigestSender{
		sender: sender,
		window: window,
		header: h,
	}, nil
}

// Send buffers the message and schedules a flush if one is not already pending.
func (d *DigestSender) Send(path, message string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.entries = append(d.entries, fmt.Sprintf("[%s] %s", path, message))

	if d.timer == nil {
		d.timer = time.AfterFunc(d.window, d.flush)
	}
	return nil
}

// Flush forces an immediate send of all buffered entries, cancelling any
// pending timer. It is safe to call concurrently.
func (d *DigestSender) Flush() error {
	d.mu.Lock()
	if d.timer != nil {
		d.timer.Stop()
		d.timer = nil
	}
	d.mu.Unlock()
	d.flush()
	return nil
}

func (d *DigestSender) flush() {
	d.mu.Lock()
	if len(d.entries) == 0 {
		d.timer = nil
		d.mu.Unlock()
		return
	}
	lines := make([]string, len(d.entries))
	copy(lines, d.entries)
	d.entries = d.entries[:0]
	d.timer = nil
	d.mu.Unlock()

	body := fmt.Sprintf("%s\n\n%s", d.header, strings.Join(lines, "\n"))
	_ = d.sender.Send("digest", body)
}
