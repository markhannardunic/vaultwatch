package audit

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"
)

// Entry represents a single audit log record.
type Entry struct {
	Timestamp time.Time `json:"timestamp"`
	Path      string    `json:"path"`
	Level     string    `json:"level"`
	Message   string    `json:"message"`
	TTL       int64     `json:"ttl_seconds"`
}

// Logger writes audit entries to an output.
type Logger struct {
	out    io.Writer
	format string
}

// NewLogger creates a Logger. format is "json" or "text".
func NewLogger(out io.Writer, format string) *Logger {
	if out == nil {
		out = os.Stdout
	}
	if format == "" {
		format = "text"
	}
	return &Logger{out: out, format: format}
}

// Write records an audit entry.
func (l *Logger) Write(e Entry) error {
	if e.Timestamp.IsZero() {
		e.Timestamp = time.Now().UTC()
	}
	switch l.format {
	case "json":
		return json.NewEncoder(l.out).Encode(e)
	default:
		_, err := fmt.Fprintf(l.out, "[%s] %s %-8s ttl=%ds %s\n",
			e.Timestamp.Format(time.RFC3339),
			e.Path, e.Level, e.TTL, e.Message)
		return err
	}
}
