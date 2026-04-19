package notify

import (
	"fmt"
	"io"
)

// Sender is anything that can send a message string.
type Sender interface {
	Send(message string) error
}

// Router dispatches alerts to one or more Senders based on level.
type Router struct {
	routes map[string][]Sender
}

// NewRouter creates an empty Router.
func NewRouter() *Router {
	return &Router{routes: make(map[string][]Sender)}
}

// Register adds a Sender for the given alert level (e.g. "warning", "critical").
// Use "*" to receive all levels.
func (r *Router) Register(level string, s Sender) {
	r.routes[level] = append(r.routes[level], s)
}

// Dispatch sends the message to all Senders registered for the given level
// and to any wildcard ("*") Senders. Errors are collected and returned together.
func (r *Router) Dispatch(level, message string) error {
	targets := append(r.routes[level], r.routes["*"]...)
	var errs []error
	for _, s := range targets {
		if err := s.Send(message); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("dispatch errors: %v", errs)
	}
	return nil
}

// WriterSender wraps an io.Writer as a Sender.
type WriterSender struct{ w io.Writer }

// NewWriterSender creates a Sender that writes to w.
func NewWriterSender(w io.Writer) *WriterSender { return &WriterSender{w: w} }

func (ws *WriterSender) Send(message string) error {
	_, err := fmt.Fprintln(ws.w, message)
	return err
}
