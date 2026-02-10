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

// LoadCredentials loads OAuth2 client credentials JSON from environment variables.
// Priority: GOOGLE_CREDENTIALS (raw JSON) then GOOGLE_APPLICATION_CREDENTIALS (file path).
func LoadCredentials() ([]byte, error) {
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

	return nil, fmt.Errorf("no OAuth2 client credentials found: set GOOGLE_CREDENTIALS env var (JSON) or GOOGLE_APPLICATION_CREDENTIALS env var (file path)")
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

// Login performs the OAuth2 PKCE browser login flow, saves the per-account
// token, updates the account store, and returns the authenticated user's email.
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

	service, err := oauthCfg.NewGmailService(ctx, token)
	if err != nil {
		return "", fmt.Errorf("failed to create Gmail service: %w", err)
	}

	profile, err := service.Users.GetProfile("me").Do()
	if err != nil {
		return "", fmt.Errorf("failed to get user profile: %w", err)
	}
	email := profile.EmailAddress

	if err := SaveTokenFor(email, token); err != nil {
		return "", fmt.Errorf("failed to save token for %s: %w", email, err)
	}

	store, err := LoadAccountStore()
	if err != nil {
		return "", fmt.Errorf("failed to load account store: %w", err)
	}
	if err := store.AddAccount(email); err != nil {
		return "", fmt.Errorf("failed to add account %s: %w", email, err)
	}
	if err := store.Save(); err != nil {
		return "", fmt.Errorf("failed to save account store: %w", err)
	}

	return email, nil
}

// NewGmailService creates an authenticated Gmail service for the given account.
// If account is empty, the active account from AccountStore is used.
// Runs EnsureMigrated to transparently upgrade legacy single-token setups.
func NewGmailService(ctx context.Context, account string) (*gmail.Service, error) {
	credJSON, err := LoadCredentials()
	if err != nil {
		return nil, fmt.Errorf("failed to load credentials: %w", err)
	}

	clientID, clientSecret, err := extractOAuth2ClientCreds(credJSON)
	if err != nil {
		return nil, err
	}

	if err := EnsureMigrated(ctx); err != nil {
		return nil, fmt.Errorf("failed to run migration: %w", err)
	}

	resolvedEmail := account
	if resolvedEmail == "" {
		store, err := LoadAccountStore()
		if err != nil {
			return nil, fmt.Errorf("failed to load account store: %w", err)
		}
		resolvedEmail, err = store.GetActive()
		if err != nil {
			return nil, fmt.Errorf("no authenticated accounts. Run 'gsuite login' first")
		}
	}

	token, err := LoadTokenFor(resolvedEmail)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("no token for account %s. Run 'gsuite login' to authenticate", resolvedEmail)
		}
		return nil, fmt.Errorf("failed to load token for %s: %w", resolvedEmail, err)
	}

	oauthCfg := NewOAuth2Config(clientID, clientSecret)
	return oauthCfg.NewGmailService(ctx, token)
}
