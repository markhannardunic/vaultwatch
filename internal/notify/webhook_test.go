package notify

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewWebhookClient_EmptyURL(t *testing.T) {
	_, err := NewWebhookClient("")
	if err == nil {
		t.Fatal("expected error for empty URL")
	}
}

func TestNewWebhookClient_ValidURL(t *testing.T) {
	c, err := NewWebhookClient("https://example.com/hook")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestWebhookClient_Send_Success(t *testing.T) {
	var received map[string]string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("unexpected content-type: %s", ct)
		}
		json.NewDecoder(r.Body).Decode(&received)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	c, _ := NewWebhookClient(ts.URL)
	if err := c.Send("critical", "cert expires soon", "secret/db/cert"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if received["level"] != "critical" {
		t.Errorf("expected level=critical, got %s", received["level"])
	}
	if received["path"] != "secret/db/cert" {
		t.Errorf("expected path=secret/db/cert, got %s", received["path"])
	}
	if received["sent_at"] == "" {
		t.Error("expected sent_at to be set")
	}
}

func TestWebhookClient_Send_NonOKStatus(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	c, _ := NewWebhookClient(ts.URL)
	err := c.Send("warning", "expiring", "secret/token")
	if err == nil {
		t.Fatal("expected error for non-2xx status")
	}
}

func TestWebhookClient_Send_BadURL(t *testing.T) {
	c, _ := NewWebhookClient("http://127.0.0.1:0/nope")
	err := c.Send("warning", "msg", "path")
	if err == nil {
		t.Fatal("expected error for unreachable server")
	}
}
