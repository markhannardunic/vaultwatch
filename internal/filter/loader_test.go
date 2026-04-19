package filter

import (
	"testing"

	"github.com/vaultwatch/internal/config"
)

func makeConfig(prefixes, excludes []string) *config.Config {
	cfg := &config.Config{}
	cfg.Filter.Prefixes = prefixes
	cfg.Filter.Excludes = excludes
	return cfg
}

func TestFromConfig_NilConfig(t *testing.T) {
	_, err := FromConfig(nil)
	if err == nil {
		t.Fatal("expected error for nil config")
	}
}

func TestFromConfig_ReturnsFilter(t *testing.T) {
	cfg := makeConfig([]string{"secret/"}, []string{"secret/internal/"})
	f, err := FromConfig(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f == nil {
		t.Fatal("expected non-nil filter")
	}
}

func TestApplyToSecrets_NilFilter(t *testing.T) {
	paths := []string{"secret/foo", "secret/bar"}
	result := ApplyToSecrets(nil, paths)
	if len(result) != len(paths) {
		t.Fatalf("expected %d paths, got %d", len(paths), len(result))
	}
}

func TestApplyToSecrets_FiltersCorrectly(t *testing.T) {
	cfg := makeConfig([]string{"secret/prod/"}, nil)
	f, _ := FromConfig(cfg)
	paths := []string{"secret/prod/db", "secret/dev/db"}
	result := ApplyToSecrets(f, paths)
	if len(result) != 1 || result[0] != "secret/prod/db" {
		t.Fatalf("unexpected result: %v", result)
	}
}

func TestApplyToSecrets_EmptyPaths(t *testing.T) {
	cfg := makeConfig([]string{"secret/"}, nil)
	f, _ := FromConfig(cfg)
	result := ApplyToSecrets(f, []string{})
	if len(result) != 0 {
		t.Fatalf("expected empty result, got %v", result)
	}
}
