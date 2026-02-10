package auth

import (
	"os"
	"testing"
	"time"
)

func TestLoadAccountStoreFileNotExist(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	store, err := LoadAccountStore()
	if err != nil {
		t.Fatalf("expected nil error for missing file, got: %v", err)
	}
	if len(store.Accounts) != 0 {
		t.Errorf("expected 0 accounts, got %d", len(store.Accounts))
	}
}

func TestAccountStoreSaveAndLoadRoundTrip(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	now := time.Now().UTC().Truncate(time.Second)
	original := &AccountStore{
		Active: "alice@example.com",
		Accounts: []AccountEntry{
			{Email: "alice@example.com", AddedAt: now},
			{Email: "bob@example.com", AddedAt: now},
		},
	}

	if err := original.Save(); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	loaded, err := LoadAccountStore()
	if err != nil {
		t.Fatalf("LoadAccountStore failed: %v", err)
	}

	if loaded.Active != original.Active {
		t.Errorf("Active = %q, want %q", loaded.Active, original.Active)
	}
	if len(loaded.Accounts) != len(original.Accounts) {
		t.Fatalf("account count = %d, want %d", len(loaded.Accounts), len(original.Accounts))
	}
	for i, acct := range loaded.Accounts {
		if acct.Email != original.Accounts[i].Email {
			t.Errorf("Accounts[%d].Email = %q, want %q", i, acct.Email, original.Accounts[i].Email)
		}
	}
}

func TestLoadAccountStoreCorruptJSON(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	path, err := AccountStorePath()
	if err != nil {
		t.Fatalf("AccountStorePath failed: %v", err)
	}

	if err := os.WriteFile(path, []byte("{not valid json!!!"), 0600); err != nil {
		t.Fatalf("failed to write corrupt file: %v", err)
	}

	_, err = LoadAccountStore()
	if err == nil {
		t.Fatal("expected error for corrupt JSON, got nil")
	}
}
