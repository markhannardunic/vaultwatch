package notify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const opsgenieAPIURL = "https://api.opsgenie.com/v2/alerts"

// OpsGenieClient sends alerts to OpsGenie.
type OpsGenieClient struct {
	apiKey  string
	apiURL  string
	httpClient *http.Client
}

type opsgeniePayload struct {
	Message     string            `json:"message"`
	Description string            `json:"description"`
	Priority    string            `json:"priority"`
	Tags        []string          `json:"tags,omitempty"`
	Details     map[string]string `json:"details,omitempty"`
}

// NewOpsGenieClient creates a new OpsGenieClient.
// Returns an error if apiKey is empty.
func NewOpsGenieClient(apiKey string) (*OpsGenieClient, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("opsgenie: api key must not be empty")
	}
	return &OpsGenieClient{
		apiKey: apiKey,
		apiURL: opsgenieAPIURL,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}, nil
}

// Send delivers an alert message to OpsGenie with the given level as priority.
func (c *OpsGenieClient) Send(level, message string) error {
	payload := opsgeniePayload{
		Message:     fmt.Sprintf("[%s] %s", level, message),
		Description: message,
		Priority:    mapOpsgeniePriority(level),
		Tags:        []string{"vaultwatch", level},
		Details:     map[string]string{"level": level},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("opsgenie: failed to marshal payload: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, c.apiURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("opsgenie: failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "GenieKey "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("opsgenie: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("opsgenie: unexpected status code: %d", resp.StatusCode)
	}
	return nil
}

func mapOpsgeniePriority(level string) string {
	switch level {
	case "critical":
		return "P1"
	case "warning":
		return "P3"
	default:
		return "P5"
	}
}
