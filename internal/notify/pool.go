package notify

import (
	"errors"
	"fmt"
	"sync"
)

// PoolConfig holds configuration for a worker pool sender.
type PoolConfig struct {
	Workers int
	QueueSize int
}

// PoolDefaults returns a PoolConfig with sensible defaults.
func PoolDefaults() PoolConfig {
	return PoolConfig{
		Workers:   4,
		QueueSize: 64,
	}
}

// PooledSender fans out Send calls across a fixed pool of worker goroutines.
type PooledSender struct {
	sender Sender
	work   chan Message
	wg     sync.WaitGroup
	once   sync.Once
	mu     sync.Mutex
	stopped bool
}

// NewPooledSender creates a PooledSender and starts the worker pool.
func NewPooledSender(s Sender, cfg PoolConfig) (*PooledSender, error) {
	if s == nil {
		return nil, errors.New("pool: sender must not be nil")
	}
	if cfg.Workers < 1 {
		return nil, fmt.Errorf("pool: workers must be >= 1, got %d", cfg.Workers)
	}
	if cfg.QueueSize < 1 {
		return nil, fmt.Errorf("pool: queue size must be >= 1, got %d", cfg.QueueSize)
	}

	p := &PooledSender{
		sender: s,
		work:   make(chan Message, cfg.QueueSize),
	}

	for i := 0; i < cfg.Workers; i++ {
		p.wg.Add(1)
		go func() {
			defer p.wg.Done()
			for msg := range p.work {
				_ = s.Send(msg)
			}
		}()
	}

	return p, nil
}

// Send enqueues a message for delivery by the worker pool.
func (p *PooledSender) Send(msg Message) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.stopped {
		return errors.New("pool: sender is stopped")
	}
	select {
	case p.work <- msg:
		return nil
	default:
		return errors.New("pool: queue is full")
	}
}

// Stop drains the queue and waits for all workers to finish.
func (p *PooledSender) Stop() {
	p.once.Do(func() {
		p.mu.Lock()
		p.stopped = true
		p.mu.Unlock()
		close(p.work)
		p.wg.Wait()
	})
}
