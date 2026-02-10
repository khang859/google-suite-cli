package auth

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const accountsFile = "accounts.json"

type AccountEntry struct {
	Email   string    `json:"email"`
	AddedAt time.Time `json:"added_at"`
}

type AccountStore struct {
	Active   string         `json:"active"`
	Accounts []AccountEntry `json:"accounts"`
}

// AccountStorePath returns the path to ~/.config/gsuite/accounts.json,
// creating the parent directory if needed.
func AccountStorePath() (string, error) {
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
		return "", fmt.Errorf("failed to create config directory %s: %w", dir, err)
	}

	return filepath.Join(dir, accountsFile), nil
}

// LoadAccountStore reads accounts.json and returns the store.
// If the file doesn't exist, returns an empty store (not an error).
func LoadAccountStore() (*AccountStore, error) {
	path, err := AccountStorePath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &AccountStore{}, nil
		}
		return nil, fmt.Errorf("failed to read accounts file %s: %w", path, err)
	}

	var store AccountStore
	if err := json.Unmarshal(data, &store); err != nil {
		return nil, fmt.Errorf("failed to parse accounts file %s: %w", path, err)
	}

	return &store, nil
}

// Save writes the account store to accounts.json with 0600 permissions.
func (s *AccountStore) Save() error {
	path, err := AccountStorePath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal account store: %w", err)
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("failed to write accounts file %s: %w", path, err)
	}

	return nil
}

// AddAccount adds an account if not already present and sets it as active.
// If the account already exists, it just updates the active field.
func (s *AccountStore) AddAccount(email string) error {
	if email == "" {
		return fmt.Errorf("email cannot be empty")
	}

	if !s.HasAccount(email) {
		s.Accounts = append(s.Accounts, AccountEntry{
			Email:   email,
			AddedAt: time.Now().UTC(),
		})
	}

	s.Active = email
	return nil
}

// RemoveAccount removes an account entry. If the removed account was active,
// sets active to the first remaining account or "" if none left.
func (s *AccountStore) RemoveAccount(email string) error {
	if email == "" {
		return fmt.Errorf("email cannot be empty")
	}

	idx := -1
	for i, a := range s.Accounts {
		if strings.EqualFold(a.Email, email) {
			idx = i
			break
		}
	}

	if idx == -1 {
		return fmt.Errorf("account %s not found", email)
	}

	s.Accounts = append(s.Accounts[:idx], s.Accounts[idx+1:]...)

	if strings.EqualFold(s.Active, email) {
		if len(s.Accounts) > 0 {
			s.Active = s.Accounts[0].Email
		} else {
			s.Active = ""
		}
	}

	return nil
}

// SetActive sets the active account. Returns error if email not in accounts list.
func (s *AccountStore) SetActive(email string) error {
	if email == "" {
		return fmt.Errorf("email cannot be empty")
	}

	if !s.HasAccount(email) {
		return fmt.Errorf("account %s not found", email)
	}

	s.Active = email
	return nil
}

// GetActive returns the active email. If active is "" but exactly one account
// exists, returns that one. If multiple accounts and no active set, returns error.
func (s *AccountStore) GetActive() (string, error) {
	if s.Active != "" {
		return s.Active, nil
	}

	switch len(s.Accounts) {
	case 0:
		return "", fmt.Errorf("no accounts configured. Run 'gsuite login' first")
	case 1:
		return s.Accounts[0].Email, nil
	default:
		return "", fmt.Errorf("multiple accounts found but none is active. Use 'gsuite accounts switch' to set one")
	}
}

// HasAccount checks if an account exists (case-insensitive).
func (s *AccountStore) HasAccount(email string) bool {
	for _, a := range s.Accounts {
		if strings.EqualFold(a.Email, email) {
			return true
		}
	}
	return false
}

// List returns a copy of the accounts slice.
func (s *AccountStore) List() []AccountEntry {
	result := make([]AccountEntry, len(s.Accounts))
	copy(result, s.Accounts)
	return result
}
