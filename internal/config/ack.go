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

// Validate checks that the AckConfig values are within acceptable bounds.
// It returns an error if any field is invalid.
func (c AckConfig) Validate() error {
	if c.TTL <= 0 {
		return fmt.Errorf("ack config: ttl must be positive, got %s", c.TTL)
	}
	if c.TTL > 30*24*time.Hour {
		return fmt.Errorf("ack config: ttl must not exceed 30 days, got %s", c.TTL)
	}
	return nil
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
		defaults.TTL = d
	}
	if err := defaults.Validate(); err != nil {
		return AckDefaults(), err
	}
	return defaults, nil
}
