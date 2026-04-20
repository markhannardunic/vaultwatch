package notify

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewTeamsClient_EmptyURL(t *testing.T) {
	_, err := NewTeamsClient("")
	if err == nil {
		t.Fatal("expected error for empty URL")
	}
}

func TestNewTeamsClient_ValidURL(t *testing.T) {
	client, err := NewTeamsClient("https://outlook.office.com/webhook/test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestTeamsClient_Send_Success(t *testing.T) {
	var received teamsPayload
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Errorf("decode body: %v", err)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client, _ := NewTeamsClient(server.URL)
	if err := client.Send("critical", "cert expires soon"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if received.ThemeColor != "FF0000" {
		t.Errorf("expected red theme color, got %s", received.ThemeColor)
	}
	if len(received.Sections) == 0 {
		t.Error("expected at least one section")
	}
}

func TestTeamsClient_Send_NonOKStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client, _ := NewTeamsClient(server.URL)
	if err := client.Send("warning", "test"); err == nil {
		t.Fatal("expected error for non-OK status")
	}
}

func TestColorForLevel(t *testing.T) {
	cases := []struct{ level, want string }{
		{"critical", "FF0000"},
		{"warning", "FFA500"},
		{"healthy", "00FF00"},
		{"unknown", "00FF00"},
	}
	for _, tc := range cases {
		got := colorForLevel(tc.level)
		if got != tc.want {
			t.Errorf("colorForLevel(%q) = %q, want %q", tc.level, got, tc.want)
		}
	}
}
