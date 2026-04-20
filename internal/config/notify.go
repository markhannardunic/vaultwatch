package config

// NotifyEntry describes a single notification target in the config file.
type NotifyEntry struct {
	// Type is the notification backend: slack, teams, webhook, pagerduty, opsgenie, email.
	Type string `yaml:"type"`

	// Level restricts delivery to a specific alert level (warning, critical).
	// Use "*" or leave empty to receive all levels.
	Level string `yaml:"level"`

	// WebhookURL is used by slack, teams, and webhook types.
	WebhookURL string `yaml:"webhook_url"`

	// APIKey is used by pagerduty and opsgenie types.
	APIKey string `yaml:"api_key"`

	// SMTP fields are used by the email type.
	SMTPHost string   `yaml:"smtp_host"`
	SMTPPort int      `yaml:"smtp_port"`
	From     string   `yaml:"from"`
	To       []string `yaml:"to"`
	Username string   `yaml:"username"`
	Password string   `yaml:"password"`
}

// Validate returns an error if the NotifyEntry is misconfigured.
func (n NotifyEntry) Validate() error {
	switch n.Type {
	case "slack", "teams", "webhook":
		if n.WebhookURL == "" {
			return &ValidationError{Field: "webhook_url", Msg: "required for type " + n.Type}
		}
	case "pagerduty", "opsgenie":
		if n.APIKey == "" {
			return &ValidationError{Field: "api_key", Msg: "required for type " + n.Type}
		}
	case "email":
		if n.SMTPHost == "" {
			return &ValidationError{Field: "smtp_host", Msg: "required for email"}
		}
		if n.From == "" {
			return &ValidationError{Field: "from", Msg: "required for email"}
		}
		if len(n.To) == 0 {
			return &ValidationError{Field: "to", Msg: "at least one recipient required for email"}
		}
	}
	return nil
}

// ValidationError represents a config field validation failure.
type ValidationError struct {
	Field string
	Msg   string
}

func (e *ValidationError) Error() string {
	return "config: " + e.Field + ": " + e.Msg
}
