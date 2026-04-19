package notify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// SlackClient sends alert messages to a Slack webhook.
type SlackClient struct {
	webhookURL string
	httpClient *http.Client
}

type slackPayload struct {
	Text string `json:"text"`
}

// NewSlackClient creates a SlackClient with the given webhook URL.
func NewSlackClient(webhookURL string) (*SlackClient, error) {
	if webhookURL == "" {
		return nil, fmt.Errorf("slack webhook URL must not be empty")
	}
	return &SlackClient{
		webhookURL: webhookURL,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}, nil
}

// Send posts the message to the configured Slack webhook.
func (s *SlackClient) Send(message string) error {
	payload := slackPayload{Text: message}
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("slack: marshal payload: %w", err)
	}
	resp, err := s.httpClient.Post(s.webhookURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("slack: post: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("slack: unexpected status %d", resp.StatusCode)
	}
	return nil
}
