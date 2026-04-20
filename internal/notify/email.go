package notify

import (
	"errors"
	"fmt"
	"net/smtp"
	"strings"
)

// EmailClient sends alert notifications via SMTP.
type EmailClient struct {
	host     string
	port     int
	from     string
	to       []string
	auth     smtp.Auth
}

// EmailConfig holds configuration for the SMTP email sender.
type EmailConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
	To       []string
}

// NewEmailClient creates a new EmailClient from the provided config.
// Returns an error if required fields are missing.
func NewEmailClient(cfg EmailConfig) (*EmailClient, error) {
	if cfg.Host == "" {
		return nil, errors.New("email: SMTP host is required")
	}
	if len(cfg.To) == 0 {
		return nil, errors.New("email: at least one recipient is required")
	}
	if cfg.From == "" {
		return nil, errors.New("email: sender address is required")
	}

	port := cfg.Port
	if port == 0 {
		port = 587
	}

	var auth smtp.Auth
	if cfg.Username != "" {
		auth = smtp.PlainAuth("", cfg.Username, cfg.Password, cfg.Host)
	}

	return &EmailClient{
		host: cfg.Host,
		port: port,
		from: cfg.From,
		to:   cfg.To,
		auth: auth,
	}, nil
}

// Send delivers the message subject and body to all configured recipients.
func (e *EmailClient) Send(subject, body string) error {
	addr := fmt.Sprintf("%s:%d", e.host, e.port)

	header := fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n",
		e.from,
		strings.Join(e.to, ", "),
		subject,
	)

	msg := []byte(header + body)

	return smtp.SendMail(addr, e.auth, e.from, e.to, msg)
}
