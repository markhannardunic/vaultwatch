package notify

import (
	"fmt"
	"time"
)

// BufferedRouter wraps a Router so that outgoing dispatches are queued in a
// BufferedSender before being forwarded to each registered sender.
type BufferedRouter struct {
	buffer *BufferedSender
	router *Router
}

// NewBufferedRouter creates a BufferedRouter using the provided Router and
// buffer settings. The caller must call Stop() when done.
func NewBufferedRouter(r *Router, maxSize int, flushEvery time.Duration, dropOnFull bool) (*BufferedRouter, error) {
	if r == nil {
		return nil, fmt.Errorf("buffered_router: router must not be nil")
	}
	// The buffer forwards each message through the underlying router.
	proxy := &routerSender{router: r}
	cfg := BufferConfig{
		MaxSize:    maxSize,
		FlushEvery: flushEvery,
		DropOnFull: dropOnFull,
	}
	buf, err := NewBufferedSender(proxy, cfg)
	if err != nil {
		return nil, fmt.Errorf("buffered_router: %w", err)
	}
	return &BufferedRouter{buffer: buf, router: r}, nil
}

// Dispatch enqueues a message into the buffer.
func (br *BufferedRouter) Dispatch(msg Message) error {
	return br.buffer.Send(msg)
}

// Flush drains the buffer immediately.
func (br *BufferedRouter) Flush() {
	br.buffer.Flush()
}

// Stop halts the background flush loop and performs a final flush.
func (br *BufferedRouter) Stop() {
	br.buffer.Stop()
}

// routerSender adapts a *Router to the Sender interface.
type routerSender struct {
	router *Router
}

func (rs *routerSender) Send(msg Message) error {
	return rs.router.Dispatch(msg)
}
