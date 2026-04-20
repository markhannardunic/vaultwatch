package notify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// TeamsClient sends alerts to a Microsoft Teams channel via incoming webhook.
type TeamsClient struct {
	webhookURL string
	httpClient *http.Client
}

type teamsPayload struct {
	Type       string         `json:"@type"`
	Context    string         `json:"@context"`
	ThemeColor string         `json:"themeColor"`
	Summary    string         `json:"summary"`
	Sections   []teamsSection `json:"sections"`
}

type teamsSection struct {
	ActivityTitle string `json:"activityTitle"`
	ActivityText  string `json:"activityText"`
}

// NewTeamsClient creates a new TeamsClient for the given webhook URL.
func NewTeamsClient(webhookURL string) (*TeamsClient, error) {
	if webhookURL == "" {
		return nil, fmt.Errorf("teams: webhook URL must not be empty")
	}
	return &TeamsClient{
		webhookURL: webhookURL,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}, nil
}

// Send delivers the alert message to the Teams channel.
func (c *TeamsClient) Send(level, message string) error {
	color := colorForLevel(level)
	payload := teamsPayload{
		Type:       "MessageCard",
		Context:    "http://schema.org/extensions",
		ThemeColor: color,
		Summary:    message,
		Sections: []teamsSection{
			{ActivityTitle: fmt.Sprintf("VaultWatch Alert [%s]", level), ActivityText: message},
		},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("teams: marshal payload: %w", err)
	}
	resp, err := c.httpClient.Post(c.webhookURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("teams: send request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("teams: unexpected status %d", resp.StatusCode)
	}
	return nil
}

func colorForLevel(level string) string {
	switch level {
	case "critical":
		return "FF0000"
	case "warning":
		return "FFA500"
	default:
		return "00FF00"
	}
}
