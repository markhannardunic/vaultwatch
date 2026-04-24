package notify

import (
	"errors"
	"testing"
)

func TestNewDeadLetterStore_InvalidMaxSize(t *testing.T) {
	_, err := NewDeadLetterStore(0)
	if err == nil {
		t.Fatal("expected error for maxSize=0")
	}
}

func TestNewDeadLetterStore_Valid(t *testing.T) {
	s, err := NewDeadLetterStore(10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s == nil {
		t.Fatal("expected non-nil store")
	}
}

func TestDeadLetterStore_Record_And_Snapshot(t *testing.T) {
	s, _ := NewDeadLetterStore(5)
	s.Record("secret/foo", "foo expires soon", "warning", 3, errors.New("timeout"))

	snap := s.Snapshot()
	if len(snap) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(snap))
	}
	e := snap[0]
	if e.Path != "secret/foo" {
		t.Errorf("unexpected path: %s", e.Path)
	}
	if e.Level != "warning" {
		t.Errorf("unexpected level: %s", e.Level)
	}
	if e.Attempts != 3 {
		t.Errorf("unexpected attempts: %d", e.Attempts)
	}
	if e.Err == nil || e.Err.Error() != "timeout" {
		t.Errorf("unexpected error: %v", e.Err)
	}
	if e.FailedAt.IsZero() {
		t.Error("expected FailedAt to be set")
	}
}

func TestDeadLetterStore_EvictsOldestWhenFull(t *testing.T) {
	s, _ := NewDeadLetterStore(3)
	for i := 0; i < 4; i++ {
		s.Record("secret/x", "msg", "critical", 1, errors.New("err"))
	}
	if s.Len() != 3 {
		t.Fatalf("expected 3 entries after eviction, got %d", s.Len())
	}
}

func TestDeadLetterStore_Clear(t *testing.T) {
	s, _ := NewDeadLetterStore(5)
	s.Record("secret/a", "msg", "warning", 1, errors.New("err"))
	s.Record("secret/b", "msg", "critical", 2, errors.New("err"))
	s.Clear()
	if s.Len() != 0 {
		t.Fatalf("expected 0 entries after clear, got %d", s.Len())
	}
}

func TestDeadLetterStore_Snapshot_ReturnsCopy(t *testing.T) {
	s, _ := NewDeadLetterStore(5)
	s.Record("secret/a", "msg", "warning", 1, errors.New("fail"))
	snap := s.Snapshot()
	snap[0].Path = "mutated"

	original := s.Snapshot()
	if original[0].Path == "mutated" {
		t.Error("snapshot should be a copy, not a reference")
	}
}
