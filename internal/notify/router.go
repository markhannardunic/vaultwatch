package notify

import (
	"fmt"
	"io"
	"strings"
)

// Sender is the interface satisfied by all notification back-ends.
type Sender interface {
	Send(level, message, path string) error
}

// Router dispatches alerts to one or more Senders based on alert level.
type Router struct {
	routes map[string][]Sender
}

// NewRouter returns an empty Router.
func NewRouter() *Router {
	return &Router{routes: make(map[string][]Sender)}
}

// Register adds a Sender for the given level. Use "*" to receive all levels.
func (r *Router) Register(level string, s Sender) {
	level = strings.ToLower(level)
	r.routes[level] = append(r.routes[level], s)
}

// Dispatch sends the alert to every Sender registered for the level or "*".
// All errors are collected and returned as a single combined error.
func (r *Router) Dispatch(level, message, path string) error {
	level = strings.ToLower(level)
	var senders []Sender
	if ss, ok := r.routes[level]; ok {
		senders = append(senders, ss...)
	}
	if wildcards, ok := r.routes["*"]; ok {
		senders = append(senders, wildcards...)
	}
	var errs []string
	for _, s := range senders {
		if err := s.Send(level, message, path); err != nil {
			errs = append(errs, err.Error())
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("router dispatch errors: %s", strings.Join(errs, "; "))
	}
	return nil
}

// writerSender is a simple Sender that writes formatted alerts to an io.Writer.
type writerSender struct{ w io.Writer }

// NewWriterSender returns a Sender that writes alerts as text to w.
func NewWriterSender(w io.Writer) Sender {
	return &writerSender{w: w}
}

func (ws *writerSender) Send(level, message, path string) error {
	_, err := fmt.Fprintf(ws.w, "[%s] %s (%s)\n", strings.ToUpper(level), message, path)
	return err
}
