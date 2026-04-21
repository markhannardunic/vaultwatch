package config

import "time"

// DedupeConfig holds configuration for alert deduplication.
type DedupeConfig struct {
	// Enabled controls whether deduplication is active.
	Enabled bool `yaml:"enabled"`
	// WindowStr is the human-readable duration string (e.g. "1h", "30m").
	WindowStr string `yaml:"window"`
}

// Window parses WindowStr into a time.Duration.
// Returns a default of 1 hour if the string is empty or invalid.
func (d *DedupeConfig) Window() time.Duration {
	if d == nil || d.WindowStr == "" {
		return 1 * time.Hour
	}
	v, err := time.ParseDuration(d.WindowStr)
	if err != nil || v <= 0 {
		return 1 * time.Hour
	}
	return v
}

// DedupeConfigFromMain extracts DedupeConfig from the top-level Config,
// returning a default-enabled config if none is set.
func DedupeConfigFromMain(cfg *Config) *DedupeConfig {
	if cfg == nil {
		return &DedupeConfig{Enabled: true, WindowStr: "1h"}
	}
	if cfg.Dedupe == nil {
		return &DedupeConfig{Enabled: true, WindowStr: "1h"}
	}
	return cfg.Dedupe
}
