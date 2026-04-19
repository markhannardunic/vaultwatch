package schedule

import (
	"context"
	"time"
)

// Scheduler runs a job on a fixed interval until the context is cancelled.
type Scheduler struct {
	interval time.Duration
	job      func(ctx context.Context) error
	onError  func(err error)
}

// New creates a Scheduler with the given interval and job function.
func New(interval time.Duration, job func(ctx context.Context) error, onError func(err error)) *Scheduler {
	if onError == nil {
		onError = func(err error) {}
	}
	return &Scheduler{
		interval: interval,
		job:      job,
		onError:  onError,
	}
}

// Run executes the job immediately and then on every interval tick.
// It blocks until ctx is cancelled.
func (s *Scheduler) Run(ctx context.Context) {
	if err := s.job(ctx); err != nil {
		s.onError(err)
	}

	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := s.job(ctx); err != nil {
				s.onError(err)
			}
		case <-ctx.Done():
			return
		}
	}
}
