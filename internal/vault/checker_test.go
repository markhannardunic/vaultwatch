package vault

import (
	"testing"
	"time"
)

func makeTestChecker(expiresAt time.Time) (*Checker, *mockClient) {
	mc := &mockClient{expiresAt: expiresAt}
	client := &Client{raw: mc}
	return NewChecker(client, 14, 3), mc
}

func TestChecker_Check_Warning(t *testing.T) {
	expiresAt := time.Now().Add(10 * 24 * time.Hour)
	checker, _ := makeTestChecker(expiresAt)

	status, err := checker.Check("secret/my-token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !status.Warning {
		t.Errorf("expected Warning=true, got false (daysLeft=%d)", status.DaysLeft)
	}
	if status.Critical {
		t.Errorf("expected Critical=false, got true")
	}
}

func TestChecker_Check_Critical(t *testing.T) {
	expiresAt := time.Now().Add(2 * 24 * time.Hour)
	checker, _ := makeTestChecker(expiresAt)

	status, err := checker.Check("secret/db-pass")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !status.Critical {
		t.Errorf("expected Critical=true, got false (daysLeft=%d)", status.DaysLeft)
	}
}

func TestChecker_Check_Healthy(t *testing.T) {
	expiresAt := time.Now().Add(30 * 24 * time.Hour)
	checker, _ := makeTestChecker(expiresAt)

	status, err := checker.Check("secret/api-key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status.Warning || status.Critical {
		t.Errorf("expected healthy status, got warning=%v critical=%v", status.Warning, status.Critical)
	}
}

func TestFormatAlert_Levels(t *testing.T) {
	base := time.Now().Add(5 * 24 * time.Hour)
	cases := []struct {
		status   *SecretStatus
		contains string
	}{
		{&SecretStatus{Path: "p", ExpiresAt: base, DaysLeft: 5, Critical: true}, "CRITICAL"},
		{&SecretStatus{Path: "p", ExpiresAt: base, DaysLeft: 10, Warning: true}, "WARNING"},
		{&SecretStatus{Path: "p", ExpiresAt: base, DaysLeft: 30}, "INFO"},
	}
	for _, tc := range cases {
		out := FormatAlert(tc.status)
		if len(out) == 0 {
			t.Error("empty alert string")
		}
		if !contains(out, tc.contains) {
			t.Errorf("expected %q in %q", tc.contains, out)
		}
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 ||
		func() bool {
			for i := 0; i <= len(s)-len(sub); i++ {
				if s[i:i+len(sub)] == sub {
					return true
				}
			}
			return false
		}())
}
