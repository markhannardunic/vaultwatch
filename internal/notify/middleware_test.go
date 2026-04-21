package notify

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"strings"
	"testing"
	"time"
)

// stubMiddlewareSender is a minimal MiddlewareSender for tests.
type stubMiddlewareSender struct {
	calls []string
	err   error
}

func (s *stubMiddlewareSender) Send(level, path, message string) error {
	s.calls = append(s.calls, fmt.Sprintf("%s:%s", level, path))
	return s.err
}

func TestNewLoggingMiddleware_NilSender(t *testing.T) {
	_, err := NewLoggingMiddleware(nil, log.New(bytes.NewBuffer(nil), "", 0))
	if err == nil {
		t.Fatal("expected error for nil sender")
	}
}

func TestNewLoggingMiddleware_NilLogger(t *testing.T) {
	_, err := NewLoggingMiddleware(&stubMiddlewareSender{}, nil)
	if err == nil {
		t.Fatal("expected error for nil logger")
	}
}

func TestLoggingMiddleware_Send_Success(t *testing.T) {
	var buf bytes.Buffer
	logger := log.New(&buf, "", 0)
	stub := &stubMiddlewareSender{}

	m, err := NewLoggingMiddleware(stub, logger)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := m.Send("warning", "secret/foo", "expiring soon"); err != nil {
		t.Fatalf("unexpected send error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "dispatching") || !strings.Contains(out, "succeeded") {
		t.Errorf("expected log to contain dispatching and succeeded, got: %s", out)
	}
}

func TestLoggingMiddleware_Send_Error(t *testing.T) {
	var buf bytes.Buffer
	logger := log.New(&buf, "", 0)
	stub := &stubMiddlewareSender{err: errors.New("send failed")}

	m, _ := NewLoggingMiddleware(stub, logger)
	err := m.Send("critical", "secret/bar", "expired")
	if err == nil {
		t.Fatal("expected error to propagate")
	}
	if !strings.Contains(buf.String(), "failed") {
		t.Errorf("expected failure log, got: %s", buf.String())
	}
}

func TestNewTimingMiddleware_NilSender(t *testing.T) {
	_, err := NewTimingMiddleware(nil, func(_ string, _ time.Duration, _ error) {})
	if err == nil {
		t.Fatal("expected error for nil sender")
	}
}

func TestNewTimingMiddleware_NilRecorder(t *testing.T) {
	_, err := NewTimingMiddleware(&stubMiddlewareSender{}, nil)
	if err == nil {
		t.Fatal("expected error for nil recorder")
	}
}

func TestTimingMiddleware_RecordsLatency(t *testing.T) {
	stub := &stubMiddlewareSender{}
	var recorded time.Duration
	var recordedLevel string

	m, err := NewTimingMiddleware(stub, func(level string, d time.Duration, _ error) {
		recordedLevel = level
		recorded = d
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	m.Send("warning", "secret/baz", "msg")

	if recordedLevel != "warning" {
		t.Errorf("expected level 'warning', got %q", recordedLevel)
	}
	if recorded < 0 {
		t.Errorf("expected non-negative duration, got %v", recorded)
	}
}
