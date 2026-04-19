package vault

import (
	"fmt"
	"time"

	vaultapi "github.com/hashicorp/vault/api"
)

// Client wraps the Vault API client with helper methods.
type Client struct {
	vc *vaultapi.Client
}

// NewClient creates a new Vault client using the provided address and token.
func NewClient(address, token string) (*Client, error) {
	cfg := vaultapi.DefaultConfig()
	cfg.Address = address

	vc, err := vaultapi.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create vault client: %w", err)
	}

	vc.SetToken(token)
	return &Client{vc: vc}, nil
}

// SecretMeta holds metadata about a secret including its expiry.
type SecretMeta struct {
	Path      string
	ExpiresAt time.Time
	TTL       time.Duration
}

// GetSecretMeta retrieves lease/TTL metadata for a secret at the given path.
func (c *Client) GetSecretMeta(path string) (*SecretMeta, error) {
	secret, err := c.vc.Logical().Read(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read secret at %s: %w", path, err)
	}
	if secret == nil {
		return nil, fmt.Errorf("no secret found at path: %s", path)
	}

	ttl := time.Duration(secret.LeaseDuration) * time.Second
	expiry := time.Now().Add(ttl)

	return &SecretMeta{
		Path:      path,
		ExpiresAt: expiry,
		TTL:       ttl,
	}, nil
}
