package config

import "fmt"

// PoolConfig holds worker-pool settings sourced from the main config.
type PoolConfig struct {
	Workers   int `yaml:"workers"`
	QueueSize int `yaml:"queue_size"`
}

// PoolDefaults returns safe default values for the worker pool.
func PoolDefaults() PoolConfig {
	return PoolConfig{
		Workers:   4,
		QueueSize: 64,
	}
}

// PoolConfigFromMain extracts PoolConfig from the top-level Config,
// falling back to defaults for any zero values.
func PoolConfigFromMain(cfg *Config) (PoolConfig, error) {
	out := PoolDefaults()
	if cfg == nil {
		return out, nil
	}

	raw, ok := cfg.Extra["pool"]
	if !ok {
		return out, nil
	}

	m, ok := raw.(map[string]interface{})
	if !ok {
		return out, fmt.Errorf("pool: config block must be a mapping")
	}

	if v, ok := m["workers"]; ok {
		n, err := toInt(v)
		if err != nil {
			return out, fmt.Errorf("pool.workers: %w", err)
		}
		if n < 1 {
			return out, fmt.Errorf("pool.workers must be >= 1, got %d", n)
		}
		out.Workers = n
	}

	if v, ok := m["queue_size"]; ok {
		n, err := toInt(v)
		if err != nil {
			return out, fmt.Errorf("pool.queue_size: %w", err)
		}
		if n < 1 {
			return out, fmt.Errorf("pool.queue_size must be >= 1, got %d", n)
		}
		out.QueueSize = n
	}

	return out, nil
}
