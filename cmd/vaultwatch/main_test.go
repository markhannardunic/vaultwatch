package main

import (
	"os"
	"path/filepath"
	"testing"
)

func writeTempConfig(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, "vaultwatch.yaml")
	if err := os.WriteFile(p, []byte(content), 0644); err != nil {
		t.Fatalf("write temp config: %v", err)
	}
	return p
}

func TestRun_MissingConfig(t *testing.T) {
	t.Setenv("VAULTWATCH_CONFIG", "/nonexistent/path/vaultwatch.yaml")
	if err := run(); err == nil {
		t.Fatal("expected error for missing config, got nil")
	}
}

func TestRun_InvalidVaultAddress(t *testing.T) {
	cfg := `
vault:
  address: "://bad-address"
  token: "root"
thresholds:
  warning: 72
  critical: 24
secrets: []
`
	p := writeTempConfig(t, cfg)
	t.Setenv("VAULTWATCH_CONFIG", p)

	if err := run(); err == nil {
		t.Fatal("expected error for invalid vault address, got nil")
	}
}

func TestRun_NoSecrets(t *testing.T) {
	cfg := `
vault:
  address: "http://127.0.0.1:8200"
  token: "root"
thresholds:
  warning: 72
  critical: 24
secrets: []
`
	p := writeTempConfig(t, cfg)
	t.Setenv("VAULTWATCH_CONFIG", p)

	// With no secrets the dispatcher should complete without error
	// even if vault is unreachable (nothing to check).
	if err := run(); err != nil {
		t.Fatalf("unexpected error with empty secrets list: %v", err)
	}
}
