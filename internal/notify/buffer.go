package notify

import (
	"fmt"
	"sync"
	"time"
)

// BufferConfig controls how the BufferedSender behaves.
type BufferConfig struct {
	MaxSize     int
	FlushEvery  time.Duration
	DropOnFull  bool
}

// BufferDefaults returns sensible defaults for BufferConfig.
func BufferDefaults() BufferConfig {
	return BufferConfig{
		MaxSize:    100,
		FlushEvery: 30 * time.Second,
		DropOnFull: false,
	}
}

// BufferedSender queues messages and forwards them to an underlying Sender
// either when the buffer is full or a flush interval elapses.
type BufferedSender struct {
	mu      sync.Mutex
	queue   []Message
	sender  Sender
	cfg     BufferConfig
	stop    chan struct{}
	wg      sync.WaitGroup
}

// NewBufferedSender creates a BufferedSender and starts its flush loop.
func NewBufferedSender(s Sender, cfg BufferConfig) (*BufferedSender, error) {
	if s == nil {
		return nil, fmt.Errorf("buffer: sender must not be nil")
	}
	if cfg.MaxSize <= 0 {
		return nil, fmt.Errorf("buffer: MaxSize must be > 0")
	}
	if cfg.FlushEvery <= 0 {
		return nil, fmt.Errorf("buffer: FlushEvery must be > 0")
	}
	b := &BufferedSender{
		sender: s,
		cfg:    cfg,
		stop:   make(chan struct{}),
	}
	b.wg.Add(1)
	go b.loop()
	return b, nil
}

// Send enqueues a message, flushing immediately if the buffer is full.
func (b *BufferedSender) Send(msg Message) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if len(b.queue) >= b.cfg.MaxSize {
		if b.cfg.DropOnFull {
			return nil
		}
		b.flushLocked()
	}
	b.queue = append(b.queue, msg)
	return nil
}

// Flush drains the queue immediately.
func (b *BufferedSender) Flush() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.flushLocked()
}

// Stop halts the background flush loop.
func (b *BufferedSender) Stop() {
	close(b.stop)
	b.wg.Wait()
	b.Flush()
}

func (b *BufferedSender) loop() {
	defer b.wg.Done()
	ticker := time.NewTicker(b.cfg.FlushEvery)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			b.Flush()
		case <-b.stop:
			return
		}
	}
}

func (b *BufferedSender) flushLocked() {
	for _, msg := range b.queue {
		_ = b.sender.Send(msg)
	}
	b.queue = b.queue[:0]
}
