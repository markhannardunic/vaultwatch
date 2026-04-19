package main

import (
	"fmt"
	"os"

	"github.com/vaultwatch/internal/alert"
	"github.com/vaultwatch/internal/config"
	"github.com/vaultwatch/internal/vault"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	cfgPath := "vaultwatch.yaml"
	if v := os.Getenv("VAULTWATCH_CONFIG"); v != "" {
		cfgPath = v
	}

	cfg, err := config.Load(cfgPath)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	config.ApplyEnvOverrides(cfg)

	client, err := vault.NewClient(cfg.Vault.Address, cfg.Vault.Token)
	if err != nil {
		return fmt.Errorf("creating vault client: %w", err)
	}

	checker := vault.NewChecker(client, cfg.Thresholds.Warning, cfg.Thresholds.Critical)

	notifier, err := alert.NewNotifier(cfg.Alert.Output)
	if err != nil {
		return fmt.Errorf("creating notifier: %w", err)
	}
	defer notifier.Close()

	dispatcher := alert.NewDispatcher(checker, notifier, cfg.Secrets)
	if err := dispatcher.Run(); err != nil {
		return fmt.Errorf("dispatcher run: %w", err)
	}

	return nil
}
