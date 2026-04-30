package config

import (
	"fmt"
	"time"
)

// WatermarkConfig controls the level-change watermark sender.
type WatermarkConfig struct {
	// Enabled turns the watermark suppression on or off.
	Enabled bool `yaml:"enabled"`
	// Window is how long a level is remembered. Repeated alerts at the same
	// level within this window are suppressed. Defaults to 30 minutes.
	Window time.Duration `yaml:"window"`
}

// WatermarkDefaults returns a WatermarkConfig with sensible defaults.
func WatermarkDefaults() WatermarkConfig {
	return WatermarkConfig{
		Enabled: true,
		Window:  30 * time.Minute,
	}
}

// WatermarkConfigFromMain extracts watermark configuration from the top-level
// config map, falling back to defaults for any missing keys.
func WatermarkConfigFromMain(raw map[string]any) (WatermarkConfig, error) {
	cfg := WatermarkDefaults()
	if raw == nil {
		return cfg, nil
	}
	section, ok := raw["watermark"]
	if !ok {
		return cfg, nil
	}
	m, ok := section.(map[string]any)
	if !ok {
		return cfg, fmt.Errorf("watermark: config section must be a mapping")
	}
	if v, ok := m["enabled"]; ok {
		b, ok := v.(bool)
		if !ok {
			return cfg, fmt.Errorf("watermark: enabled must be a boolean")
		}
		cfg.Enabled = b
	}
	if v, ok := m["window"]; ok {
		s, ok := v.(string)
		if !ok {
			return cfg, fmt.Errorf("watermark: window must be a duration string")
		}
		d, err := time.ParseDuration(s)
		if err != nil {
			return cfg, fmt.Errorf("watermark: invalid window %q: %w", s, err)
		}
		if d <= 0 {
			return cfg, fmt.Errorf("watermark: window must be positive")
		}
		cfg.Window = d
	}
	return cfg, nil
}
