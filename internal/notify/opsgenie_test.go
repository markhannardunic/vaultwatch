package notify

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewOpsGenieClient_EmptyKey(t *testing.T) {
	_, err := NewOpsGenieClient("")
	if err == nil {
		t.Fatal("expected error for empty api key, got nil")
	}
}

func TestNewOpsGenieClient_ValidKey(t *testing.T) {
	client, err := NewOpsGenieClient("test-api-key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestOpsGenieClient_Send_Success(t *testing.T) {
	var received opsgeniePayload
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") == "" {
			t.Error("expected Authorization header")
		}
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Errorf("failed to decode body: %v", err)
		}
		w.WriteHeader(http.StatusAccepted)
	}))
	defer server.Close()

	client, _ := NewOpsGenieClient("key-123")
	client.apiURL = server.URL

	if err := client.Send("critical", "cert expires in 2 days"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if received.Priority != "P1" {
		t.Errorf("expected P1, got %s", received.Priority)
	}
}

func TestOpsGenieClient_Send_NonOKStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer server.Close()

	client, _ := NewOpsGenieClient("key-123")
	client.apiURL = server.URL

	if err := client.Send("warning", "some warning"); err == nil {
		t.Fatal("expected error for non-2xx status, got nil")
	}
}

func TestMapOpsgeniePriority(t *testing.T) {
	cases := []struct {
		level    string
		wantPrio string
	}{
		{"critical", "P1"},
		{"warning", "P3"},
		{"healthy", "P5"},
		{"unknown", "P5"},
	}
	for _, tc := range cases {
		got := mapOpsgeniePriority(tc.level)
		if got != tc.wantPrio {
			t.Errorf("mapOpsgeniePriority(%q) = %q, want %q", tc.level, got, tc.wantPrio)
		}
	}
}
