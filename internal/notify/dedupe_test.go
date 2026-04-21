package notify

import (
	"testing"
	"time"
)

func TestDedupeStore_FirstCallNotDuplicate(t *testing.T) {
	d := NewDedupeStore(1 * time.Hour)
	if d.IsDuplicate("key1") {
		t.Fatal("expected first call to not be a duplicate")
	}
}

func TestDedupeStore_SecondCallIsDuplicate(t *testing.T) {
	d := NewDedupeStore(1 * time.Hour)
	d.IsDuplicate("key1")
	if !d.IsDuplicate("key1") {
		t.Fatal("expected second call within window to be a duplicate")
	}
}

func TestDedupeStore_AfterWindowExpiry_NotDuplicate(t *testing.T) {
	d := NewDedupeStore(10 * time.Millisecond)
	d.IsDuplicate("key1")
	time.Sleep(20 * time.Millisecond)
	if d.IsDuplicate("key1") {
		t.Fatal("expected call after window expiry to not be a duplicate")
	}
}

func TestDedupeStore_DifferentKeys_Independent(t *testing.T) {
	d := NewDedupeStore(1 * time.Hour)
	d.IsDuplicate("key1")
	if d.IsDuplicate("key2") {
		t.Fatal("expected different key to not be a duplicate")
	}
}

func TestDedupeStore_Purge_RemovesExpired(t *testing.T) {
	d := NewDedupeStore(10 * time.Millisecond)
	d.IsDuplicate("old-key")
	time.Sleep(20 * time.Millisecond)
	d.Purge()
	d.mu.Lock()
	_, exists := d.seen["old-key"]
	d.mu.Unlock()
	if exists {
		t.Fatal("expected purge to remove expired entry")
	}
}

func TestFingerprint_Deterministic(t *testing.T) {
	a := Fingerprint("secret/db", "critical")
	b := Fingerprint("secret/db", "critical")
	if a != b {
		t.Fatalf("expected identical fingerprints, got %s and %s", a, b)
	}
}

func TestFingerprint_DifferentInputs_DifferentOutputs(t *testing.T) {
	a := Fingerprint("secret/db", "critical")
	b := Fingerprint("secret/db", "warning")
	if a == b {
		t.Fatal("expected different fingerprints for different inputs")
	}
}

func TestNewDedupeStore_ZeroWindow_UsesDefault(t *testing.T) {
	d := NewDedupeStore(0)
	if d.window != 1*time.Hour {
		t.Fatalf("expected default window of 1h, got %v", d.window)
	}
}
