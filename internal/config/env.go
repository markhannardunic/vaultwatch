package config

import (
	"os"
)

// ApplyEnvOverrides overrides config values with environment variables
// when present, allowing secrets to stay out of config files.
func ApplyEnvOverrides(cfg *Config) {
	if v := os.Getenv("VAULT_ADDR"); v != "" {
		cfg.Vault.Address = v
	}
	if v := os.Getenv("VAULT_TOKEN"); v != "" {
		cfg.Vault.Token = v
	}
	if v := os.Getenv("VAULTWATCH_OUTPUT"); v != "" {
		cfg.Alert.Output = v
	}
	if v := os.Getenv("VAULTWATCH_OUTPUT_FILE"); v != "" {
		cfg.Alert.OutputFile = v
	}
}
