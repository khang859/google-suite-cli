package auth

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/oauth2"
)

const (
	tokenDir  = "gsuite"
	tokenFile = "token.json"
	tokensDir = "tokens"
)

// LegacyTokenPath returns the path to the legacy single-user token file
// at ~/.config/gsuite/token.json. Preserved for migration (11-02).
func LegacyTokenPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		home, homeErr := os.UserHomeDir()
		if homeErr != nil {
			return "", fmt.Errorf("failed to determine config directory: %w", err)
		}
		configDir = filepath.Join(home, ".config")
	}

	dir := filepath.Join(configDir, tokenDir)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", fmt.Errorf("failed to create token directory %s: %w", dir, err)
	}

	return filepath.Join(dir, tokenFile), nil
}

// saveLegacyToken writes a token to the legacy single-user path.
// Unexported — only used by migration.
func saveLegacyToken(token *oauth2.Token) error {
	path, err := LegacyTokenPath()
	if err != nil {
		return fmt.Errorf("failed to resolve token path: %w", err)
	}

	data, err := json.MarshalIndent(token, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal token: %w", err)
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("failed to write token file %s: %w", path, err)
	}

	return nil
}

// LoadLegacyToken reads the legacy single-user token from ~/.config/gsuite/token.json.
// Preserved for migration (11-02). Returns os.ErrNotExist if file is missing.
func LoadLegacyToken() (*oauth2.Token, error) {
	path, err := LegacyTokenPath()
	if err != nil {
		return nil, fmt.Errorf("failed to resolve token path: %w", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var token oauth2.Token
	if err := json.Unmarshal(data, &token); err != nil {
		return nil, fmt.Errorf("failed to parse token file %s: %w", path, err)
	}

	return &token, nil
}

// TokensDir returns the path to ~/.config/gsuite/tokens/,
// creating the directory with 0700 permissions if needed.
func TokensDir() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		home, homeErr := os.UserHomeDir()
		if homeErr != nil {
			return "", fmt.Errorf("failed to determine config directory: %w", err)
		}
		configDir = filepath.Join(home, ".config")
	}

	dir := filepath.Join(configDir, tokenDir, tokensDir)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", fmt.Errorf("failed to create tokens directory %s: %w", dir, err)
	}

	return dir, nil
}

// TokenPathFor returns the per-account token file path:
// ~/.config/gsuite/tokens/<email>.json
func TokenPathFor(email string) (string, error) {
	if email == "" {
		return "", fmt.Errorf("email cannot be empty")
	}

	dir, err := TokensDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(dir, email+".json"), nil
}

// SaveTokenFor writes an OAuth2 token to the per-account token file
// with 0600 permissions.
func SaveTokenFor(email string, token *oauth2.Token) error {
	path, err := TokenPathFor(email)
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(token, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal token: %w", err)
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("failed to write token file %s: %w", path, err)
	}

	return nil
}

// LoadTokenFor reads the per-account OAuth2 token. Returns os.ErrNotExist
// if the token file is not found.
func LoadTokenFor(email string) (*oauth2.Token, error) {
	path, err := TokenPathFor(email)
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var token oauth2.Token
	if err := json.Unmarshal(data, &token); err != nil {
		return nil, fmt.Errorf("failed to parse token file %s: %w", path, err)
	}

	return &token, nil
}

// DeleteTokenFor removes the per-account token file.
// No error if the file doesn't exist.
func DeleteTokenFor(email string) error {
	path, err := TokenPathFor(email)
	if err != nil {
		return err
	}

	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove token file %s: %w", path, err)
	}

	return nil
}
