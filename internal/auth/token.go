package auth

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/oauth2"
)

const (
	// tokenDir is the directory name under the user config dir.
	tokenDir = "gsuite"
	// tokenFile is the filename for the persisted OAuth2 token.
	tokenFile = "token.json"
)

// TokenPath returns the path to the persisted OAuth2 token file.
// The token is stored at ~/.config/gsuite/token.json (XDG-compatible).
// If os.UserConfigDir fails, it falls back to $HOME/.config.
// Parent directories are created with 0700 permissions if they don't exist.
func TokenPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		// Fallback to $HOME/.config
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

// SaveToken serializes an OAuth2 token to JSON and writes it to the token file
// with 0600 permissions. The token file includes access_token, refresh_token,
// token_type, and expiry as marshaled by the oauth2.Token type.
func SaveToken(token *oauth2.Token) error {
	path, err := TokenPath()
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

// LoadToken reads the persisted OAuth2 token from the token file and
// deserializes it. If the token file does not exist, the underlying
// os.ErrNotExist error is returned unwrapped so callers can check with
// errors.Is(err, os.ErrNotExist).
func LoadToken() (*oauth2.Token, error) {
	path, err := TokenPath()
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
