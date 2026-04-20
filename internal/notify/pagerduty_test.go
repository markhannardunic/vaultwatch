package notify

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewPagerDutyClient_EmptyKey(t *testing.T) {
	_, err := NewPagerDutyClient("")
	if err == nil {
		t.Fatal("expected error for empty integration key")
	}
}

func TestNewPagerDutyClient_ValidKey(t *testing.T) {
	c, err := NewPagerDutyClient("test-key-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestPagerDutyClient_Send_Success(t *testing.T) {
	var received pdPayload
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Errorf("decode body: %v", err)
		}
		w.WriteHeader(http.StatusAccepted)
	}))
	defer server.Close()

	c, _ := NewPagerDutyClient("key-abc")
	c.eventURL = server.URL

	if err := c.Send("critical", "cert expires in 1h"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if received.RoutingKey != "key-abc" {
		t.Errorf("expected routing key 'key-abc', got %q", received.RoutingKey)
	}
	if received.Payload.Severity != "critical" {
		t.Errorf("expected severity 'critical', got %q", received.Payload.Severity)
	}
	if received.Payload.Source != "vaultwatch" {
		t.Errorf("expected source 'vaultwatch', got %q", received.Payload.Source)
	}
}

func TestPagerDutyClient_Send_NonOKStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	c, _ := NewPagerDutyClient("bad-key")
	c.eventURL = server.URL

	if err := c.Send("warning", "expiring soon"); err == nil {
		t.Fatal("expected error for non-2xx status")
	}
}

func TestMapSeverity(t *testing.T) {
	cases := []struct{ in, want string }{
		{"critical", "critical"},
		{"warning", "warning"},
		{"healthy", "info"},
		{"", "info"},
	}
	for _, tc := range cases {
		if got := mapSeverity(tc.in); got != tc.want {
			t.Errorf("mapSeverity(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}
