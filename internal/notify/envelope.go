package notify

import (
	"fmt"
	"time"
)

// Envelope wraps a notification message with metadata used for routing,
// deduplication, and audit purposes.
type Envelope struct {
	// Path is the secret path this notification relates to.
	Path string

	// Level is the alert severity: "warning", "critical", or "healthy".
	Level string

	// Message is the human-readable alert body.
	Message string

	// CreatedAt is when the envelope was constructed.
	CreatedAt time.Time

	// TTL is how long the envelope is considered valid for delivery.
	// Zero means no expiry.
	TTL time.Duration

	// Attempt tracks how many send attempts have been made.
	Attempt int

	// Meta holds arbitrary key/value annotations.
	Meta map[string]string
}

// NewEnvelope constructs an Envelope with CreatedAt set to now.
func NewEnvelope(path, level, message string) *Envelope {
	return &Envelope{
		Path:      path,
		Level:     level,
		Message:   message,
		CreatedAt: time.Now().UTC(),
		Meta:      make(map[string]string),
	}
}

// Expired reports whether the envelope's TTL has elapsed.
// If TTL is zero the envelope never expires.
func (e *Envelope) Expired() bool {
	if e.TTL == 0 {
		return false
	}
	return time.Now().UTC().After(e.CreatedAt.Add(e.TTL))
}

// Key returns a stable string identifier suitable for deduplication
// and throttle keying.
func (e *Envelope) Key() string {
	return fmt.Sprintf("%s::%s", e.Level, e.Path)
}

// WithMeta returns the envelope after setting a metadata key/value pair.
func (e *Envelope) WithMeta(key, value string) *Envelope {
	e.Meta[key] = value
	return e
}
