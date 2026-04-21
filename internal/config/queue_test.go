package config_test

import (
	"testing"

	"github.com/youorg/vaultwatch/internal/config"
)

func TestQueueConfigDefaults(t *testing.T) {
	d := config.QueueConfigDefaults()
	if d.MaxSize != 256 {
		t.Errorf("expected MaxSize 256, got %d", d.MaxSize)
	}
	if d.Workers != 2 {
		t.Errorf("expected Workers 2, got %d", d.Workers)
	}
	if !d.DrainOnStop {
		t.Error("expected DrainOnStop to be true")
	}
	if d.StopTimeout != "5s" {
		t.Errorf("expected StopTimeout '5s', got %s", d.StopTimeout)
	}
}

func TestQueueConfigFromMain_NilConfig(t *testing.T) {
	cfg, err := config.QueueConfigFromMain(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.MaxSize != 256 {
		t.Errorf("expected default MaxSize 256, got %d", cfg.MaxSize)
	}
}

func TestQueueConfigFromMain_NoQueueKey(t *testing.T) {
	cfg, err := config.QueueConfigFromMain(&config.Config{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Workers != 2 {
		t.Errorf("expected default Workers 2, got %d", cfg.Workers)
	}
}

func TestQueueConfigFromMain_ValidConfig(t *testing.T) {
	main := &config.Config{
		Queue: &config.QueueConfig{
			MaxSize:     64,
			Workers:     4,
			DrainOnStop: false,
			StopTimeout: "10s",
		},
	}
	cfg, err := config.QueueConfigFromMain(main)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.MaxSize != 64 {
		t.Errorf("expected MaxSize 64, got %d", cfg.MaxSize)
	}
	if cfg.Workers != 4 {
		t.Errorf("expected Workers 4, got %d", cfg.Workers)
	}
	if cfg.DrainOnStop {
		t.Error("expected DrainOnStop false")
	}
}

func TestQueueConfigFromMain_InvalidDuration(t *testing.T) {
	main := &config.Config{
		Queue: &config.QueueConfig{
			MaxSize:     10,
			Workers:     1,
			StopTimeout: "not-a-duration",
		},
	}
	_, err := config.QueueConfigFromMain(main)
	if err == nil {
		t.Fatal("expected error for invalid stop_timeout")
	}
}

func TestQueueConfigFromMain_ZeroValuesUseDefaults(t *testing.T) {
	main := &config.Config{
		Queue: &config.QueueConfig{},
	}
	cfg, err := config.QueueConfigFromMain(main)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.MaxSize != 256 {
		t.Errorf("expected default MaxSize 256, got %d", cfg.MaxSize)
	}
	if cfg.Workers != 2 {
		t.Errorf("expected default Workers 2, got %d", cfg.Workers)
	}
}
