package report_test

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/your-org/vaultwatch/internal/report"
)

func makeEntry(path, level string, hoursUntil float64) report.Entry {
	return report.Entry{
		Path:      path,
		Level:     level,
		ExpiresAt: time.Now().Add(time.Duration(hoursUntil * float64(time.Hour))),
		Message:   level + " alert for " + path,
	}
}

func TestReport_Render_ContainsEntries(t *testing.T) {
	var buf bytes.Buffer
	r := report.New(&buf)
	r.Add(makeEntry("secret/db", "critical", 2))
	r.Add(makeEntry("secret/api", "warning", 48))

	if err := r.Render(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "secret/db") {
		t.Error("expected secret/db in output")
	}
	if !strings.Contains(out, "secret/api") {
		t.Error("expected secret/api in output")
	}
}

func TestReport_Summary_Counts(t *testing.T) {
	r := report.New(nil)
	r.Add(makeEntry("a", "critical", 1))
	r.Add(makeEntry("b", "critical", 2))
	r.Add(makeEntry("c", "warning", 50))
	r.Add(makeEntry("d", "healthy", 200))

	s := r.Summary()
	if s["critical"] != 2 {
		t.Errorf("expected 2 critical, got %d", s["critical"])
	}
	if s["warning"] != 1 {
		t.Errorf("expected 1 warning, got %d", s["warning"])
	}
	if s["healthy"] != 1 {
		t.Errorf("expected 1 healthy, got %d", s["healthy"])
	}
}

func TestReport_Render_SummaryLine(t *testing.T) {
	var buf bytes.Buffer
	r := report.New(&buf)
	r.Add(makeEntry("x", "critical", 3))

	_ = r.Render()

	if !strings.Contains(buf.String(), "Summary:") {
		t.Error("expected Summary line in output")
	}
}

func TestReport_Empty_Render(t *testing.T) {
	var buf bytes.Buffer
	r := report.New(&buf)
	if err := r.Render(); err != nil {
		t.Fatalf("unexpected error on empty report: %v", err)
	}
}
