package filter

import (
	"fmt"

	"github.com/vaultwatch/internal/config"
)

// Rules holds include/exclude configuration for secret path filtering.
type Rules struct {
	Prefixes []string
	Excludes []string
}

// FromConfig builds a Filter from the application config.
func FromConfig(cfg *config.Config) (*Filter, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config must not be nil")
	}
	f := New(cfg.Filter.Prefixes, cfg.Filter.Excludes)
	return f, nil
}

// ApplyToSecrets filters a list of secret paths using the given Filter.
// Returns only the paths that pass the filter.
func ApplyToSecrets(f *Filter, paths []string) []string {
	if f == nil || len(paths) == 0 {
		return paths
	}
	return f.Apply(paths)
}
