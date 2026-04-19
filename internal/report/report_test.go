package report

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func makeEntry(path, level string, days int) Entry {
	return Entry{
		Path:     path,
		Level:    level,
		Expiry:   time.Now().AddDate(0, 0, days),
		DaysLeft: days,
	}
}

func TestReport_Render_ContainsEntries(t *testing.T) {
	r := New()
	r.entries = []Entry{
		makeEntry("secret/db", "critical", 3),
		makeEntry("secret/api", "warning", 20),
	}
	var buf bytes.Buffer
	r.Render(&buf)
	out := buf.String()
	if !strings.Contains(out, "secret/db") {
		t.Error("expected secret/db in output")
	}
	if !strings.Contains(out, "secret/api") {
		t.Error("expected secret/api in output")
	}
}

func TestReport_Summary_Counts(t *testing.T) {
	r := New()
	r.entries = []Entry{
		makeEntry("a", "critical", 2),
		makeEntry("b", "critical", 4),
		makeEntry("c", "warning", 15),
		makeEntry("d", "healthy", 60),
	}
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
	r := New()
	r.entries = []Entry{makeEntry("secret/x", "warning", 10)}
	var buf bytes.Buffer
	r.Render(&buf)
	if !strings.Contains(buf.String(), "Summary:") {
		t.Error("expected Summary line in output")
	}
}

func TestReport_Empty_Render(t *testing.T) {
	r := New()
	var buf bytes.Buffer
	r.Render(&buf)
	if !strings.Contains(buf.String(), "No secrets audited") {
		t.Error("expected empty message")
	}
}
