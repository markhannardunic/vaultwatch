package config

import (
	"fmt"
	"time"
)

// TimeoutConfig holds per-sender timeout settings.
type TimeoutConfig struct {
	SendTimeout time.Duration
}

// TimeoutDefaults returns a TimeoutConfig with sensible defaults.
func TimeoutDefaults() TimeoutConfig {
	return TimeoutConfig{
		SendTimeout: 10 * time.Second,
	}
}

// TimeoutConfigFromMain extracts timeout configuration from the top-level
// config map, falling back to defaults when keys are absent.
func TimeoutConfigFromMain(raw map[string]any) (TimeoutConfig, error) {
	cfg := TimeoutDefaults()
	if raw == nil {
		return cfg, nil
	}
	timeoutRaw, ok := raw["timeout"]
	if !ok {
		return cfg, nil
	}
	timeoutMap, ok := timeoutRaw.(map[string]any)
	if !ok {
		return cfg, fmt.Errorf("timeout: expected map, got %T", timeoutRaw)
	}
	if v, ok := timeoutMap["send_timeout"]; ok {
		s, ok := v.(string)
		if !ok {
			return cfg, fmt.Errorf("timeout.send_timeout: expected string, got %T", v)
		}
		d, err := time.ParseDuration(s)
		if err != nil {
			return cfg, fmt.Errorf("timeout.send_timeout: invalid duration %q: %w", s, err)
		}
		if d <= 0 {
			return cfg, fmt.Errorf("timeout.send_timeout: must be positive, got %s", d)
		}
		cfg.SendTimeout = d
	}
	return cfg, nil
}
