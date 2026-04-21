package notify

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func makeRateLimitedRouter(t *testing.T, max int, window time.Duration) (*RateLimitedRouter, *bytes.Buffer) {
	t.Helper()
	buf := &bytes.Buffer{}
	router := NewRouter()
	router.Register("*", NewWriterSender(buf))
	rr, err := NewRateLimitedRouter(router, max, window)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	return rr, buf
}

func TestRateLimitedRouter_FirstDispatch_Succeeds(t *testing.T) {
	rr, buf := makeRateLimitedRouter(t, 2, time.Minute)
	if err := rr.Dispatch("warning", "secret/foo", "expiring soon"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "expiring soon") {
		t.Error("expected message in output")
	}
}

func TestRateLimitedRouter_SecondDispatch_Suppressed(t *testing.T) {
	rr, buf := makeRateLimitedRouter(t, 1, time.Minute)
	rr.Dispatch("warning", "secret/foo", "first")
	buf.Reset()
	rr.Dispatch("warning", "secret/foo", "second")
	if buf.Len() != 0 {
		t.Error("expected second dispatch to be suppressed")
	}
}

func TestRateLimitedRouter_DifferentPaths_BothSent(t *testing.T) {
	rr, buf := makeRateLimitedRouter(t, 1, time.Minute)
	rr.Dispatch("warning", "secret/a", "msg-a")
	rr.Dispatch("warning", "secret/b", "msg-b")
	out := buf.String()
	if !strings.Contains(out, "msg-a") || !strings.Contains(out, "msg-b") {
		t.Error("expected both paths to produce output")
	}
}

func TestNewRateLimitedRouter_InvalidConfig_ReturnsError(t *testing.T) {
	router := NewRouter()
	_, err := NewRateLimitedRouter(router, 0, time.Minute)
	if err == nil {
		t.Fatal("expected error for max=0")
	}
}
