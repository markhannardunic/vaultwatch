package notify

import (
	"errors"
	"time"
)

// QueuedRouter wraps a Router so that Dispatch calls are handled
// asynchronously via a QueuedSender per registered level.
type QueuedRouter struct {
	router  *Router
	queued  *QueuedSender
	timeout time.Duration
}

// routerSenderAdapter adapts a *Router to the Sender interface so that
// QueuedSender can drive dispatch calls.
type routerSenderAdapter struct {
	router *Router
}

func (a *routerSenderAdapter) Send(msg Message) error {
	return a.router.Dispatch(msg)
}

// NewQueuedRouter creates a QueuedRouter that dispatches messages through
// the provided Router using an async queue.
func NewQueuedRouter(router *Router, cfg QueueConfig, stopTimeout time.Duration) (*QueuedRouter, error) {
	if router == nil {
		return nil, errors.New("queued_router: router must not be nil")
	}
	adapter := &routerSenderAdapter{router: router}
	qs, err := NewQueuedSender(adapter, cfg)
	if err != nil {
		return nil, err
	}
	qs.Start()
	return &QueuedRouter{
		router:  router,
		queued:  qs,
		timeout: stopTimeout,
	}, nil
}

// Dispatch enqueues the message for async delivery.
func (qr *QueuedRouter) Dispatch(msg Message) error {
	return qr.queued.Send(msg)
}

// Stop drains the queue and shuts down workers.
func (qr *QueuedRouter) Stop() {
	qr.queued.Stop(qr.timeout)
}
