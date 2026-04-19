package notify_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/vaultwatch/internal/notify"
)

func TestNewSlackClient_EmptyURL(t *testing.T) {
	_, err := notify.NewSlackClient("")
	if err == nil {
		t.Fatal("expected error for empty webhook URL")
	}
}

func TestNewSlackClient_ValidURL(t *testing.T) {
	c, err := notify.NewSlackClient("https://hooks.slack.com/test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestSlackClient_Send_Success(t *testing.T) {
	var received map[string]string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&received)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client, _ := notify.NewSlackClient(server.URL)
	if err := client.Send("test alert"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if received["text"] != "test alert" {
		t.Errorf("expected 'test alert', got %q", received["text"])
	}
}

func TestSlackClient_Send_NonOKStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client, _ := notify.NewSlackClient(server.URL)
	if err := client.Send("test"); err == nil {
		t.Fatal("expected error for non-OK status")
	}
}
