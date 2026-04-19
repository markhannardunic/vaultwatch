package alert

import (
	"fmt"
	"io"
	"os"
	"time"
)

// Severity represents the urgency level of an alert.
type Severity string

const (
	SeverityCritical Severity = "CRITICAL"
	SeverityWarning  Severity = "WARNING"
	SeverityInfo     Severity = "INFO"
)

// Alert holds information about an expiring secret.
type Alert struct {
	Path      string
	ExpiresAt time.Time
	TTL       time.Duration
	Severity  Severity
}

// Notifier sends alerts to one or more outputs.
type Notifier struct {
	writer    io.Writer
	thresholds map[Severity]time.Duration
}

// NewNotifier creates a Notifier writing to the given writer.
// If writer is nil, os.Stdout is used.
func NewNotifier(writer io.Writer) *Notifier {
	if writer == nil {
		writer = os.Stdout
	}
	return &Notifier{
		writer: writer,
		thresholds: map[Severity]time.Duration{
			SeverityCritical: 24 * time.Hour,
			SeverityWarning:  72 * time.Hour,
			SeverityInfo:     168 * time.Hour,
		},
	}
}

// Classify returns the severity for a given TTL.
func (n *Notifier) Classify(ttl time.Duration) Severity {
	switch {
	case ttl <= n.thresholds[SeverityCritical]:
		return SeverityCritical
	case ttl <= n.thresholds[SeverityWarning]:
		return SeverityWarning
	default:
		return SeverityInfo
	}
}

// Send writes a formatted alert line to the notifier's writer.
func (n *Notifier) Send(a Alert) error {
	_, err := fmt.Fprintf(
		n.writer,
		"[%s] secret=%s expires=%s ttl=%s\n",
		a.Severity,
		a.Path,
		a.ExpiresAt.UTC().Format(time.RFC3339),
		a.TTL.Round(time.Second),
	)
	return err
}
