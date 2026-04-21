package notify

import (
	"errors"
	"testing"
	"time"
)

func makeRecord(path, level string, success bool, err error) DispatchRecord {
	return DispatchRecord{
		Path:    path,
		Level:   level,
		SentAt:  time.Now(),
		Success: success,
		Err:     err,
	}
}

func TestSnapshotStore_RecordAndSnapshot(t *testing.T) {
	s := NewSnapshotStore(10)
	s.Record(makeRecord("secret/a", "warning", true, nil))
	s.Record(makeRecord("secret/b", "critical", false, errors.New("timeout")))

	snap := s.Snapshot()
	if len(snap) != 2 {
		t.Fatalf("expected 2 records, got %d", len(snap))
	}
	if snap[0].Path != "secret/a" {
		t.Errorf("unexpected path: %s", snap[0].Path)
	}
	if snap[1].Success {
		t.Error("expected second record to be a failure")
	}
}

func TestSnapshotStore_EvictsOldestWhenFull(t *testing.T) {
	s := NewSnapshotStore(3)
	for i := 0; i < 4; i++ {
		s.Record(makeRecord("secret/x", "warning", true, nil))
	}
	snap := s.Snapshot()
	if len(snap) != 3 {
		t.Fatalf("expected 3 records after eviction, got %d", len(snap))
	}
}

func TestSnapshotStore_CountBySuccess(t *testing.T) {
	s := NewSnapshotStore(10)
	s.Record(makeRecord("a", "warning", true, nil))
	s.Record(makeRecord("b", "critical", true, nil))
	s.Record(makeRecord("c", "warning", false, errors.New("err")))

	ok, failed := s.CountBySuccess()
	if ok != 2 {
		t.Errorf("expected 2 successes, got %d", ok)
	}
	if failed != 1 {
		t.Errorf("expected 1 failure, got %d", failed)
	}
}

func TestSnapshotStore_Clear(t *testing.T) {
	s := NewSnapshotStore(10)
	s.Record(makeRecord("a", "warning", true, nil))
	s.Clear()
	if snap := s.Snapshot(); len(snap) != 0 {
		t.Errorf("expected empty snapshot after clear, got %d", len(snap))
	}
}

func TestSnapshotStore_DefaultMaxSize(t *testing.T) {
	s := NewSnapshotStore(0)
	if s.maxSize != 100 {
		t.Errorf("expected default maxSize 100, got %d", s.maxSize)
	}
}

func TestSnapshotStore_SnapshotReturnsCopy(t *testing.T) {
	s := NewSnapshotStore(10)
	s.Record(makeRecord("a", "warning", true, nil))
	snap := s.Snapshot()
	snap[0].Path = "mutated"

	original := s.Snapshot()
	if original[0].Path == "mutated" {
		t.Error("Snapshot should return an independent copy")
	}
}
