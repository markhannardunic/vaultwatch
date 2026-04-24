package config

import (
	"fmt"
	"time"
)

// AckConfig holds configuration for the acknowledgement store.
type AckConfig struct {
	// TTL is how long an alert remains acknowledgeable before it expires.
	TTL time.Duration
}

// AckDefaults returns a sensible default AckConfig.
func AckDefaults() AckConfig {
	return AckConfig{
		TTL: 24 * time.Hour,
	}
}

// AckConfigFromMain extracts AckConfig from the top-level config map.
// The config map may contain an "ack" sub-map with a "ttl" duration string.
func AckConfigFromMain(cfg map[string]interface{}) (AckConfig, error) {
	defaults := AckDefaults()
	if cfg == nil {
		return defaults, nil
	}
	raw, ok := cfg["ack"]
	if !ok {
		return defaults, nil
	}
	sub, ok := raw.(map[string]interface{})
	if !ok {
		return defaults, fmt.Errorf("ack config: expected map, got %T", raw)
	}
	if v, ok := sub["ttl"]; ok {
		s, ok := v.(string)
		if !ok {
			return defaults, fmt.Errorf("ack config: ttl must be a string, got %T", v)
		}
		d, err := time.ParseDuration(s)
		if err != nil {
			return defaults, fmt.Errorf("ack config: invalid ttl %q: %w", s, err)
		}
		if d <= 0 {
			return defaults, fmt.Errorf("ack config: ttl must be positive, got %s", d)
		}
		defaults.TTL = d
	}
	return defaults, nil
}
