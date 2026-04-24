package config

import (
	"testing"
	"time"
)

func TestProbeDefaults(t *testing.T) {
	cfg := ProbeDefaults()
	if !cfg.Enabled {
		t.Error("expected enabled to default to true")
	}
	if cfg.Timeout != 5*time.Second {
		t.Errorf("unexpected default timeout: %v", cfg.Timeout)
	}
	if cfg.Interval != 60*time.Second {
		t.Errorf("unexpected default interval: %v", cfg.Interval)
	}
}

func TestProbeConfigFromMain_NilConfig(t *testing.T) {
	cfg, err := ProbeConfigFromMain(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Timeout != 5*time.Second {
		t.Errorf("expected default timeout, got %v", cfg.Timeout)
	}
}

func TestProbeConfigFromMain_NoProbeKey(t *testing.T) {
	cfg, err := ProbeConfigFromMain(map[string]interface{}{"other": "val"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Interval != 60*time.Second {
		t.Errorf("expected default interval, got %v", cfg.Interval)
	}
}

func TestProbeConfigFromMain_ValidConfig(t *testing.T) {
	raw := map[string]interface{}{
		"probe": map[string]interface{}{
			"enabled":  false,
			"timeout":  "3s",
			"interval": "30s",
		},
	}
	cfg, err := ProbeConfigFromMain(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Enabled {
		t.Error("expected enabled=false")
	}
	if cfg.Timeout != 3*time.Second {
		t.Errorf("expected 3s timeout, got %v", cfg.Timeout)
	}
	if cfg.Interval != 30*time.Second {
		t.Errorf("expected 30s interval, got %v", cfg.Interval)
	}
}

func TestProbeConfigFromMain_InvalidDuration(t *testing.T) {
	raw := map[string]interface{}{
		"probe": map[string]interface{}{
			"timeout": "not-a-duration",
		},
	}
	_, err := ProbeConfigFromMain(raw)
	if err == nil {
		t.Fatal("expected error for invalid duration")
	}
}

func TestProbeConfigFromMain_BadType(t *testing.T) {
	raw := map[string]interface{}{
		"probe": "not-a-map",
	}
	_, err := ProbeConfigFromMain(raw)
	if err == nil {
		t.Fatal("expected error for non-map probe value")
	}
}
