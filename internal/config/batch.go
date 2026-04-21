package config

import (
	"fmt"
	"time"
)

// BatchConfig holds configuration for the batch notification sender.
type BatchConfig struct {
	Enabled bool          `yaml:"enabled"`
	MaxSize int           `yaml:"max_size"`
	MaxWait time.Duration `yaml:"max_wait"`
}

// BatchDefaults returns a BatchConfig populated with sensible defaults.
func BatchDefaults() BatchConfig {
	return BatchConfig{
		Enabled: false,
		MaxSize: 10,
		MaxWait: 30 * time.Second,
	}
}

// BatchConfigFromMain extracts and validates the batch section of MainConfig.
func BatchConfigFromMain(m *MainConfig) (BatchConfig, error) {
	if m == nil || m.Notify == nil {
		return BatchDefaults(), nil
	}
	cfg := BatchDefaults()
	if m.Notify.Batch == nil {
		return cfg, nil
	}
	b := m.Notify.Batch
	if b.MaxSize > 0 {
		cfg.MaxSize = b.MaxSize
	}
	if b.MaxWait > 0 {
		cfg.MaxWait = b.MaxWait
	}
	cfg.Enabled = b.Enabled
	if cfg.MaxSize < 1 {
		return cfg, fmt.Errorf("batch: max_size must be at least 1")
	}
	if cfg.MaxWait < time.Second {
		return cfg, fmt.Errorf("batch: max_wait must be at least 1s")
	}
	return cfg, nil
}
