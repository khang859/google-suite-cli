package auth

import (
	"context"
	"fmt"
	"os"
)

// MigrateIfNeeded transparently migrates a legacy single-token setup
// (token.json) to the multi-account format (tokens/<email>.json + accounts.json).
// Safe to call multiple times — returns nil if already migrated or nothing to migrate.
func MigrateIfNeeded(ctx context.Context, credJSON []byte) error {
	store, err := LoadAccountStore()
	if err != nil {
		return fmt.Errorf("migration: failed to load account store: %w", err)
	}
	if len(store.Accounts) > 0 {
		return nil
	}

	legacyPath, err := LegacyTokenPath()
	if err != nil {
		return fmt.Errorf("migration: failed to resolve legacy token path: %w", err)
	}
	if _, err := os.Stat(legacyPath); os.IsNotExist(err) {
		return nil
	}

	token, err := LoadLegacyToken()
	if err != nil {
		return fmt.Errorf("migration: failed to load legacy token: %w", err)
	}

	clientID, clientSecret, err := extractOAuth2ClientCreds(credJSON)
	if err != nil {
		return fmt.Errorf("migration: failed to extract OAuth2 client credentials: %w", err)
	}

	oauthCfg := NewOAuth2Config(clientID, clientSecret)
	service, err := oauthCfg.NewGmailService(ctx, token)
	if err != nil {
		return fmt.Errorf("migration: failed to create Gmail service from legacy token: %w", err)
	}

	profile, err := service.Users.GetProfile("me").Do()
	if err != nil {
		return fmt.Errorf("migration: failed to discover email from legacy token: %w", err)
	}
	email := profile.EmailAddress

	if err := SaveTokenFor(email, token); err != nil {
		return fmt.Errorf("migration: failed to save per-account token for %s: %w", email, err)
	}

	if err := store.AddAccount(email); err != nil {
		return fmt.Errorf("migration: failed to add account %s: %w", email, err)
	}
	if err := store.Save(); err != nil {
		return fmt.Errorf("migration: failed to save account store: %w", err)
	}

	backupPath := legacyPath + ".bak"
	if err := os.Rename(legacyPath, backupPath); err != nil {
		return fmt.Errorf("migration: failed to rename legacy token to %s: %w", backupPath, err)
	}

	return nil
}

// EnsureMigrated is a convenience wrapper that loads credentials and calls
// MigrateIfNeeded. If no credentials are configured yet, migration is
// silently skipped (user hasn't set up the CLI).
func EnsureMigrated(ctx context.Context) error {
	credJSON, err := LoadCredentials()
	if err != nil {
		return nil
	}
	return MigrateIfNeeded(ctx, credJSON)
}
