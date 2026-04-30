package config

import (
	"fmt"
	"time"
)

// ProbeConfig holds connectivity-probe settings.
type ProbeConfig struct {
	Enabled  bool          `yaml:"enabled"`
	Timeout  time.Duration `yaml:"timeout"`
	Interval time.Duration `yaml:"interval"`
}

// ProbeDefaults returns a ProbeConfig with sensible defaults.
func ProbeDefaults() ProbeConfig {
	return ProbeConfig{
		Enabled:  true,
		Timeout:  5 * time.Second,
		Interval: 60 * time.Second,
	}
}

// Validate checks that the ProbeConfig fields are within acceptable ranges.
// It returns an error if Timeout or Interval are non-positive when probing is enabled.
func (p ProbeConfig) Validate() error {
	if !p.Enabled {
		return nil
	}
	if p.Timeout <= 0 {
		return fmt.Errorf("probe.timeout must be positive, got %s", p.Timeout)
	}
	if p.Interval <= 0 {
		return fmt.Errorf("probe.interval must be positive, got %s", p.Interval)
	}
	if p.Timeout >= p.Interval {
		return fmt.Errorf("probe.timeout (%s) must be less than probe.interval (%s)", p.Timeout, p.Interval)
	}
	return nil
}

// ProbeConfigFromMain extracts ProbeConfig from the raw config map.
// It falls back to defaults for any missing or zero-value field.
func ProbeConfigFromMain(raw map[string]interface{}) (ProbeConfig, error) {
	cfg := ProbeDefaults()
	if raw == nil {
		return cfg, nil
	}
	v, ok := raw["probe"]
	if !ok {
		return cfg, nil
	}
	m, ok := v.(map[string]interface{})
	if !ok {
		return cfg, fmt.Errorf("probe: expected map, got %T", v)
	}
	if enabled, ok := m["enabled"].(bool); ok {
		cfg.Enabled = enabled
	}
	if s, ok := m["timeout"].(string); ok {
		d, err := time.ParseDuration(s)
		if err != nil {
			return cfg, fmt.Errorf("probe.timeout: %w", err)
		}
		cfg.Timeout = d
	}
	if s, ok := m["interval"].(string); ok {
		d, err := time.ParseDuration(s)
		if err != nil {
			return cfg, fmt.Errorf("probe.interval: %w", err)
		}
		cfg.Interval = d
	}
	return cfg, nil
}
