package notify

import (
	"fmt"
	"strings"
)

// Sender is the interface for dispatching alert messages.
type FallbackSender interface {
	Send(level, path, message string) error
}

// FallbackRouter attempts to send via a primary sender and falls back to one
// or more secondary senders if the primary fails.
type FallbackRouter struct {
	primary   FallbackSender
	fallbacks []FallbackSender
}

// NewFallbackRouter creates a FallbackRouter with a primary sender and at least
// one fallback. Returns an error if either argument is empty.
func NewFallbackRouter(primary FallbackSender, fallbacks ...FallbackSender) (*FallbackRouter, error) {
	if primary == nil {
		return nil, fmt.Errorf("fallback router: primary sender must not be nil")
	}
	if len(fallbacks) == 0 {
		return nil, fmt.Errorf("fallback router: at least one fallback sender is required")
	}
	for i, f := range fallbacks {
		if f == nil {
			return nil, fmt.Errorf("fallback router: fallback sender at index %d is nil", i)
		}
	}
	return &FallbackRouter{
		primary:   primary,
		fallbacks: fallbacks,
	}, nil
}

// Dispatch sends the alert via the primary sender. If the primary fails, it
// attempts each fallback in order. Returns an error only if all senders fail,
// combining all errors into a single message.
func (r *FallbackRouter) Dispatch(level, path, message string) error {
	if err := r.primary.Send(level, path, message); err == nil {
		return nil
	}

	var errs []string
	for _, fb := range r.fallbacks {
		if err := fb.Send(level, path, message); err == nil {
			return nil
		} else {
			errs = append(errs, err.Error())
		}
	}

	return fmt.Errorf("fallback router: all senders failed: %s", strings.Join(errs, "; "))
}
