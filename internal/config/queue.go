package config

import "time"

// QueueConfig mirrors notify.QueueConfig for YAML unmarshalling.
type QueueConfig struct {
	MaxSize     int    `yaml:"max_size"`
	Workers     int    `yaml:"workers"`
	DrainOnStop bool   `yaml:"drain_on_stop"`
	StopTimeout string `yaml:"stop_timeout"`
}

// QueueDefaults returns a QueueConfig with sensible defaults.
func QueueConfigDefaults() QueueConfig {
	return QueueConfig{
		MaxSize:     256,
		Workers:     2,
		DrainOnStop: true,
		StopTimeout: "5s",
	}
}

// QueueConfigFromMain extracts queue settings from the top-level Config,
// falling back to defaults for any zero values.
func QueueConfigFromMain(cfg *Config) (QueueConfig, error) {
	defaults := QueueConfigDefaults()
	if cfg == nil || cfg.Queue == nil {
		return defaults, nil
	}
	out := *cfg.Queue
	if out.MaxSize <= 0 {
		out.MaxSize = defaults.MaxSize
	}
	if out.Workers <= 0 {
		out.Workers = defaults.Workers
	}
	if out.StopTimeout == "" {
		out.StopTimeout = defaults.StopTimeout
	}
	// validate parseable
	if _, err := time.ParseDuration(out.StopTimeout); err != nil {
		return QueueConfig{}, err
	}
	return out, nil
}
