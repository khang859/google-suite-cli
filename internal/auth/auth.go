// Package auth provides authentication functionality for Google APIs
// using service accounts or OAuth2 client credentials.
package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

// credentialType represents the type of Google credentials JSON.
type credentialType int

const (
	credServiceAccount credentialType = iota
	credOAuth2Client
)

// Config holds the authentication configuration.
type Config struct {
	// CredentialsFile is the path to the service account JSON file.
	CredentialsFile string
	// CredentialsJSON is the raw JSON content of service account credentials.
	CredentialsJSON []byte
	// UserEmail is the email of the user to impersonate (required for domain-wide delegation).
	UserEmail string
}

// LoadCredentials loads credentials JSON from various sources.
// Priority order:
// 1. cfg.CredentialsJSON if set
// 2. cfg.CredentialsFile if set
// 3. GOOGLE_CREDENTIALS env var (JSON content)
// 4. GOOGLE_APPLICATION_CREDENTIALS env var (file path)
func LoadCredentials(cfg Config) ([]byte, error) {
	// Option 1: Direct JSON content from config
	if len(cfg.CredentialsJSON) > 0 {
		return cfg.CredentialsJSON, nil
	}

	// Option 2: File path from config
	if cfg.CredentialsFile != "" {
		data, err := os.ReadFile(cfg.CredentialsFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read credentials file %s: %w", cfg.CredentialsFile, err)
		}
		return data, nil
	}

	// Option 3: GOOGLE_CREDENTIALS env var (JSON content)
	if jsonContent := os.Getenv("GOOGLE_CREDENTIALS"); jsonContent != "" {
		return []byte(jsonContent), nil
	}

	// Option 4: GOOGLE_APPLICATION_CREDENTIALS env var (file path)
	if filePath := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"); filePath != "" {
		data, err := os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read GOOGLE_APPLICATION_CREDENTIALS file %s: %w", filePath, err)
		}
		return data, nil
	}

	return nil, fmt.Errorf("no credentials found: set --credentials-file flag, GOOGLE_CREDENTIALS env var (JSON), or GOOGLE_APPLICATION_CREDENTIALS env var (file path)")
}

// detectCredentialType parses the JSON and determines whether it represents
// service account credentials or OAuth2 client credentials.
func detectCredentialType(jsonData []byte) (credentialType, error) {
	var raw map[string]interface{}
	if err := json.Unmarshal(jsonData, &raw); err != nil {
		return 0, fmt.Errorf("failed to parse credentials JSON: %w", err)
	}

	// Service account JSON has "type": "service_account"
	if t, ok := raw["type"].(string); ok && t == "service_account" {
		return credServiceAccount, nil
	}

	// OAuth2 client credentials JSON has "installed" or "web" key
	if _, ok := raw["installed"]; ok {
		return credOAuth2Client, nil
	}
	if _, ok := raw["web"]; ok {
		return credOAuth2Client, nil
	}

	return 0, fmt.Errorf("unrecognized credentials format: expected service account JSON (with \"type\": \"service_account\"), or OAuth2 client JSON (with \"installed\" or \"web\" key)")
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

// Login performs the OAuth2 browser-based login flow. It validates that the
// credentials are OAuth2 client credentials (not service account), runs the
// PKCE authorization flow, saves the token, and returns the authenticated
// user's email address.
func Login(ctx context.Context, credJSON []byte) (string, error) {
	// Verify credential type is OAuth2
	credType, err := detectCredentialType(credJSON)
	if err != nil {
		return "", err
	}
	if credType == credServiceAccount {
		return "", fmt.Errorf("login is only for OAuth2 client credentials; service accounts authenticate automatically")
	}

	// Extract client credentials
	clientID, clientSecret, err := extractOAuth2ClientCreds(credJSON)
	if err != nil {
		return "", fmt.Errorf("failed to extract OAuth2 client credentials: %w", err)
	}

	// Run OAuth2 PKCE flow
	oauthCfg := NewOAuth2Config(clientID, clientSecret)
	token, err := oauthCfg.Authenticate(ctx)
	if err != nil {
		return "", fmt.Errorf("authentication failed: %w", err)
	}

	// Save token
	if err := SaveToken(token); err != nil {
		return "", fmt.Errorf("failed to save token: %w", err)
	}

	// Create a temporary Gmail service to get the user's email
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

// NewGmailService creates an authenticated Gmail service. It auto-detects the
// credential type from the JSON and dispatches to the appropriate auth flow:
//   - Service account: uses domain-wide delegation (requires cfg.UserEmail)
//   - OAuth2 client: uses cached token from ~/.config/gsuite/token.json
func NewGmailService(ctx context.Context, cfg Config) (*gmail.Service, error) {
	// Load credentials
	credJSON, err := LoadCredentials(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to load credentials: %w", err)
	}

	// Detect credential type
	credType, err := detectCredentialType(credJSON)
	if err != nil {
		return nil, err
	}

	switch credType {
	case credServiceAccount:
		return newServiceAccountGmailService(ctx, cfg, credJSON)
	case credOAuth2Client:
		return newOAuth2GmailService(ctx, credJSON)
	default:
		return nil, fmt.Errorf("unsupported credential type")
	}
}

// newServiceAccountGmailService creates a Gmail service using service account
// credentials with domain-wide delegation.
func newServiceAccountGmailService(ctx context.Context, cfg Config, credJSON []byte) (*gmail.Service, error) {
	if cfg.UserEmail == "" {
		return nil, fmt.Errorf("user email is required for service account authentication: set --user flag")
	}

	jwtConfig, err := google.JWTConfigFromJSON(credJSON, gmail.GmailModifyScope)
	if err != nil {
		return nil, fmt.Errorf("failed to parse service account credentials: %w", err)
	}

	jwtConfig.Subject = cfg.UserEmail
	client := jwtConfig.Client(ctx)

	service, err := gmail.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("failed to create Gmail service: %w", err)
	}

	return service, nil
}

// newOAuth2GmailService creates a Gmail service using OAuth2 client credentials
// and a cached token.
func newOAuth2GmailService(ctx context.Context, credJSON []byte) (*gmail.Service, error) {
	clientID, clientSecret, err := extractOAuth2ClientCreds(credJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to extract OAuth2 client credentials: %w", err)
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
