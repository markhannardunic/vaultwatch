package config

import (
	"testing"
	"time"
)

func TestAckDefaults(t *testing.T) {
	cfg := AckDefaults()
	if cfg.TTL != 24*time.Hour {
		t.Errorf("expected 24h TTL, got %s", cfg.TTL)
	}
}

func TestAckConfigFromMain_NilConfig(t *testing.T) {
	cfg, err := AckConfigFromMain(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.TTL != 24*time.Hour {
		t.Errorf("expected default TTL, got %s", cfg.TTL)
	}
}

func TestAckConfigFromMain_NoAckKey(t *testing.T) {
	cfg, err := AckConfigFromMain(map[string]interface{}{"other": "value"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.TTL != 24*time.Hour {
		t.Errorf("expected default TTL, got %s", cfg.TTL)
	}
}

func TestAckConfigFromMain_ValidConfig(t *testing.T) {
	m := map[string]interface{}{
		"ack": map[string]interface{}{
			"ttl": "12h",
		},
	}
	cfg, err := AckConfigFromMain(m)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.TTL != 12*time.Hour {
		t.Errorf("expected 12h TTL, got %s", cfg.TTL)
	}
}

func TestAckConfigFromMain_InvalidDuration(t *testing.T) {
	m := map[string]interface{}{
		"ack": map[string]interface{}{
			"ttl": "not-a-duration",
		},
	}
	_, err := AckConfigFromMain(m)
	if err == nil {
		t.Fatal("expected error for invalid duration")
	}
}

func TestAckConfigFromMain_ZeroTTL(t *testing.T) {
	m := map[string]interface{}{
		"ack": map[string]interface{}{
			"ttl": "0s",
		},
	}
	_, err := AckConfigFromMain(m)
	if err == nil {
		t.Fatal("expected error for zero TTL")
	}
}

func TestAckConfigFromMain_InvalidSubMap(t *testing.T) {
	m := map[string]interface{}{
		"ack": "not-a-map",
	}
	_, err := AckConfigFromMain(m)
	if err == nil {
		t.Fatal("expected error for non-map ack value")
	}
}
