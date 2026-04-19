package output

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func makeEntry(path, status string, days int) Entry {
	return Entry{
		Path:      path,
		Status:    status,
		ExpiresAt: time.Now().Add(time.Duration(days) * 24 * time.Hour),
		DaysLeft:  days,
	}
}

func TestFormatter_Text_ContainsHeader(t *testing.T) {
	var buf bytes.Buffer
	f := NewFormatter(&buf, FormatText)
	_ = f.Write([]Entry{makeEntry("secret/db", "WARNING", 10)})
	out := buf.String()
	if !strings.Contains(out, "PATH") || !strings.Contains(out, "STATUS") {
		t.Errorf("expected header in output, got: %s", out)
	}
}

func TestFormatter_Text_ContainsEntry(t *testing.T) {
	var buf bytes.Buffer
	f := NewFormatter(&buf, FormatText)
	_ = f.Write([]Entry{makeEntry("secret/api", "CRITICAL", 3)})
	out := buf.String()
	if !strings.Contains(out, "secret/api") || !strings.Contains(out, "CRITICAL") {
		t.Errorf("expected entry in output, got: %s", out)
	}
}

func TestFormatter_JSON_ValidStructure(t *testing.T) {
	var buf bytes.Buffer
	f := NewFormatter(&buf, FormatJSON)
	_ = f.Write([]Entry{makeEntry("secret/token", "WARNING", 7)})
	var report jsonReport
	if err := json.Unmarshal(buf.Bytes(), &report); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if report.Total != 1 {
		t.Errorf("expected total=1, got %d", report.Total)
	}
	if report.Entries[0].Path != "secret/token" {
		t.Errorf("unexpected path: %s", report.Entries[0].Path)
	}
}

func TestFormatter_JSON_EmptyEntries(t *testing.T) {
	var buf bytes.Buffer
	f := NewFormatter(&buf, FormatJSON)
	_ = f.Write([]Entry{})
	var report jsonReport
	if err := json.Unmarshal(buf.Bytes(), &report); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if report.Total != 0 {
		t.Errorf("expected total=0, got %d", report.Total)
	}
}
