package alert

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestClassify(t *testing.T) {
	n := NewNotifier(nil)

	cases := []struct {
		ttl      time.Duration
		want     Severity
	}{
		{12 * time.Hour, SeverityCritical},
		{24 * time.Hour, SeverityCritical},
		{48 * time.Hour, SeverityWarning},
		{72 * time.Hour, SeverityWarning},
		{100 * time.Hour, SeverityInfo},
	}

	for _, c := range cases {
		got := n.Classify(c.ttl)
		if got != c.want {
			t.Errorf("Classify(%v) = %v, want %v", c.ttl, got, c.want)
		}
	}
}

func TestSend_WritesFormattedAlert(t *testing.T) {
	var buf bytes.Buffer
	n := NewNotifier(&buf)

	expires := time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC)
	a := Alert{
		Path:      "secret/db/password",
		ExpiresAt: expires,
		TTL:       20 * time.Hour,
		Severity:  SeverityCritical,
	}

	if err := n.Send(a); err != nil {
		t.Fatalf("Send() error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "[CRITICAL]") {
		t.Errorf("expected CRITICAL in output, got: %s", output)
	}
	if !strings.Contains(output, "secret/db/password") {
		t.Errorf("expected path in output, got: %s", output)
	}
	if !strings.Contains(output, "2024-06-01T12:00:00Z") {
		t.Errorf("expected expiry timestamp in output, got: %s", output)
	}
}

func TestNewNotifier_DefaultsToStdout(t *testing.T) {
	n := NewNotifier(nil)
	if n.writer == nil {
		t.Error("expected non-nil writer")
	}
}
