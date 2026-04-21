package notify

import (
	"strings"
	"testing"
	"time"
)

func makeDedupedRouter(t *testing.T, window time.Duration) (*DedupedRouter, *strings.Builder) {
	t.Helper()
	buf := &strings.Builder{}
	sender := NewWriterSender(buf)
	router := NewRouter()
	router.Register("*", sender)
	return NewDedupedRouter(router, window), buf
}

func TestDedupedRouter_FirstDispatch_Succeeds(t *testing.T) {
	dr, buf := makeDedupedRouter(t, 1*time.Hour)
	err := dr.Dispatch("critical", "secret/db", "expires soon")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !strings.Contains(buf.String(), "expires soon") {
		t.Fatal("expected message to be written")
	}
}

func TestDedupedRouter_SecondDispatch_Suppressed(t *testing.T) {
	dr, _ := makeDedupedRouter(t, 1*time.Hour)
	dr.Dispatch("critical", "secret/db", "expires soon")
	err := dr.Dispatch("critical", "secret/db", "expires soon")
	if err == nil {
		t.Fatal("expected duplicate suppression error")
	}
	if !strings.Contains(err.Error(), "dedupe") {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestDedupedRouter_AfterWindowExpiry_AllowsResend(t *testing.T) {
	dr, _ := makeDedupedRouter(t, 10*time.Millisecond)
	dr.Dispatch("critical", "secret/db", "expires soon")
	time.Sleep(20 * time.Millisecond)
	err := dr.Dispatch("critical", "secret/db", "expires soon")
	if err != nil {
		t.Fatalf("expected resend after window, got %v", err)
	}
}

func TestDedupedRouter_DifferentPaths_BothSent(t *testing.T) {
	dr, buf := makeDedupedRouter(t, 1*time.Hour)
	dr.Dispatch("critical", "secret/db", "db alert")
	err := dr.Dispatch("critical", "secret/api", "api alert")
	if err != nil {
		t.Fatalf("expected different path to be sent, got %v", err)
	}
	if !strings.Contains(buf.String(), "api alert") {
		t.Fatal("expected second message to appear in output")
	}
}

func TestDedupedRouter_PurgeExpired_DoesNotPanic(t *testing.T) {
	dr, _ := makeDedupedRouter(t, 10*time.Millisecond)
	dr.Dispatch("warning", "secret/x", "msg")
	time.Sleep(20 * time.Millisecond)
	dr.PurgeExpired() // should not panic
}
