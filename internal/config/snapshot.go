package config

import "fmt"

// SnapshotConfig controls the in-memory dispatch snapshot store.
type SnapshotConfig struct {
	// MaxSize is the maximum number of dispatch records to retain.
	MaxSize int `yaml:"max_size"`
}

// SnapshotDefaults returns a SnapshotConfig with sensible defaults.
func SnapshotDefaults() SnapshotConfig {
	return SnapshotConfig{
		MaxSize: 200,
	}
}

// SnapshotConfigFromMain extracts SnapshotConfig from the top-level config map.
// It expects an optional "snapshot" key whose value is a map[string]interface{}.
func SnapshotConfigFromMain(raw map[string]interface{}) (SnapshotConfig, error) {
	cfg := SnapshotDefaults()
	if raw == nil {
		return cfg, nil
	}
	v, ok := raw["snapshot"]
	if !ok {
		return cfg, nil
	}
	m, ok := v.(map[string]interface{})
	if !ok {
		return cfg, fmt.Errorf("snapshot: expected a mapping, got %T", v)
	}
	if ms, ok := m["max_size"]; ok {
		switch val := ms.(type) {
		case int:
			if val <= 0 {
				return cfg, fmt.Errorf("snapshot.max_size must be positive")
			}
			cfg.MaxSize = val
		case float64:
			if int(val) <= 0 {
				return cfg, fmt.Errorf("snapshot.max_size must be positive")
			}
			cfg.MaxSize = int(val)
		default:
			return cfg, fmt.Errorf("snapshot.max_size: unexpected type %T", ms)
		}
	}
	return cfg, nil
}
