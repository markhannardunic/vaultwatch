package notify

import (
	"errors"
	"strings"
)

// TransformFunc is a function that transforms a message before sending.
type TransformFunc func(msg string) (string, error)

// TransformSender wraps a Sender and applies a chain of TransformFuncs
// to the message before forwarding it.
type TransformSender struct {
	sender     Sender
	transforms []TransformFunc
}

// NewTransformSender creates a TransformSender that applies each transform in
// order before delegating to the wrapped sender. At least one transform must
// be provided and sender must not be nil.
func NewTransformSender(sender Sender, transforms ...TransformFunc) (*TransformSender, error) {
	if sender == nil {
		return nil, errors.New("transform: sender must not be nil")
	}
	if len(transforms) == 0 {
		return nil, errors.New("transform: at least one transform func required")
	}
	for i, t := range transforms {
		if t == nil {
			return nil, fmt.Errorf("transform: transform at index %d is nil", i)
		}
	}
	return &TransformSender{sender: sender, transforms: transforms}, nil
}

// Send applies all registered transforms to msg in order, then forwards the
// result to the underlying sender. Returns on the first transform error.
func (t *TransformSender) Send(path, level, msg string) error {
	current := msg
	for _, fn := range t.transforms {
		var err error
		current, err = fn(current)
		if err != nil {
			return err
		}
	}
	return t.sender.Send(path, level, current)
}

// TrimSpaceTransform returns a TransformFunc that trims leading and trailing
// whitespace from the message.
func TrimSpaceTransform() TransformFunc {
	return func(msg string) (string, error) {
		return strings.TrimSpace(msg), nil
	}
}

// PrefixTransform returns a TransformFunc that prepends prefix to the message.
func PrefixTransform(prefix string) TransformFunc {
	return func(msg string) (string, error) {
		return prefix + msg, nil
	}
}

// TruncateTransform returns a TransformFunc that truncates the message to at
// most maxLen runes. If maxLen <= 0 an error is returned at construction time.
func TruncateTransform(maxLen int) (TransformFunc, error) {
	if maxLen <= 0 {
		return nil, errors.New("transform: maxLen must be positive")
	}
	return func(msg string) (string, error) {
		runes := []rune(msg)
		if len(runes) > maxLen {
			return string(runes[:maxLen]), nil
		}
		return msg, nil
	}, nil
}
