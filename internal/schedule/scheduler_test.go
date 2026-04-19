package schedule_test

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/vaultwatch/internal/schedule"
)

func TestScheduler_RunsJobImmediately(t *testing.T) {
	var count int32
	job := func(ctx context.Context) error {
		atomic.AddInt32(&count, 1)
		return nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	s := schedule.New(10*time.Second, job, nil)

	done := make(chan struct{})
	go func() {
		s.Run(ctx)
		close(done)
	}()

	time.Sleep(50 * time.Millisecond)
	cancel()
	<-done

	if atomic.LoadInt32(&count) < 1 {
		t.Error("expected job to run at least once immediately")
	}
}

func TestScheduler_RunsOnInterval(t *testing.T) {
	var count int32
	job := func(ctx context.Context) error {
		atomic.AddInt32(&count, 1)
		return nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	s := schedule.New(30*time.Millisecond, job, nil)

	done := make(chan struct{})
	go func() {
		s.Run(ctx)
		close(done)
	}()

	time.Sleep(120 * time.Millisecond)
	cancel()
	<-done

	if atomic.LoadInt32(&count) < 3 {
		t.Errorf("expected at least 3 runs, got %d", atomic.LoadInt32(&count))
	}
}

func TestScheduler_CallsOnError(t *testing.T) {
	var errCount int32
	job := func(ctx context.Context) error {
		return errors.New("job failed")
	}
	onError := func(err error) {
		atomic.AddInt32(&errCount, 1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	s := schedule.New(10*time.Second, job, onError)

	done := make(chan struct{})
	go func() {
		s.Run(ctx)
		close(done)
	}()

	time.Sleep(50 * time.Millisecond)
	cancel()
	<-done

	if atomic.LoadInt32(&errCount) < 1 {
		t.Error("expected onError to be called at least once")
	}
}
