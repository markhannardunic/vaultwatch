package alert

import (
	"time"

	"github.com/user/vaultwatch/internal/vault"
)

// Dispatcher coordinates checking secrets and dispatching alerts.
type Dispatcher struct {
	checker  *vault.Checker
	notifier *Notifier
	paths    []string
}

// NewDispatcher creates a Dispatcher for the given secret paths.
func NewDispatcher(checker *vault.Checker, notifier *Notifier, paths []string) *Dispatcher {
	return &Dispatcher{
		checker:  checker,
		notifier: notifier,
		paths:    paths,
	}
}

// Run iterates over all configured paths, checks expiry, and sends alerts
// for secrets expiring within the info threshold (7 days).
func (d *Dispatcher) Run() error {
	const infoThreshold = 168 * time.Hour

	for _, p := range d.paths {
		result, err := d.checker.Check(p)
		if err != nil {
			return err
		}
		if result.TTL > infoThreshold {
			continue
		}
		a := Alert{
			Path:      p,
			ExpiresAt: time.Now().Add(result.TTL),
			TTL:       result.TTL,
			Severity:  d.notifier.Classify(result.TTL),
		}
		if err := d.notifier.Send(a); err != nil {
			return err
		}
	}
	return nil
}
