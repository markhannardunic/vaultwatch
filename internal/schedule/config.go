package schedule

import (
	"errors"
	"time"
)

// Config holds scheduling configuration parsed from the user config.
type Config struct {
	// Interval is how often the audit job runs.
	Interval time.Duration
}

// DefaultInterval is used when no interval is specified.
const DefaultInterval = 5 * time.Minute

// Parse validates and returns a Config, applying defaults where needed.
func Parse(intervalStr string) (Config, error) {
	if intervalStr == "" {
		return Config{Interval: DefaultInterval}, nil
	}

	d, err := time.ParseDuration(intervalStr)
	if err != nil {
		return Config{}, errors.New("schedule: invalid interval: " + err.Error())
	}
	if d < time.Second {
		return Config{}, errors.New("schedule: interval must be at least 1s")
	}

	return Config{Interval: d}, nil
}
