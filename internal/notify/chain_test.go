package notify

import (
	"errors"
	"testing"
)

type mockChainSender struct {
	called  bool
	returnErr error
}

func (m *mockChainSender) Send(level, path, message string) error {
	m.called = true
	return m.returnErr
}

func TestNewChain_NoSenders_ReturnsError(t *testing.T) {
	_, err := NewChain(false)
	if err == nil {
		t.Fatal("expected error for empty senders, got nil")
	}
}

func TestNewChain_NilSender_ReturnsError(t *testing.T) {
	_, err := NewChain(false, nil)
	if err == nil {
		t.Fatal("expected error for nil sender, got nil")
	}
}

func TestChain_Send_CallsAllSendersOnSuccess(t *testing.T) {
	a := &mockChainSender{}
	b := &mockChainSender{}
	chain, err := NewChain(false, a, b)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := chain.Send("warning", "secret/foo", "expiring soon"); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !a.called || !b.called {
		t.Error("expected both senders to be called")
	}
}

func TestChain_Send_StopsOnFirstError_WhenNotContinuing(t *testing.T) {
	a := &mockChainSender{returnErr: errors.New("first failed")}
	b := &mockChainSender{}
	chain, _ := NewChain(false, a, b)
	err := chain.Send("critical", "secret/bar", "expired")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if b.called {
		t.Error("expected second sender to be skipped after first error")
	}
}

func TestChain_Send_ContinuesOnError_CollectsAllErrors(t *testing.T) {
	a := &mockChainSender{returnErr: errors.New("err a")}
	b := &mockChainSender{returnErr: errors.New("err b")}
	chain, _ := NewChain(true, a, b)
	err := chain.Send("critical", "secret/baz", "expired")
	if err == nil {
		t.Fatal("expected combined error, got nil")
	}
	if !a.called || !b.called {
		t.Error("expected both senders to be called when continueOnError is true")
	}
}

func TestChain_Len_ReturnsCorrectCount(t *testing.T) {
	a := &mockChainSender{}
	b := &mockChainSender{}
	chain, _ := NewChain(false, a, b)
	if chain.Len() != 2 {
		t.Errorf("expected Len 2, got %d", chain.Len())
	}
}
