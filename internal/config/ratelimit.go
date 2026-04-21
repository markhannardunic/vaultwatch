package config

import (
	"fmt"
	"time"
)

// RateLimitConfig holds configuration for per-alert rate limiting.
type RateLimitConfig struct {
	Enabled bool   `yaml:"enabled"`
	Max     int    `yaml:"max"`
	Window  string `yaml:"window"`
}

// RateLimitDefaults returns a RateLimitConfig with sensible defaults.
func RateLimitDefaults() RateLimitConfig {
	return RateLimitConfig{
		Enabled: false,
		Max:     5,
		Window:  "1h",
	}
}

// WindowDuration parses the Window string into a time.Duration.
func (r RateLimitConfig) WindowDuration() (time.Duration, error) {
	if r.Window == "" {
		return 0, fmt.Errorf("ratelimit: window must not be empty")
	}
	d, err := time.ParseDuration(r.Window)
	if err != nil {
		return 0, fmt.Errorf("ratelimit: invalid window %q: %w", r.Window, err)
	}
	return d, nil
}

// Validate checks that the RateLimitConfig is valid when enabled.
func (r RateLimitConfig) Validate() error {
	if !r.Enabled {
		return nil
	}
	if r.Max <= 0 {
		return fmt.Errorf("ratelimit: max must be greater than zero")
	}
	if _, err := r.WindowDuration(); err != nil {
		return err
	}
	return nil
}
