package notify

import (
	"fmt"
	"log"
	"time"
)

// Sender is the interface for dispatching alert messages.
type MiddlewareSender interface {
	Send(level, path, message string) error
}

// LoggingMiddleware wraps a Sender and logs each dispatch attempt.
type LoggingMiddleware struct {
	inner  MiddlewareSender
	logger *log.Logger
}

// NewLoggingMiddleware returns a Sender that logs before and after each Send.
func NewLoggingMiddleware(inner MiddlewareSender, logger *log.Logger) (*LoggingMiddleware, error) {
	if inner == nil {
		return nil, fmt.Errorf("notify: logging middleware requires a non-nil sender")
	}
	if logger == nil {
		return nil, fmt.Errorf("notify: logging middleware requires a non-nil logger")
	}
	return &LoggingMiddleware{inner: inner, logger: logger}, nil
}

// Send logs the dispatch attempt, delegates to the inner sender, and logs the outcome.
func (m *LoggingMiddleware) Send(level, path, message string) error {
	start := time.Now()
	m.logger.Printf("[notify] dispatching level=%s path=%s", level, path)
	err := m.inner.Send(level, path, message)
	elapsed := time.Since(start)
	if err != nil {
		m.logger.Printf("[notify] dispatch failed level=%s path=%s elapsed=%s err=%v", level, path, elapsed, err)
		return err
	}
	m.logger.Printf("[notify] dispatch succeeded level=%s path=%s elapsed=%s", level, path, elapsed)
	return nil
}

// TimingMiddleware wraps a Sender and records per-call latency via a callback.
type TimingMiddleware struct {
	inner    MiddlewareSender
	recorder func(level string, d time.Duration, err error)
}

// NewTimingMiddleware returns a Sender that calls recorder with the elapsed duration after each Send.
func NewTimingMiddleware(inner MiddlewareSender, recorder func(level string, d time.Duration, err error)) (*TimingMiddleware, error) {
	if inner == nil {
		return nil, fmt.Errorf("notify: timing middleware requires a non-nil sender")
	}
	if recorder == nil {
		return nil, fmt.Errorf("notify: timing middleware requires a non-nil recorder")
	}
	return &TimingMiddleware{inner: inner, recorder: recorder}, nil
}

// Send delegates to the inner sender and records the call latency.
func (m *TimingMiddleware) Send(level, path, message string) error {
	start := time.Now()
	err := m.inner.Send(level, path, message)
	m.recorder(level, time.Since(start), err)
	return err
}
