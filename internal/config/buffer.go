package config

import (
	"fmt"
	"time"
)

// BufferConfig holds raw config values for the buffered sender.
type BufferConfig struct {
	MaxSize    int    `yaml:"max_size"`
	FlushEvery string `yaml:"flush_every"`
	DropOnFull bool   `yaml:"drop_on_full"`
}

// BufferNotifyConfig is the top-level key under notify.
type BufferNotifyConfig struct {
	Buffer *BufferConfig `yaml:"buffer"`
}

// BufferDefaults returns default values for buffer configuration.
func BufferDefaults() BufferConfig {
	return BufferConfig{
		MaxSize:    100,
		FlushEvery: "30s",
		DropOnFull: false,
	}
}

// ParsedBuffer holds parsed, ready-to-use buffer settings.
type ParsedBuffer struct {
	MaxSize    int
	FlushEvery time.Duration
	DropOnFull bool
}

// BufferConfigFromMain extracts and validates buffer config from the main Config.
func BufferConfigFromMain(cfg *Config) (ParsedBuffer, error) {
	defaults := BufferDefaults()
	if cfg == nil || cfg.Notify == nil {
		d, _ := time.ParseDuration(defaults.FlushEvery)
		return ParsedBuffer{MaxSize: defaults.MaxSize, FlushEvery: d, DropOnFull: defaults.DropOnFull}, nil
	}

	raw, ok := cfg.Notify["buffer"]
	if !ok || raw == nil {
		d, _ := time.ParseDuration(defaults.FlushEvery)
		return ParsedBuffer{MaxSize: defaults.MaxSize, FlushEvery: d, DropOnFull: defaults.DropOnFull}, nil
	}

	bc, ok := raw.(*BufferConfig)
	if !ok {
		return ParsedBuffer{}, fmt.Errorf("buffer: invalid config type")
	}

	size := bc.MaxSize
	if size <= 0 {
		size = defaults.MaxSize
	}

	intervalStr := bc.FlushEvery
	if intervalStr == "" {
		intervalStr = defaults.FlushEvery
	}

	d, err := time.ParseDuration(intervalStr)
	if err != nil {
		return ParsedBuffer{}, fmt.Errorf("buffer: invalid flush_every %q: %w", intervalStr, err)
	}

	return ParsedBuffer{
		MaxSize:    size,
		FlushEvery: d,
		DropOnFull: bc.DropOnFull,
	}, nil
}
