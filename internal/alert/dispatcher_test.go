package alert

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/user/vaultwatch/internal/vault"
)

func makeChecker(ttls map[string]time.Duration) *vault.Checker {
	// vault.NewChecker accepts a client; we use a stub via a pre-built Checker.
	// For unit tests we rely on vault.Checker exposing a testable seam.
	// Here we build a minimal fake using the exported constructor.
	return vault.NewCheckerWithStub(ttls)
}

func TestDispatcher_Run_SendsAlertForExpiringSecret(t *testing.T) {
	ttls := map[string]time.Duration{
		"secret/db": 12 * time.Hour,
	}
	checker := makeChecker(ttls)

	var buf bytes.Buffer
	notifier := NewNotifier(&buf)
	d := NewDispatcher(checker, notifier, []string{"secret/db"})

	if err := d.Run(); err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "[CRITICAL]") {
		t.Errorf("expected CRITICAL alert, got: %s", output)
	}
}

func TestDispatcher_Run_SkipsHealthySecret(t *testing.T) {
	ttls := map[string]time.Duration{
		"secret/healthy": 720 * time.Hour,
	}
	checker := makeChecker(ttls)

	var buf bytes.Buffer
	notifier := NewNotifier(&buf)
	d := NewDispatcher(checker, notifier, []string{"secret/healthy"})

	if err := d.Run(); err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	if buf.Len() != 0 {
		t.Errorf("expected no output for healthy secret, got: %s", buf.String())
	}
}
