package config

import "fmt"

// DeadLetterConfig controls the dead-letter store behaviour.
type DeadLetterConfig struct {
	MaxSize int `yaml:"max_size"`
}

// DeadLetterDefaults returns a DeadLetterConfig populated with sensible defaults.
func DeadLetterDefaults() DeadLetterConfig {
	return DeadLetterConfig{
		MaxSize: 200,
	}
}

// DeadLetterConfigFromMain extracts DeadLetterConfig from the top-level Config.
// If no dead_letter section is present the defaults are returned.
func DeadLetterConfigFromMain(cfg *Config) (DeadLetterConfig, error) {
	defaults := DeadLetterDefaults()
	if cfg == nil {
		return defaults, nil
	}

	raw, ok := cfg.Extra["dead_letter"]
	if !ok {
		return defaults, nil
	}

	section, ok := raw.(map[string]interface{})
	if !ok {
		return defaults, fmt.Errorf("dead_letter: invalid configuration block")
	}

	out := defaults
	if v, ok := section["max_size"]; ok {
		switch n := v.(type) {
		case int:
			out.MaxSize = n
		case float64:
			out.MaxSize = int(n)
		default:
			return defaults, fmt.Errorf("dead_letter: max_size must be an integer")
		}
		if out.MaxSize <= 0 {
			return defaults, fmt.Errorf("dead_letter: max_size must be greater than zero")
		}
	}

	return out, nil
}
