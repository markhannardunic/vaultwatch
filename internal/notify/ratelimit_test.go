package notify

import (
	"testing"
	"time"
)

func TestNewRateLimiter_InvalidMax(t *testing.T) {
	_, err := NewRateLimiter(0, time.Minute)
	if err == nil {
		t.Fatal("expected error for max=0")
	}
}

func TestNewRateLimiter_InvalidWindow(t *testing.T) {
	_, err := NewRateLimiter(5, 0)
	if err == nil {
		t.Fatal("expected error for window=0")
	}
}

func TestRateLimiter_Allow_WithinLimit(t *testing.T) {
	rl, _ := NewRateLimiter(3, time.Minute)
	for i := 0; i < 3; i++ {
		if !rl.Allow("secret/foo") {
			t.Fatalf("expected allow on call %d", i+1)
		}
	}
}

func TestRateLimiter_Allow_ExceedsLimit(t *testing.T) {
	rl, _ := NewRateLimiter(2, time.Minute)
	rl.Allow("secret/foo")
	rl.Allow("secret/foo")
	if rl.Allow("secret/foo") {
		t.Fatal("expected deny after exceeding limit")
	}
}

func TestRateLimiter_Allow_WindowExpiry_Resets(t *testing.T) {
	now := time.Now()
	rl, _ := NewRateLimiter(2, 5*time.Second)
	rl.nowFunc = func() time.Time { return now }
	rl.Allow("k")
	rl.Allow("k")

	// advance past window
	rl.nowFunc = func() time.Time { return now.Add(6 * time.Second) }
	if !rl.Allow("k") {
		t.Fatal("expected allow after window expiry")
	}
}

func TestRateLimiter_Allow_DifferentKeys_Independent(t *testing.T) {
	rl, _ := NewRateLimiter(1, time.Minute)
	rl.Allow("a")
	if !rl.Allow("b") {
		t.Fatal("different key should be independent")
	}
}

func TestRateLimiter_Reset_AllowsImmediateResend(t *testing.T) {
	rl, _ := NewRateLimiter(1, time.Minute)
	rl.Allow("x")
	rl.Reset("x")
	if !rl.Allow("x") {
		t.Fatal("expected allow after reset")
	}
}
