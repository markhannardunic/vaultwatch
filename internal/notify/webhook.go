package notify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// WebhookClient sends alert payloads to a generic HTTP webhook endpoint.
type WebhookClient struct {
	url    string
	client *http.Client
}

type webhookPayload struct {
	Level   string `json:"level"`
	Message string `json:"message"`
	Path    string `json:"path"`
	SentAt  string `json:"sent_at"`
}

// NewWebhookClient creates a WebhookClient targeting the given URL.
// Returns an error if the URL is empty.
func NewWebhookClient(url string) (*WebhookClient, error) {
	if url == "" {
		return nil, fmt.Errorf("webhook url must not be empty")
	}
	return &WebhookClient{
		url: url,
		client: &http.Client{Timeout: 10 * time.Second},
	}, nil
}

// Send serialises the alert as JSON and POSTs it to the webhook URL.
func (w *WebhookClient) Send(level, message, path string) error {
	payload := webhookPayload{
		Level:   level,
		Message: message,
		Path:    path,
		SentAt:  time.Now().UTC().Format(time.RFC3339),
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("webhook: marshal payload: %w", err)
	}
	resp, err := w.client.Post(w.url, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("webhook: post: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook: unexpected status %d", resp.StatusCode)
	}
	return nil
}
