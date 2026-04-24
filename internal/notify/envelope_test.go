package notify

import (
	"testing"
	"time"
)

func TestNewEnvelope_FieldsSet(t *testing.T) {
	before := time.Now().UTC()
	e := NewEnvelope("secret/db", "critical", "expires soon")
	after := time.Now().UTC()

	if e.Path != "secret/db" {
		t.Errorf("expected path secret/db, got %s", e.Path)
	}
	if e.Level != "critical" {
		t.Errorf("expected level critical, got %s", e.Level)
	}
	if e.Message != "expires soon" {
		t.Errorf("unexpected message: %s", e.Message)
	}
	if e.CreatedAt.Before(before) || e.CreatedAt.After(after) {
		t.Errorf("CreatedAt %v out of expected range", e.CreatedAt)
	}
	if e.Meta == nil {
		t.Error("expected non-nil Meta map")
	}
}

func TestEnvelope_Key_Format(t *testing.T) {
	e := NewEnvelope("secret/api", "warning", "msg")
	got := e.Key()
	want := "warning::secret/api"
	if got != want {
		t.Errorf("Key() = %q, want %q", got, want)
	}
}

func TestEnvelope_Expired_NoTTL(t *testing.T) {
	e := NewEnvelope("secret/x", "warning", "msg")
	if e.Expired() {
		t.Error("envelope with zero TTL should never expire")
	}
}

func TestEnvelope_Expired_WithinTTL(t *testing.T) {
	e := NewEnvelope("secret/x", "warning", "msg")
	e.TTL = 10 * time.Minute
	if e.Expired() {
		t.Error("envelope should not be expired within TTL")
	}
}

func TestEnvelope_Expired_PastTTL(t *testing.T) {
	e := NewEnvelope("secret/x", "warning", "msg")
	e.CreatedAt = time.Now().UTC().Add(-5 * time.Minute)
	e.TTL = 1 * time.Minute
	if !e.Expired() {
		t.Error("envelope should be expired past TTL")
	}
}

func TestEnvelope_WithMeta_Chaining(t *testing.T) {
	e := NewEnvelope("secret/y", "critical", "msg").
		WithMeta("team", "platform").
		WithMeta("env", "prod")

	if e.Meta["team"] != "platform" {
		t.Errorf("expected meta team=platform, got %s", e.Meta["team"])
	}
	if e.Meta["env"] != "prod" {
		t.Errorf("expected meta env=prod, got %s", e.Meta["env"])
	}
}
