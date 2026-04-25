package notify

import (
	"errors"
	"strings"
	"testing"
)

type recordingSender struct {
	received []string
	err      error
}

func (r *recordingSender) Send(msg string) error {
	r.received = append(r.received, msg)
	return r.err
}

func TestNewTeeSender_NoSenders_ReturnsError(t *testing.T) {
	_, err := NewTeeSender()
	if err == nil {
		t.Fatal("expected error for zero senders, got nil")
	}
}

func TestNewTeeSender_NilSender_ReturnsError(t *testing.T) {
	a := &recordingSender{}
	_, err := NewTeeSender(a, nil)
	if err == nil {
		t.Fatal("expected error for nil sender, got nil")
	}
}

func TestNewTeeSender_ValidSenders(t *testing.T) {
	a := &recordingSender{}
	b := &recordingSender{}
	tee, err := NewTeeSender(a, b)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tee.Len() != 2 {
		t.Fatalf("expected 2 senders, got %d", tee.Len())
	}
}

func TestTeeSender_Send_DeliveriesToAll(t *testing.T) {
	a := &recordingSender{}
	b := &recordingSender{}
	tee, _ := NewTeeSender(a, b)

	if err := tee.Send("hello"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(a.received) != 1 || a.received[0] != "hello" {
		t.Errorf("sender a did not receive expected message")
	}
	if len(b.received) != 1 || b.received[0] != "hello" {
		t.Errorf("sender b did not receive expected message")
	}
}

func TestTeeSender_Send_ContinuesAfterError(t *testing.T) {
	a := &recordingSender{err: errors.New("a failed")}
	b := &recordingSender{}
	tee, _ := NewTeeSender(a, b)

	_ = tee.Send("msg")

	// b must still have been called despite a failing
	if len(b.received) != 1 {
		t.Errorf("sender b should have been called even after sender a failed")
	}
}

func TestTeeSender_Send_SingleError_ReturnedDirectly(t *testing.T) {
	a := &recordingSender{err: errors.New("only error")}
	tee, _ := NewTeeSender(a)

	err := tee.Send("msg")
	if err == nil || err.Error() != "only error" {
		t.Errorf("expected original error, got %v", err)
	}
}

func TestTeeSender_Send_MultipleErrors_Combined(t *testing.T) {
	a := &recordingSender{err: errors.New("err-a")}
	b := &recordingSender{err: errors.New("err-b")}
	tee, _ := NewTeeSender(a, b)

	err := tee.Send("msg")
	if err == nil {
		t.Fatal("expected combined error, got nil")
	}
	if !strings.Contains(err.Error(), "err-a") || !strings.Contains(err.Error(), "err-b") {
		t.Errorf("expected both errors in message, got: %v", err)
	}
}
