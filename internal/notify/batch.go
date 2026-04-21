package notify

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

// BatchConfig controls how alerts are grouped before dispatch.
type BatchConfig struct {
	MaxSize  int
	MaxWait  time.Duration
}

// BatchSender buffers alerts and flushes them as a single combined message.
type BatchSender struct {
	mu      sync.Mutex
	cfg     BatchConfig
	sender  Sender
	buf     []string
	timer   *time.Timer
}

// NewBatchSender creates a BatchSender that wraps the given Sender.
// Alerts are flushed when MaxSize is reached or MaxWait elapses.
func NewBatchSender(cfg BatchConfig, s Sender) (*BatchSender, error) {
	if s == nil {
		return nil, fmt.Errorf("batch: sender must not be nil")
	}
	if cfg.MaxSize <= 0 {
		cfg.MaxSize = 10
	}
	if cfg.MaxWait <= 0 {
		cfg.MaxWait = 30 * time.Second
	}
	bs := &BatchSender{cfg: cfg, sender: s}
	bs.resetTimer()
	return bs, nil
}

// Send adds a message to the batch buffer.
func (bs *BatchSender) Send(path, message string) error {
	bs.mu.Lock()
	defer bs.mu.Unlock()
	bs.buf = append(bs.buf, fmt.Sprintf("[%s] %s", path, message))
	if len(bs.buf) >= bs.cfg.MaxSize {
		return bs.flushLocked(path)
	}
	return nil
}

// Flush forces immediate dispatch of any buffered alerts.
func (bs *BatchSender) Flush(path string) error {
	bs.mu.Lock()
	defer bs.mu.Unlock()
	return bs.flushLocked(path)
}

func (bs *BatchSender) flushLocked(path string) error {
	if len(bs.buf) == 0 {
		return nil
	}
	combined := strings.Join(bs.buf, "\n")
	bs.buf = bs.buf[:0]
	bs.resetTimer()
	return bs.sender.Send(path, combined)
}

func (bs *BatchSender) resetTimer() {
	if bs.timer != nil {
		bs.timer.Stop()
	}
	bs.timer = time.AfterFunc(bs.cfg.MaxWait, func() {
		bs.mu.Lock()
		defer bs.mu.Unlock()
		_ = bs.flushLocked("batch/timer")
	})
}
