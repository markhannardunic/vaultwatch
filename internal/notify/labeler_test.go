package notify

import (
	"errors"
	"strings"
	"testing"
)

func TestNewLabeler_NilSender(t *testing.T) {
	_, err := NewLabeler(nil, Labels{"env": "prod"})
	if err == nil {
		t.Fatal("expected error for nil sender")
	}
}

func TestNewLabeler_EmptyLabels(t *testing.T) {
	_, err := NewLabeler(&captureSender{}, Labels{})
	if err == nil {
		t.Fatal("expected error for empty labels")
	}
}

func TestNewLabeler_ValidConfig(t *testing.T) {
	l, err := NewLabeler(&captureSender{}, Labels{"env": "prod"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if l == nil {
		t.Fatal("expected non-nil labeler")
	}
}

func TestLabeler_Send_EnrichesMessage(t *testing.T) {
	cap := &captureSender{}
	l, err := NewLabeler(cap, Labels{"env": "staging"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := l.Send("secret expiring soon"); err != nil {
		t.Fatalf("unexpected send error: %v", err)
	}

	if !strings.Contains(cap.last, "env=staging") {
		t.Errorf("expected label in message, got: %s", cap.last)
	}
	if !strings.Contains(cap.last, "secret expiring soon") {
		t.Errorf("expected original message body, got: %s", cap.last)
	}
}

func TestLabeler_Send_PropagatesError(t *testing.T) {
	sender := &errorSender{err: errors.New("downstream failure")}
	l, _ := NewLabeler(sender, Labels{"team": "platform"})

	err := l.Send("test")
	if err == nil || !strings.Contains(err.Error(), "downstream failure") {
		t.Errorf("expected downstream error, got: %v", err)
	}
}

func TestLabeler_Labels_ReturnsCopy(t *testing.T) {
	original := Labels{"region": "us-east-1"}
	l, _ := NewLabeler(&captureSender{}, original)

	copy := l.Labels()
	copy["injected"] = "yes"

	if _, ok := l.Labels()["injected"]; ok {
		t.Error("modifying returned labels should not affect internal state")
	}
}

func TestLabeler_Send_MultipleLabels(t *testing.T) {
	cap := &captureSender{}
	l, err := NewLabeler(cap, Labels{"env": "prod", "service": "vaultwatch"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := l.Send("check"); err != nil {
		t.Fatalf("send error: %v", err)
	}

	if !strings.Contains(cap.last, "env=prod") || !strings.Contains(cap.last, "service=vaultwatch") {
		t.Errorf("expected both labels in output, got: %s", cap.last)
	}
}
