package notify

import (
	"errors"
	"sync"
	"time"
)

// QueueConfig holds configuration for the persistent send queue.
type QueueConfig struct {
	MaxSize     int
	Workers     int
	DrainOnStop bool
}

// QueueDefaults returns a QueueConfig with sensible defaults.
func QueueDefaults() QueueConfig {
	return QueueConfig{
		MaxSize:     256,
		Workers:     2,
		DrainOnStop: true,
	}
}

// QueuedSender wraps a Sender with an async work queue.
type QueuedSender struct {
	sender  Sender
	cfg     QueueConfig
	queue   chan Message
	wg      sync.WaitGroup
	stopCh  chan struct{}
	once    sync.Once
	started bool
}

// NewQueuedSender creates a QueuedSender wrapping the given Sender.
func NewQueuedSender(sender Sender, cfg QueueConfig) (*QueuedSender, error) {
	if sender == nil {
		return nil, errors.New("queue: sender must not be nil")
	}
	if cfg.MaxSize <= 0 {
		return nil, errors.New("queue: max size must be positive")
	}
	if cfg.Workers <= 0 {
		cfg.Workers = 1
	}
	return &QueuedSender{
		sender: sender,
		cfg:    cfg,
		queue:  make(chan Message, cfg.MaxSize),
		stopCh: make(chan struct{}),
	}, nil
}

// Start launches the background worker goroutines.
func (q *QueuedSender) Start() {
	q.once.Do(func() {
		q.started = true
		for i := 0; i < q.cfg.Workers; i++ {
			q.wg.Add(1)
			go q.work()
		}
	})
}

// Send enqueues a message for async delivery.
func (q *QueuedSender) Send(msg Message) error {
	select {
	case q.queue <- msg:
		return nil
	default:
		return errors.New("queue: send queue is full")
	}
}

// Stop signals workers to stop. If DrainOnStop is true, it waits for the
// queue to be drained before returning.
func (q *QueuedSender) Stop(timeout time.Duration) {
	close(q.stopCh)
	done := make(chan struct{})
	go func() {
		q.wg.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(timeout):
	}
}

func (q *QueuedSender) work() {
	defer q.wg.Done()
	for {
		select {
		case msg := <-q.queue:
			_ = q.sender.Send(msg)
		case <-q.stopCh:
			if q.cfg.DrainOnStop {
				for {
					select {
					case msg := <-q.queue:
						_ = q.sender.Send(msg)
					default:
						return
					}
				}
			}
			return
		}
	}
}
