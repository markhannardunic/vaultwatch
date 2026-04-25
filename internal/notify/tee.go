package notify

import "fmt"

// TeeSender fans out a single Send call to multiple senders, collecting all
// errors. Unlike FallbackRouter, TeeSender always calls every sender regardless
// of prior failures.
type TeeSender struct {
	senders []Sender
}

// NewTeeSender returns a TeeSender that broadcasts every message to all
// provided senders. At least one sender must be supplied.
func NewTeeSender(senders ...Sender) (*TeeSender, error) {
	if len(senders) == 0 {
		return nil, fmt.Errorf("tee: at least one sender is required")
	}
	for i, s := range senders {
		if s == nil {
			return nil, fmt.Errorf("tee: sender at index %d is nil", i)
		}
	}
	return &TeeSender{senders: senders}, nil
}

// Send delivers msg to every registered sender. All senders are called even
// when earlier ones fail. If multiple senders fail the errors are combined into
// a single joined error.
func (t *TeeSender) Send(msg string) error {
	var errs []error
	for _, s := range t.senders {
		if err := s.Send(msg); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) == 0 {
		return nil
	}
	if len(errs) == 1 {
		return errs[0]
	}
	combined := errs[0].Error()
	for _, e := range errs[1:] {
		combined += "; " + e.Error()
	}
	return fmt.Errorf("tee: multiple send errors: %s", combined)
}

// Len returns the number of senders registered with this TeeSender.
func (t *TeeSender) Len() int {
	return len(t.senders)
}
