package notify

import (
	"net"
	"net/smtp"
	"strings"
	"testing"
)

func TestNewEmailClient_MissingHost(t *testing.T) {
	_, err := NewEmailClient(EmailConfig{
		From: "sender@example.com",
		To:   []string{"recipient@example.com"},
	})
	if err == nil || !strings.Contains(err.Error(), "host") {
		t.Fatalf("expected host error, got %v", err)
	}
}

func TestNewEmailClient_MissingRecipients(t *testing.T) {
	_, err := NewEmailClient(EmailConfig{
		Host: "smtp.example.com",
		From: "sender@example.com",
	})
	if err == nil || !strings.Contains(err.Error(), "recipient") {
		t.Fatalf("expected recipient error, got %v", err)
	}
}

func TestNewEmailClient_MissingFrom(t *testing.T) {
	_, err := NewEmailClient(EmailConfig{
		Host: "smtp.example.com",
		To:   []string{"recipient@example.com"},
	})
	if err == nil || !strings.Contains(err.Error(), "sender") {
		t.Fatalf("expected sender error, got %v", err)
	}
}

func TestNewEmailClient_DefaultPort(t *testing.T) {
	client, err := NewEmailClient(EmailConfig{
		Host: "smtp.example.com",
		From: "sender@example.com",
		To:   []string{"recipient@example.com"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client.port != 587 {
		t.Errorf("expected default port 587, got %d", client.port)
	}
}

func TestNewEmailClient_WithAuth(t *testing.T) {
	client, err := NewEmailClient(EmailConfig{
		Host:     "smtp.example.com",
		Port:     465,
		Username: "user",
		Password: "pass",
		From:     "sender@example.com",
		To:       []string{"recipient@example.com"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client.auth == nil {
		t.Error("expected auth to be set when username is provided")
	}
	_ = smtp.PlainAuth // ensure smtp is used
}

func TestEmailClient_Send_ConnectionRefused(t *testing.T) {
	// Use a port that is guaranteed to be unavailable.
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	port := l.Addr().(*net.TCPAddr).Port
	l.Close() // immediately close so the port is unavailable

	client := &EmailClient{
		host: "127.0.0.1",
		port: port,
		from: "sender@example.com",
		to:   []string{"recipient@example.com"},
	}

	err = client.Send("Test Subject", "Test body")
	if err == nil {
		t.Error("expected connection error, got nil")
	}
}
