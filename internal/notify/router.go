package notify

import (
	"fmt"
	"io"
	"strings"
)

// Sender is the interface for notification backends.
type Sender interface {
	Send(level, message string) error
}

// Router dispatches alerts to one or more Senders based on alert level.
type Router struct {
	routes map[string][]Sender
}

// NewRouter creates an empty Router.
func NewRouter() *Router {
	return &Router{routes: make(map[string][]Sender)}
}

// Register adds a Sender for the given level. Use "*" to receive all levels.
func (r *Router) Register(level string, s Sender) {
	level = strings.ToLower(level)
	r.routes[level] = append(r.routes[level], s)
}

// Dispatch sends the message to all Senders registered for the level
// and to any wildcard ("*") Senders. Errors from all senders are collected.
func (r *Router) Dispatch(level, message string) error {
	level = strings.ToLower(level)
	var errs []string
	for _, s := range r.routes[level] {
		if err := s.Send(level, message); err != nil {
			errs = append(errs, err.Error())
		}
	}
	if level != "*" {
		for _, s := range r.routes["*"] {
			if err := s.Send(level, message); err != nil {
				errs = append(errs, err.Error())
			}
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("router: %d sender(s) failed: %s", len(errs), strings.Join(errs, "; "))
	}
	return nil
}

// WriterSender is a Sender that writes formatted alerts to an io.Writer.
type WriterSender struct {
	w io.Writer
}

// NewWriterSender returns a Sender that writes to w.
func NewWriterSender(w io.Writer) *WriterSender {
	return &WriterSender{w: w}
}

// Send writes "[level] message" to the underlying writer.
func (ws *WriterSender) Send(level, message string) error {
	_, err := fmt.Fprintf(ws.w, "[%s] %s\n", strings.ToUpper(level), message)
	return err
}
