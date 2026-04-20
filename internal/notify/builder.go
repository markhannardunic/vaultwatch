package notify

import (
	"fmt"
	"io"

	"github.com/youorg/vaultwatch/internal/config"
)

// SenderConfig holds configuration for a single notification target.
type SenderConfig struct {
	Type       string
	WebhookURL string
	SlackURL   string
	TeamsURL   string
	PDKey      string
	OGKey      string
	SMTPHost   string
	SMTPPort   int
	From       string
	To         []string
	Level      string
}

// BuildRouter constructs a Router from the application config, writing
// fallback output to w when no senders match.
func BuildRouter(cfg *config.Config, w io.Writer) (*Router, error) {
	router := NewRouter()
	if cfg == nil || len(cfg.Notify) == 0 {
		router.Register("*", NewWriterSender(w))
		return router, nil
	}
	for _, n := range cfg.Notify {
		sender, err := buildSender(n)
		if err != nil {
			return nil, fmt.Errorf("notify builder: %w", err)
		}
		level := n.Level
		if level == "" {
			level = "*"
		}
		router.Register(level, sender)
	}
	return router, nil
}

func buildSender(n config.NotifyEntry) (Sender, error) {
	switch n.Type {
	case "slack":
		return NewSlackClient(n.WebhookURL)
	case "teams":
		return NewTeamsClient(n.WebhookURL)
	case "webhook":
		return NewWebhookClient(n.WebhookURL)
	case "pagerduty":
		return NewPagerDutyClient(n.APIKey)
	case "opsgenie":
		return NewOpsGenieClient(n.APIKey)
	default:
		return nil, fmt.Errorf("unknown notify type %q", n.Type)
	}
}
