// Package auth provides authentication functionality for Google APIs using service accounts.
package auth

import (
	"context"
	"fmt"
	"os"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
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

// LoadCredentials loads service account credentials from various sources.
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

// NewGmailService creates an authenticated Gmail service using service account credentials.
// The cfg.UserEmail field is required for domain-wide delegation - it specifies which user
// the service account will impersonate.
func NewGmailService(ctx context.Context, cfg Config) (*gmail.Service, error) {
	// Validate that UserEmail is set (required for domain-wide delegation)
	if cfg.UserEmail == "" {
		return nil, fmt.Errorf("user email is required for domain-wide delegation: set --user flag")
	}

	// Load credentials
	credJSON, err := LoadCredentials(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to load credentials: %w", err)
	}

	// Parse the service account JSON and create JWT config
	// Scopes: GmailModifyScope provides full read/write access to mailbox
	jwtConfig, err := google.JWTConfigFromJSON(credJSON, gmail.GmailModifyScope)
	if err != nil {
		return nil, fmt.Errorf("failed to parse service account credentials: %w", err)
	}

	// Set the subject (user to impersonate) for domain-wide delegation
	jwtConfig.Subject = cfg.UserEmail

	// Create HTTP client with the JWT config
	client := jwtConfig.Client(ctx)

	// Create and return Gmail service
	service, err := gmail.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("failed to create Gmail service: %w", err)
	}

	return service, nil
}
