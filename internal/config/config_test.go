package config_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/vaultwatch/internal/config"
)

func writeTemp(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(p, []byte(content), 0600); err != nil {
		t.Fatalf("writing temp config: %v", err)
	}
	return p
}

func TestLoad_ValidConfig(t *testing.T) {
	path := writeTemp(t, `
vault:
  address: "http://127.0.0.1:8200"
  token: "root"
  paths:
    - "secret/myapp"
alert:
  warn_within: 168h
  critical_within: 24h
`)
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Vault.Address != "http://127.0.0.1:8200" {
		t.Errorf("unexpected address: %s", cfg.Vault.Address)
	}
	if cfg.Alert.WarnWithin != 168*time.Hour {
		t.Errorf("unexpected warn_within: %v", cfg.Alert.WarnWithin)
	}
}

func TestLoad_DefaultThresholds(t *testing.T) {
	path := writeTemp(t, `
vault:
  address: "http://127.0.0.1:8200"
  token: "root"
  paths:
    - "secret/app"
`)
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Alert.WarnWithin != 7*24*time.Hour {
		t.Errorf("expected default warn_within, got %v", cfg.Alert.WarnWithin)
	}
	if cfg.Alert.CriticalWithin != 24*time.Hour {
		t.Errorf("expected default critical_within, got %v", cfg.Alert.CriticalWithin)
	}
}

func TestLoad_MissingAddress(t *testing.T) {
	path := writeTemp(t, `
vault:
  token: "root"
  paths:
    - "secret/app"
`)
	_, err := config.Load(path)
	if err == nil {
		t.Fatal("expected error for missing vault.address")
	}
}

func TestLoad_FileNotFound(t *testing.T) {
	_, err := config.Load("/nonexistent/path/config.yaml")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}
