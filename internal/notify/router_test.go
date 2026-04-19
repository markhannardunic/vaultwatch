package notify_test

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/vaultwatch/internal/notify"
)

type failSender struct{}

func (f *failSender) Send(_ string) error { return fmt.Errorf("send failed") }

func TestRouter_Dispatch_MatchingLevel(t *testing.T) {
	var buf bytes.Buffer
	r := notify.NewRouter()
	r.Register("critical", notify.NewWriterSender(&buf))

	if err := r.Dispatch("critical", "disk full"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "disk full") {
		t.Errorf("expected message in output, got: %s", buf.String())
	}
}

func TestRouter_Dispatch_WildcardReceivesAll(t *testing.T) {
	var buf bytes.Buffer
	r := notify.NewRouter()
	r.Register("*", notify.NewWriterSender(&buf))

	r.Dispatch("warning", "expiring soon")
	if !strings.Contains(buf.String(), "expiring soon") {
		t.Errorf("wildcard sender did not receive message")
	}
}

func TestRouter_Dispatch_NoMatchingLevel(t *testing.T) {
	var buf bytes.Buffer
	r := notify.NewRouter()
	r.Register("critical", notify.NewWriterSender(&buf))

	r.Dispatch("warning", "low priority")
	if buf.Len() != 0 {
		t.Errorf("expected no output for unregistered level")
	}
}

func TestRouter_Dispatch_CollectsErrors(t *testing.T) {
	r := notify.NewRouter()
	r.Register("critical", &failSender{})
	r.Register("critical", &failSender{})

	err := r.Dispatch("critical", "msg")
	if err == nil {
		t.Fatal("expected aggregated error")
	}
}
