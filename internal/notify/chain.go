package notify

import (
	"errors"
	"fmt"
)

// Sender is the interface for sending alert messages.
type ChainSender interface {
	Send(level, path, message string) error
}

// Chain executes a list of senders in order, stopping on the first error
// unless ContinueOnError is set to true.
type Chain struct {
	senders         []ChainSender
	continueOnError bool
}

// NewChain creates a new Chain with the given senders.
// If continueOnError is true, all senders are called even if one fails;
// all errors are collected and returned as a combined error.
func NewChain(continueOnError bool, senders ...ChainSender) (*Chain, error) {
	if len(senders) == 0 {
		return nil, errors.New("chain: at least one sender is required")
	}
	for i, s := range senders {
		if s == nil {
			return nil, fmt.Errorf("chain: sender at index %d is nil", i)
		}
	}
	return &Chain{
		senders:         senders,
		continueOnError: continueOnError,
	}, nil
}

// Send calls each sender in order.
// If continueOnError is false, it stops and returns the first error.
// If continueOnError is true, it collects all errors and returns a joined error.
func (c *Chain) Send(level, path, message string) error {
	var errs []error
	for _, s := range c.senders {
		if err := s.Send(level, path, message); err != nil {
			if !c.continueOnError {
				return err
			}
			errs = append(errs, err)
		}
	}
	if len(errs) == 0 {
		return nil
	}
	return errors.Join(errs...)
}

// Len returns the number of senders in the chain.
func (c *Chain) Len() int {
	return len(c.senders)
}
