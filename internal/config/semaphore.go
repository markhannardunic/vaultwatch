package config

import "time"

// SemaphoreConfig controls concurrency limiting for outbound notifications.
type SemaphoreConfig struct {
	// MaxConcurrent is the maximum number of simultaneous Send calls allowed.
	MaxConcurrent int `yaml:"max_concurrent"`
	// AcquireTimeout is how long to wait for a concurrency slot before failing.
	AcquireTimeout time.Duration `yaml:"acquire_timeout"`
}

// SemaphoreDefaults returns a SemaphoreConfig with sensible defaults.
func SemaphoreDefaults() SemaphoreConfig {
	return SemaphoreConfig{
		MaxConcurrent:  5,
		AcquireTimeout: 2 * time.Second,
	}
}

// SemaphoreConfigFromMain extracts semaphore config from the top-level config
// map, falling back to defaults for any missing keys.
func SemaphoreConfigFromMain(raw map[string]interface{}) SemaphoreConfig {
	cfg := SemaphoreDefaults()
	if raw == nil {
		return cfg
	}
	sec, ok := raw["semaphore"]
	if !ok {
		return cfg
	}
	m, ok := sec.(map[string]interface{})
	if !ok {
		return cfg
	}
	if v, ok := m["max_concurrent"].(int); ok && v > 0 {
		cfg.MaxConcurrent = v
	}
	if v, ok := m["acquire_timeout"].(string); ok {
		if d, err := time.ParseDuration(v); err == nil {
			cfg.AcquireTimeout = d
		}
	}
	return cfg
}
