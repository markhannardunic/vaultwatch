package audit_test

import (
	"testing"

	"github.com/vaultwatch/internal/audit"
)

func TestRecord_Add_And_Entries(t *testing.T) {
	r := audit.NewRecord()
	r.Add(audit.Entry{Path: "secret/a", Level: "warning"})
	r.Add(audit.Entry{Path: "secret/b", Level: "critical"})

	entries := r.Entries()
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	if entries[0].Path != "secret/a" {
		t.Errorf("expected secret/a, got %s", entries[0].Path)
	}
}

func TestRecord_CountByLevel(t *testing.T) {
	r := audit.NewRecord()
	r.Add(audit.Entry{Level: "warning"})
	r.Add(audit.Entry{Level: "warning"})
	r.Add(audit.Entry{Level: "critical"})
	r.Add(audit.Entry{Level: "healthy"})

	counts := r.CountByLevel()
	if counts["warning"] != 2 {
		t.Errorf("expected 2 warnings, got %d", counts["warning"])
	}
	if counts["critical"] != 1 {
		t.Errorf("expected 1 critical, got %d", counts["critical"])
	}
	if counts["healthy"] != 1 {
		t.Errorf("expected 1 healthy, got %d", counts["healthy"])
	}
}

func TestRecord_Entries_ReturnsCopy(t *testing.T) {
	r := audit.NewRecord()
	r.Add(audit.Entry{Path: "secret/x", Level: "healthy"})

	e1 := r.Entries()
	e1[0].Path = "mutated"

	e2 := r.Entries()
	if e2[0].Path == "mutated" {
		t.Error("Entries should return a copy, not a reference")
	}
}

func TestRecord_TimestampStamped(t *testing.T) {
	r := audit.NewRecord()
	r.Add(audit.Entry{Path: "secret/y", Level: "warning"})
	entries := r.Entries()
	if entries[0].Timestamp.IsZero() {
		t.Error("expected timestamp to be set automatically")
	}
}
