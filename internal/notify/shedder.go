package notify

import (
	"errors"
	"sync/atomic"
	"time"
)

// LoadShedder drops outbound notifications when the system is under load,
// protecting downstream senders from being overwhelmed during traffic spikes.
// It tracks an exponentially weighted moving average (EWMA) of send latency
// and begins shedding when the average exceeds a configurable threshold.
type LoadShedder struct {
	sender    Sender
	threshold time.Duration
	dropRate  float64 // fraction of messages to drop when overloaded [0,1]

	// ewma state — stored as nanoseconds in an atomic int64
	ewma    atomic.Int64
	alpha   float64 // smoothing factor
	drops   atomic.Int64
	sent    atomic.Int64
}

// ErrLoadShed is returned when a message is intentionally dropped due to load.
var ErrLoadShed = errors.New("notify: message shed due to high load")

// NewLoadShedder wraps sender and sheds messages when the EWMA of send
// latency exceeds threshold. dropRate controls what fraction of messages are
// dropped while the system is overloaded (0.0 = drop none, 1.0 = drop all).
// alpha is the EWMA smoothing factor; a value of 0.2 is a sensible default.
func NewLoadShedder(sender Sender, threshold time.Duration, dropRate, alpha float64) (*LoadShedder, error) {
	if sender == nil {
		return nil, errors.New("notify: LoadShedder requires a non-nil sender")
	}
	if threshold <= 0 {
		return nil, errors.New("notify: LoadShedder threshold must be positive")
	}
	if dropRate < 0 || dropRate > 1 {
		return nil, errors.New("notify: LoadShedder dropRate must be in [0, 1]")
	}
	if alpha <= 0 || alpha > 1 {
		return nil, errors.New("notify: LoadShedder alpha must be in (0, 1]")
	}
	return &LoadShedder{
		sender:    sender,
		threshold: threshold,
		dropRate:  dropRate,
		alpha:     alpha,
	}, nil
}

// Send delivers msg to the underlying sender, updating the latency EWMA after
// each call. If the EWMA exceeds the configured threshold and the message falls
// within the shed fraction, ErrLoadShed is returned immediately without
// forwarding to the downstream sender.
func (l *LoadShedder) Send(msg string) error {
	if l.overloaded() && l.shouldShed() {
		l.drops.Add(1)
		return ErrLoadShed
	}

	start := time.Now()
	err := l.sender.Send(msg)
	elapsed := time.Since(start)

	l.updateEWMA(elapsed)
	l.sent.Add(1)
	return err
}

// Stats returns the current EWMA latency, total sent count, and total dropped
// count. Useful for exposing metrics to monitoring systems.
func (l *LoadShedder) Stats() (ewma time.Duration, sent, dropped int64) {
	return time.Duration(l.ewma.Load()), l.sent.Load(), l.drops.Load()
}

// overloaded reports whether the current EWMA latency exceeds the threshold.
func (l *LoadShedder) overloaded() bool {
	return time.Duration(l.ewma.Load()) > l.threshold
}

// shouldShed uses a simple counter-based approach to honour the configured
// drop rate without importing math/rand: every 1/dropRate-th message is kept.
func (l *LoadShedder) shouldShed() bool {
	if l.dropRate == 0 {
		return false
	}
	if l.dropRate == 1 {
		return true
	}
	// Use the running sent+drops counter as a deterministic selector.
	total := l.sent.Load() + l.drops.Load()
	keepEvery := int64(1.0 / (1.0 - l.dropRate))
	if keepEvery < 1 {
		keepEvery = 1
	}
	return total%keepEvery != 0
}

// updateEWMA updates the exponentially weighted moving average of latency.
func (l *LoadShedder) updateEWMA(sample time.Duration) {
	prev := l.ewma.Load()
	next := int64(l.alpha*float64(sample) + (1-l.alpha)*float64(prev))
	l.ewma.Store(next)
}
