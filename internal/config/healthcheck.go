package config

import (
	"fmt"
	"time"
)

// HealthCheckConfig holds configuration for the sender health checker.
type HealthCheckConfig struct {
	Enabled  bool          `yaml:"enabled"`
	Interval time.Duration `yaml:"interval"`
}

// HealthCheckDefaults returns a HealthCheckConfig populated with safe defaults.
func HealthCheckDefaults() HealthCheckConfig {
	return HealthCheckConfig{
		Enabled:  true,
		Interval: 30 * time.Second,
	}
}

// HealthCheckConfigFromMain extracts and validates HealthCheckConfig from the
// top-level Config, applying defaults for any unset fields.
func HealthCheckConfigFromMain(cfg *Config) (HealthCheckConfig, error) {
	if cfg == nil {
		return HealthCheckDefaults(), nil
	}

	hc := cfg.HealthCheck

	if hc.Interval == 0 {
		hc.Interval = HealthCheckDefaults().Interval
	}

	if hc.Interval < 5*time.Second {
		return HealthCheckConfig{}, fmt.Errorf(
			"healthcheck: interval %v is below minimum of 5s", hc.Interval,
		)
	}

	return hc, nil
}
