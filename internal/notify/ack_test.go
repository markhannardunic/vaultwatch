package notify

import (
	"testing"
	"time"
)

func TestNewAckStore_InvalidTTL(t *testing.T) {
	_, err := NewAckStore(0)
	if err == nil {
		t.Fatal("expected error for zero TTL")
	}
	_, err = NewAckStore(-1 * time.Second)
	if err == nil {
		t.Fatal("expected error for negative TTL")
	}
}

func TestNewAckStore_ValidTTL(t *testing.T) {
	s, err := NewAckStore(5 * time.Minute)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s == nil {
		t.Fatal("expected non-nil store")
	}
}

func TestAckStore_Track_And_Snapshot(t *testing.T) {
	s, _ := NewAckStore(5 * time.Minute)
	s.Track("secret/db", "warning")

	snap := s.Snapshot()
	if len(snap) != 1 {
		t.Fatalf("expected 1 record, got %d", len(snap))
	}
	if snap[0].Path != "secret/db" {
		t.Errorf("expected path 'secret/db', got %q", snap[0].Path)
	}
	if snap[0].Status != AckPending {
		t.Errorf("expected status Pending, got %v", snap[0].Status)
	}
}

func TestAckStore_Acknowledge_Success(t *testing.T) {
	s, _ := NewAckStore(5 * time.Minute)
	s.Track("secret/api", "critical")

	if err := s.Acknowledge("secret/api"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	snap := s.Snapshot()
	if snap[0].Status != AckAcknowledged {
		t.Errorf("expected Acknowledged, got %v", snap[0].Status)
	}
	if snap[0].AckedAt == nil {
		t.Error("expected AckedAt to be set")
	}
}

func TestAckStore_Acknowledge_NotFound(t *testing.T) {
	s, _ := NewAckStore(5 * time.Minute)
	err := s.Acknowledge("secret/missing")
	if err == nil {
		t.Fatal("expected error for unknown path")
	}
}

func TestAckStore_Acknowledge_Expired(t *testing.T) {
	s, _ := NewAckStore(1 * time.Millisecond)
	s.Track("secret/old", "warning")
	time.Sleep(5 * time.Millisecond)

	err := s.Acknowledge("secret/old")
	if err == nil {
		t.Fatal("expected error for expired ack window")
	}
}

func TestAckStore_Purge_RemovesAckedAndExpired(t *testing.T) {
	s, _ := NewAckStore(1 * time.Millisecond)
	s.Track("secret/a", "warning")
	s.Track("secret/b", "critical")
	_ = s.Acknowledge("secret/b")
	time.Sleep(5 * time.Millisecond)

	s.Purge()
	snap := s.Snapshot()
	if len(snap) != 0 {
		t.Errorf("expected 0 records after purge, got %d", len(snap))
	}
}
