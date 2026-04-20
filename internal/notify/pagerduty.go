package notify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const pagerDutyEventURL = "https://events.pagerduty.com/v2/enqueue"

// PagerDutyClient sends alerts to PagerDuty via the Events API v2.
type PagerDutyClient struct {
	integrationKey string
	httpClient     *http.Client
	eventURL       string
}

type pdPayload struct {
	RoutingKey  string    `json:"routing_key"`
	EventAction string    `json:"event_action"`
	Payload     pdDetails `json:"payload"`
}

type pdDetails struct {
	Summary  string `json:"summary"`
	Severity string `json:"severity"`
	Source   string `json:"source"`
	Timestamp string `json:"timestamp"`
}

// NewPagerDutyClient creates a PagerDutyClient with the given integration key.
func NewPagerDutyClient(integrationKey string) (*PagerDutyClient, error) {
	if integrationKey == "" {
		return nil, fmt.Errorf("pagerduty: integration key must not be empty")
	}
	return &PagerDutyClient{
		integrationKey: integrationKey,
		httpClient:     &http.Client{Timeout: 10 * time.Second},
		eventURL:       pagerDutyEventURL,
	}, nil
}

// Send triggers a PagerDuty alert with the given message and severity.
func (c *PagerDutyClient) Send(level, message string) error {
	severity := mapSeverity(level)
	body := pdPayload{
		RoutingKey:  c.integrationKey,
		EventAction: "trigger",
		Payload: pdDetails{
			Summary:   message,
			Severity:  severity,
			Source:    "vaultwatch",
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		},
	}
	data, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("pagerduty: marshal payload: %w", err)
	}
	resp, err := c.httpClient.Post(c.eventURL, "application/json", bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("pagerduty: send request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("pagerduty: unexpected status %d", resp.StatusCode)
	}
	return nil
}

func mapSeverity(level string) string {
	switch level {
	case "critical":
		return "critical"
	case "warning":
		return "warning"
	default:
		return "info"
	}
}
