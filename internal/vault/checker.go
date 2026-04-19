package vault

import (
	"fmt"
	"time"
)

// SecretStatus represents the expiry status of a secret.
type SecretStatus struct {
	Path      string
	ExpiresAt time.Time
	DaysLeft  int
	Critical  bool
	Warning   bool
}

// Checker evaluates secrets against configured thresholds.
type Checker struct {
	client           *Client
	WarnThresholdDays int
	CritThresholdDays int
}

// NewChecker creates a Checker with the given client and thresholds.
func NewChecker(client *Client, warnDays, critDays int) *Checker {
	return &Checker{
		client:           client,
		WarnThresholdDays: warnDays,
		CritThresholdDays: critDays,
	}
}

// Check retrieves metadata for the secret at path and evaluates its expiry.
func (c *Checker) Check(path string) (*SecretStatus, error) {
	meta, err := c.client.GetSecretMeta(path)
	if err != nil {
		return nil, fmt.Errorf("checker: %w", err)
	}

	daysLeft := int(time.Until(meta.ExpiresAt).Hours() / 24)
	status := &SecretStatus{
		Path:      path,
		ExpiresAt: meta.ExpiresAt,
		DaysLeft:  daysLeft,
		Critical:  daysLeft <= c.CritThresholdDays,
		Warning:   daysLeft <= c.WarnThresholdDays && daysLeft > c.CritThresholdDays,
	}
	return status, nil
}

// FormatAlert returns a human-readable alert string for the given status.
func FormatAlert(s *SecretStatus) string {
	level := "INFO"
	if s.Critical {
		level = "CRITICAL"
	} else if s.Warning {
		level = "WARNING"
	}
	return fmt.Sprintf("[%s] secret %q expires in %d day(s) (at %s)",
		level, s.Path, s.DaysLeft, s.ExpiresAt.Format(time.RFC3339))
}
