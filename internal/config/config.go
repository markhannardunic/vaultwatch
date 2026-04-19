package config

import (
	"errors"
	"os"

	"gopkg.in/yaml.v3"

	"github.com/vaultwatch/internal/filter"
)

// Thresholds holds warning/critical day thresholds.
type Thresholds struct {
	Warning  int `yaml:"warning"`
	Critical int `yaml:"critical"`
}

// Config holds the full application configuration.
type Config struct {
	VaultAddress string     `yaml:"vault_address"`
	Token        string     `yaml:"token"`
	Secrets      []string   `yaml:"secrets"`
	Thresholds   Thresholds `yaml:"thresholds"`
	Output       string     `yaml:"output"`
	Filter       filter.Rule `yaml:"filter"`
}

// Load reads and parses a YAML config file.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	if cfg.VaultAddress == "" {
		return nil, errors.New("vault_address is required")
	}
	if cfg.Thresholds.Warning == 0 {
		cfg.Thresholds.Warning = 14
	}
	if cfg.Thresholds.Critical == 0 {
		cfg.Thresholds.Critical = 7
	}
	if cfg.Output == "" {
		cfg.Output = "text"
	}
	return &cfg, nil
}
