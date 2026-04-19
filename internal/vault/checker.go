package vault

import (
	"fmt"
	"time"
)

// ExpiryAlert represents a secret that is expiring soon.
type ExpiryAlert struct {
	Path      string
	ExpiresAt time.Time
	TTL       time.Duration
}

// Checker audits a list of secret paths and returns alerts for secrets
// expiring within the given threshold duration.
type Checker struct {
	client    *Client
	threshold time.Duration
}

// NewChecker creates a Checker with the given Vault client and alert threshold.
func NewChecker(client *Client, threshold time.Duration) *Checker {
	return &Checker{client: client, threshold: threshold}
}

// Audit checks each path and returns alerts for secrets expiring within the threshold.
func (ch *Checker) Audit(paths []string) ([]ExpiryAlert, error) {
	var alerts []ExpiryAlert

	for _, path := range paths {
		meta, err := ch.client.GetSecretMeta(path)
		if err != nil {
			return nil, fmt.Errorf("audit failed for path %s: %w", path, err)
		}

		if meta.TTL <= ch.threshold {
			alerts = append(alerts, ExpiryAlert{
				Path:      meta.Path,
				ExpiresAt: meta.ExpiresAt,
				TTL:       meta.TTL,
			})
		}
	}

	return alerts, nil
}

// FormatAlert returns a human-readable string for an ExpiryAlert.
func FormatAlert(a ExpiryAlert) string {
	return fmt.Sprintf("[EXPIRING] %s — TTL: %s, Expires At: %s",
		a.Path,
		a.TTL.Round(time.Second),
		a.ExpiresAt.Format(time.RFC3339),
	)
}
