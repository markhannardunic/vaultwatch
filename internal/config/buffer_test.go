package config

import (
	"testing"
	"time"
)

func TestBufferDefaults(t *testing.T) {
	d := BufferDefaults()
	if d.MaxSize != 100 {
		t.Errorf("expected MaxSize 100, got %d", d.MaxSize)
	}
	if d.FlushEvery != "30s" {
		t.Errorf("expected FlushEvery '30s', got %q", d.FlushEvery)
	}
	if d.DropOnFull {
		t.Error("expected DropOnFull to be false by default")
	}
}

func TestBufferConfigFromMain_NilConfig(t *testing.T) {
	p, err := BufferConfigFromMain(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.MaxSize != 100 {
		t.Errorf("expected default MaxSize 100, got %d", p.MaxSize)
	}
	if p.FlushEvery != 30*time.Second {
		t.Errorf("expected 30s, got %v", p.FlushEvery)
	}
}

func TestBufferConfigFromMain_NoBufferKey(t *testing.T) {
	cfg := &Config{Notify: map[string]interface{}{}}
	p, err := BufferConfigFromMain(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.MaxSize != 100 {
		t.Errorf("expected default MaxSize, got %d", p.MaxSize)
	}
}

func TestBufferConfigFromMain_ValidConfig(t *testing.T) {
	cfg := &Config{
		Notify: map[string]interface{}{
			"buffer": &BufferConfig{
				MaxSize:    25,
				FlushEvery: "10s",
				DropOnFull: true,
			},
		},
	}
	p, err := BufferConfigFromMain(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.MaxSize != 25 {
		t.Errorf("expected MaxSize 25, got %d", p.MaxSize)
	}
	if p.FlushEvery != 10*time.Second {
		t.Errorf("expected 10s, got %v", p.FlushEvery)
	}
	if !p.DropOnFull {
		t.Error("expected DropOnFull true")
	}
}

func TestBufferConfigFromMain_InvalidDuration(t *testing.T) {
	cfg := &Config{
		Notify: map[string]interface{}{
			"buffer": &BufferConfig{
				MaxSize:    10,
				FlushEvery: "not-a-duration",
			},
		},
	}
	_, err := BufferConfigFromMain(cfg)
	if err == nil {
		t.Fatal("expected error for invalid duration")
	}
}

func TestBufferConfigFromMain_ZeroMaxSize_UsesDefault(t *testing.T) {
	cfg := &Config{
		Notify: map[string]interface{}{
			"buffer": &BufferConfig{
				MaxSize:    0,
				FlushEvery: "5s",
			},
		},
	}
	p, err := BufferConfigFromMain(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.MaxSize != 100 {
		t.Errorf("expected fallback MaxSize 100, got %d", p.MaxSize)
	}
}
