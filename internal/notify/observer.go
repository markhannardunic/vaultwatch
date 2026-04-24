package notify

import (
	"fmt"
	"sync"
	"time"
)

// EventType represents the kind of notification lifecycle event.
type EventType string

const (
	EventSent    EventType = "sent"
	EventFailed  EventType = "failed"
	EventSkipped EventType = "skipped"
)

// Event describes a single notification dispatch outcome.
type Event struct {
	Path      string
	Level     string
	Type      EventType
	Err       error
	Timestamp time.Time
}

// Observer receives notification lifecycle events.
type Observer interface {
	OnEvent(e Event)
}

// ObserverFunc is a functional adapter for Observer.
type ObserverFunc func(e Event)

func (f ObserverFunc) OnEvent(e Event) { f(e) }

// ObservableSender wraps a Sender and emits events to registered observers.
type ObservableSender struct {
	mu        sync.RWMutex
	sender    Sender
	observers []Observer
}

// NewObservableSender creates an ObservableSender wrapping the given sender.
func NewObservableSender(s Sender) (*ObservableSender, error) {
	if s == nil {
		return nil, fmt.Errorf("notify: ObservableSender requires a non-nil sender")
	}
	return &ObservableSender{sender: s}, nil
}

// Register adds an observer to receive future events.
func (o *ObservableSender) Register(obs Observer) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.observers = append(o.observers, obs)
}

// Send dispatches the message and notifies observers of the outcome.
func (o *ObservableSender) Send(path, message string) error {
	err := o.sender.Send(path, message)

	ev := Event{
		Path:      path,
		Timestamp: time.Now(),
	}

	if err != nil {
		ev.Type = EventFailed
		ev.Err = err
	} else {
		ev.Type = EventSent
	}

	o.mu.RLock()
	obs := make([]Observer, len(o.observers))
	copy(obs, o.observers)
	o.mu.RUnlock()

	for _, ob := range obs {
		ob.OnEvent(ev)
	}

	return err
}
