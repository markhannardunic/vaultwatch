package notify

import (
	"testing"
	"time"
)

func TestThrottler_Allow_FirstCallPermitted(t *testing.T) {
	th := NewThrottler(5 * time.Minute)
	key := ThrottleKey{Target: "slack", Level: "critical"}

	if !th.Allow(key) {
		t.Fatal("expected first call to be allowed")
	}
}

func TestThrottler_Allow_SecondCallBlocked(t *testing.T) {
	th := NewThrottler(5 * time.Minute)
	key := ThrottleKey{Target: "slack", Level: "critical"}

	th.Allow(key)
	if th.Allow(key) {
		t.Fatal("expected second call within cooldown to be blocked")
	}
}

func TestThrottler_Allow_AfterCooldown(t *testing.T) {
	now := time.Now()
	th := NewThrottler(5 * time.Minute)
	th.now = func() time.Time { return now }

	key := ThrottleKey{Target: "email", Level: "warning"}
	th.Allow(key)

	// Advance time past cooldown
	th.now = func() time.Time { return now.Add(6 * time.Minute) }
	if !th.Allow(key) {
		t.Fatal("expected call after cooldown to be allowed")
	}
}

func TestThrottler_Allow_DifferentKeysIndependent(t *testing.T) {
	th := NewThrottler(5 * time.Minute)
	keyA := ThrottleKey{Target: "slack", Level: "critical"}
	keyB := ThrottleKey{Target: "slack", Level: "warning"}

	th.Allow(keyA)
	if !th.Allow(keyB) {
		t.Fatal("expected different key to be allowed independently")
	}
}

func TestThrottler_Reset_AllowsImmediateResend(t *testing.T) {
	th := NewThrottler(5 * time.Minute)
	key := ThrottleKey{Target: "teams", Level: "critical"}

	th.Allow(key)
	th.Reset(key)

	if !th.Allow(key) {
		t.Fatal("expected allow after reset")
	}
}

func TestThrottler_ResetAll_ClearsState(t *testing.T) {
	th := NewThrottler(5 * time.Minute)
	keys := []ThrottleKey{
		{Target: "slack", Level: "critical"},
		{Target: "email", Level: "warning"},
	}
	for _, k := range keys {
		th.Allow(k)
	}
	th.ResetAll()
	for _, k := range keys {
		if !th.Allow(k) {
			t.Fatalf("expected key %v to be allowed after ResetAll", k)
		}
	}
}
