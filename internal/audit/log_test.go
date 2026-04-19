package audit_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/vaultwatch/internal/audit"
)

func TestLogger_Text_ContainsPath(t *testing.T) {
	var buf bytes.Buffer
	l := audit.NewLogger(&buf, "text")
	e := audit.Entry{
		Timestamp: time.Now().UTC(),
		Path:      "secret/db",
		Level:     "warning",
		Message:   "expiring soon",
		TTL:       172800,
	}
	if err := l.Write(e); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "secret/db") {
		t.Errorf("expected path in output, got: %s", buf.String())
	}
}

func TestLogger_JSON_ValidStructure(t *testing.T) {
	var buf bytes.Buffer
	l := audit.NewLogger(&buf, "json")
	e := audit.Entry{
		Path:    "secret/api",
		Level:   "critical",
		Message: "expires soon",
		TTL:     3600,
	}
	if err := l.Write(e); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var out audit.Entry
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if out.Path != "secret/api" {
		t.Errorf("expected path secret/api, got %s", out.Path)
	}
}

func TestLogger_DefaultsToText(t *testing.T) {
	var buf bytes.Buffer
	l := audit.NewLogger(&buf, "")
	e := audit.Entry{Path: "secret/x", Level: "healthy", TTL: 9999}
	if err := l.Write(e); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "secret/x") {
		t.Errorf("expected text output, got: %s", buf.String())
	}
}

func TestLogger_NilWriter_UsesStdout(t *testing.T) {
	l := audit.NewLogger(nil, "text")
	if l == nil {
		t.Fatal("expected non-nil logger")
	}
}
