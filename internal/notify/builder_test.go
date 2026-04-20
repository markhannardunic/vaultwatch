package notify

import (
	"bytes"
	"testing"

	"github.com/youorg/vaultwatch/internal/config"
)

func TestBuildRouter_NilConfig_UsesWriter(t *testing.T) {
	var buf bytes.Buffer
	router, err := BuildRouter(nil, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if router == nil {
		t.Fatal("expected non-nil router")
	}
}

func TestBuildRouter_EmptyNotify_UsesWriter(t *testing.T) {
	var buf bytes.Buffer
	cfg := &config.Config{}
	router, err := BuildRouter(cfg, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if router == nil {
		t.Fatal("expected non-nil router")
	}
}

func TestBuildRouter_UnknownType_ReturnsError(t *testing.T) {
	var buf bytes.Buffer
	cfg := &config.Config{
		Notify: []config.NotifyEntry{
			{Type: "carrier-pigeon", Level: "critical"},
		},
	}
	_, err := BuildRouter(cfg, &buf)
	if err == nil {
		t.Fatal("expected error for unknown notify type")
	}
}

func TestBuildRouter_SlackEntry(t *testing.T) {
	var buf bytes.Buffer
	cfg := &config.Config{
		Notify: []config.NotifyEntry{
			{Type: "slack", WebhookURL: "https://hooks.slack.com/test", Level: "warning"},
		},
	}
	router, err := BuildRouter(cfg, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if router == nil {
		t.Fatal("expected non-nil router")
	}
}

func TestBuildRouter_TeamsEntry(t *testing.T) {
	var buf bytes.Buffer
	cfg := &config.Config{
		Notify: []config.NotifyEntry{
			{Type: "teams", WebhookURL: "https://outlook.office.com/webhook/x", Level: "critical"},
		},
	}
	router, err := BuildRouter(cfg, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if router == nil {
		t.Fatal("expected non-nil router")
	}
}
