package config

import (
	"errors"
	"time"
)

// CircuitBreakerConfig holds configuration for the circuit breaker wrapper.
type CircuitBreakerConfig struct {
	Enabled           bool          `yaml:"enabled"`
	FailureThreshold  int           `yaml:"failure_threshold"`
	RecoveryDuration  time.Duration `yaml:"recovery_duration"`
}

// CircuitBreakerDefaults returns a CircuitBreakerConfig with sensible defaults.
func CircuitBreakerDefaults() CircuitBreakerConfig {
	return CircuitBreakerConfig{
		Enabled:          true,
		FailureThreshold: 5,
		RecoveryDuration: 30 * time.Second,
	}
}

// Validate checks that the CircuitBreakerConfig values are acceptable.
func (c *CircuitBreakerConfig) Validate() error {
	if !c.Enabled {
		return nil
	}
	if c.FailureThreshold < 1 {
		return errors.New("circuit_breaker: failure_threshold must be at least 1")
	}
	if c.RecoveryDuration <= 0 {
		return errors.New("circuit_breaker: recovery_duration must be positive")
	}
	return nil
}

// CircuitBreakerConfigFromMain extracts circuit breaker config from the main
// Config, applying defaults for any zero values.
func CircuitBreakerConfigFromMain(cfg *Config) CircuitBreakerConfig {
	if cfg == nil {
		return CircuitBreakerDefaults()
	}
	out := cfg.CircuitBreaker
	if out.FailureThreshold == 0 {
		out.FailureThreshold = CircuitBreakerDefaults().FailureThreshold
	}
	if out.RecoveryDuration == 0 {
		out.RecoveryDuration = CircuitBreakerDefaults().RecoveryDuration
	}
	return out
}
