package notify

import (
	"errors"
	"strings"
	"testing"
)

type stubSender struct {
	called bool
	lastMsg string
}

func (s *stubSender) Send(_, _, msg string) error {
	s.called = true
	s.lastMsg = msg
	return nil
}

func TestNewTransformSender_NilSender(t *testing.T) {
	_, err := NewTransformSender(nil, TrimSpaceTransform())
	if err == nil {
		t.Fatal("expected error for nil sender")
	}
}

func TestNewTransformSender_NoTransforms(t *testing.T) {
	_, err := NewTransformSender(&stubSender{})
	if err == nil {
		t.Fatal("expected error when no transforms provided")
	}
}

func TestNewTransformSender_NilTransform(t *testing.T) {
	_, err := NewTransformSender(&stubSender{}, nil)
	if err == nil {
		t.Fatal("expected error for nil transform func")
	}
}

func TestTransformSender_AppliesTrimSpace(t *testing.T) {
	stub := &stubSender{}
	ts, err := NewTransformSender(stub, TrimSpaceTransform())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := ts.Send("sec/foo", "warning", "  hello  "); err != nil {
		t.Fatalf("Send error: %v", err)
	}
	if stub.lastMsg != "hello" {
		t.Errorf("expected 'hello', got %q", stub.lastMsg)
	}
}

func TestTransformSender_AppliesPrefix(t *testing.T) {
	stub := &stubSender{}
	ts, err := NewTransformSender(stub, PrefixTransform("[ALERT] "))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = ts.Send("sec/foo", "critical", "cert expiring")
	if !strings.HasPrefix(stub.lastMsg, "[ALERT] ") {
		t.Errorf("expected prefix, got %q", stub.lastMsg)
	}
}

func TestTransformSender_ChainedTransforms(t *testing.T) {
	stub := &stubSender{}
	trunc, _ := TruncateTransform(5)
	ts, err := NewTransformSender(stub, TrimSpaceTransform(), trunc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = ts.Send("sec/foo", "warning", "  hello world  ")
	if stub.lastMsg != "hello" {
		t.Errorf("expected 'hello', got %q", stub.lastMsg)
	}
}

func TestTransformSender_TransformError_StopsChain(t *testing.T) {
	stub := &stubSender{}
	failing := TransformFunc(func(msg string) (string, error) {
		return "", errors.New("transform failed")
	})
	ts, err := NewTransformSender(stub, failing, TrimSpaceTransform())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := ts.Send("sec/foo", "warning", "msg"); err == nil {
		t.Fatal("expected error from failing transform")
	}
	if stub.called {
		t.Error("underlying sender should not be called after transform error")
	}
}

func TestTruncateTransform_InvalidMaxLen(t *testing.T) {
	_, err := TruncateTransform(0)
	if err == nil {
		t.Fatal("expected error for maxLen=0")
	}
}
