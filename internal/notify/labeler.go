package notify

import (
	"fmt"
	"strings"
)

// Labels is a map of key-value metadata attached to a notification.
type Labels map[string]string

// Labeler wraps a Sender and injects a fixed set of labels into every
// message payload before forwarding it downstream.
type Labeler struct {
	sender Sender
	labels Labels
}

// NewLabeler returns a Labeler that enriches outgoing messages with the
// provided labels. An error is returned if sender is nil or labels is empty.
func NewLabeler(sender Sender, labels Labels) (*Labeler, error) {
	if sender == nil {
		return nil, fmt.Errorf("labeler: sender must not be nil")
	}
	if len(labels) == 0 {
		return nil, fmt.Errorf("labeler: labels must not be empty")
	}
	copy := make(Labels, len(labels))
	for k, v := range labels {
		copy[k] = v
	}
	return &Labeler{sender: sender, labels: copy}, nil
}

// Send enriches msg with the configured labels and forwards it to the
// underlying sender. Label keys are injected as "[key=value]" prefixes
// prepended to the message body.
func (l *Labeler) Send(msg string) error {
	enriched := l.enrich(msg)
	return l.sender.Send(enriched)
}

// enrich builds the label prefix and prepends it to msg.
func (l *Labeler) enrich(msg string) string {
	parts := make([]string, 0, len(l.labels))
	for k, v := range l.labels {
		parts = append(parts, fmt.Sprintf("%s=%s", k, v))
	}
	prefix := "[" + strings.Join(parts, " ") + "]"
	return prefix + " " + msg
}

// Labels returns a copy of the labels configured on this Labeler.
func (l *Labeler) Labels() Labels {
	copy := make(Labels, len(l.labels))
	for k, v := range l.labels {
		copy[k] = v
	}
	return copy
}
