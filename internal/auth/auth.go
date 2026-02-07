// Package auth provides OAuth2 PKCE authentication for Google Gmail API.
package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"google.golang.org/api/gmail/v1"
)

// Config holds the authentication configuration.
type Config struct {
	// CredentialsFile is the path to the OAuth2 client credentials JSON file.
	CredentialsFile string
	// CredentialsJSON is the raw JSON content of OAuth2 client credentials.
	CredentialsJSON []byte
	// UserEmail is deprecated and ignored. Kept for Phase 10 CLI flag cleanup.
	UserEmail string
}

// LoadCredentials loads credentials JSON from various sources.
// Priority order:
// 1. cfg.CredentialsJSON if set
// 2. cfg.CredentialsFile if set
// 3. GOOGLE_CREDENTIALS env var (JSON content)
// 4. GOOGLE_APPLICATION_CREDENTIALS env var (file path)
func LoadCredentials(cfg Config) ([]byte, error) {
	if len(cfg.CredentialsJSON) > 0 {
		return cfg.CredentialsJSON, nil
	}

	if cfg.CredentialsFile != "" {
		data, err := os.ReadFile(cfg.CredentialsFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read credentials file %s: %w", cfg.CredentialsFile, err)
		}
		return data, nil
	}

	if jsonContent := os.Getenv("GOOGLE_CREDENTIALS"); jsonContent != "" {
		return []byte(jsonContent), nil
	}

	if filePath := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"); filePath != "" {
		data, err := os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read GOOGLE_APPLICATION_CREDENTIALS file %s: %w", filePath, err)
		}
		return data, nil
	}

	return nil, fmt.Errorf("no credentials found: set --credentials-file flag, GOOGLE_CREDENTIALS env var (JSON), or GOOGLE_APPLICATION_CREDENTIALS env var (file path)")
}

// extractOAuth2ClientCreds extracts the client_id and client_secret from
// an OAuth2 client credentials JSON file (either "installed" or "web" format).
func extractOAuth2ClientCreds(jsonData []byte) (clientID, clientSecret string, err error) {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(jsonData, &raw); err != nil {
		return "", "", fmt.Errorf("failed to parse credentials JSON: %w", err)
	}

	var clientJSON json.RawMessage
	if data, ok := raw["installed"]; ok {
		clientJSON = data
	} else if data, ok := raw["web"]; ok {
		clientJSON = data
	} else {
		return "", "", fmt.Errorf("credentials JSON has neither \"installed\" nor \"web\" key")
	}

	var creds struct {
		ClientID     string `json:"client_id"`
		ClientSecret string `json:"client_secret"`
	}
	if err := json.Unmarshal(clientJSON, &creds); err != nil {
		return "", "", fmt.Errorf("failed to parse client credentials: %w", err)
	}

	if creds.ClientID == "" {
		return "", "", fmt.Errorf("client_id is empty in credentials JSON")
	}

	return creds.ClientID, creds.ClientSecret, nil
}

// Login performs the OAuth2 PKCE browser login flow, saves the token, and
// returns the authenticated user's email address.
func Login(ctx context.Context, credJSON []byte) (string, error) {
	clientID, clientSecret, err := extractOAuth2ClientCreds(credJSON)
	if err != nil {
		return "", fmt.Errorf("failed to extract OAuth2 client credentials: %w", err)
	}

	oauthCfg := NewOAuth2Config(clientID, clientSecret)
	token, err := oauthCfg.Authenticate(ctx)
	if err != nil {
		return "", fmt.Errorf("authentication failed: %w", err)
	}

	if err := SaveToken(token); err != nil {
		return "", fmt.Errorf("failed to save token: %w", err)
	}

	service, err := oauthCfg.NewGmailService(ctx, token)
	if err != nil {
		return "", fmt.Errorf("failed to create Gmail service: %w", err)
	}

	profile, err := service.Users.GetProfile("me").Do()
	if err != nil {
		return "", fmt.Errorf("failed to get user profile: %w", err)
	}

	return profile.EmailAddress, nil
}

// NewGmailService creates an authenticated Gmail service using OAuth2 client
// credentials and a cached token from a prior login.
func NewGmailService(ctx context.Context, cfg Config) (*gmail.Service, error) {
	credJSON, err := LoadCredentials(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to load credentials: %w", err)
	}

	clientID, clientSecret, err := extractOAuth2ClientCreds(credJSON)
	if err != nil {
		return nil, err
	}

	token, err := LoadToken()
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("no OAuth2 token found. Run 'gsuite login' first to authenticate")
		}
		return nil, fmt.Errorf("failed to load OAuth2 token: %w", err)
	}

	oauthCfg := NewOAuth2Config(clientID, clientSecret)
	return oauthCfg.NewGmailService(ctx, token)
}
