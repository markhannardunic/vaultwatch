package vault

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewClient_InvalidAddress(t *testing.T) {
	// Even an invalid address should not error at construction time.
	_, err := NewClient("http://127.0.0.1:1", "test-token")
	if err != nil {
		t.Fatalf("expected no error constructing client, got: %v", err)
	}
}

func TestGetSecretMeta_Success(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"lease_id": "secret/data/myapp/db#abc123",
			"lease_duration": 3600,
			"renewable": true,
			"data": {"username": "admin"}
		}`))
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	client, err := NewClient(server.URL, "test-token")
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	meta, err := client.GetSecretMeta("secret/data/myapp/db")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedTTL := 3600 * time.Second
	if meta.TTL != expectedTTL {
		t.Errorf("expected TTL %v, got %v", expectedTTL, meta.TTL)
	}

	if meta.Path != "secret/data/myapp/db" {
		t.Errorf("expected path 'secret/data/myapp/db', got %q", meta.Path)
	}

	if meta.ExpiresAt.Before(time.Now()) {
		t.Error("expected ExpiresAt to be in the future")
	}
}

func TestGetSecretMeta_NotFound(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"errors": []}`))
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	client, _ := NewClient(server.URL, "test-token")
	_, err := client.GetSecretMeta("secret/data/missing")
	if err == nil {
		t.Fatal("expected error for missing secret, got nil")
	}
}
