package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config holds the top-level vaultwatch configuration.
type Config struct {
	Vault  VaultConfig  `yaml:"vault"`
	Alert  AlertConfig  `yaml:"alert"`
}

// VaultConfig holds Vault connection settings.
type VaultConfig struct {
	Address string `yaml:"address"`
	Token   string `yaml:"token"`
	Paths   []string `yaml:"paths"`
}

// AlertConfig holds alerting thresholds and output settings.
type AlertConfig struct {
	WarnWithin    time.Duration `yaml:"warn_within"`
	CriticalWithin time.Duration `yaml:"critical_within"`
	Output        string        `yaml:"output"` // stdout | file
	OutputFile    string        `yaml:"output_file"`
}

// Load reads and parses a YAML config file from the given path.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config file: %w", err)
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &cfg, nil
}

func (c *Config) validate() error {
	if c.Vault.Address == "" {
		return fmt.Errorf("vault.address is required")
	}
	if c.Vault.Token == "" {
		return fmt.Errorf("vault.token is required")
	}
	if len(c.Vault.Paths) == 0 {
		return fmt.Errorf("vault.paths must contain at least one path")
	}
	if c.Alert.WarnWithin == 0 {
		c.Alert.WarnWithin = 7 * 24 * time.Hour
	}
	if c.Alert.CriticalWithin == 0 {
		c.Alert.CriticalWithin = 24 * time.Hour
	}
	return nil
}
