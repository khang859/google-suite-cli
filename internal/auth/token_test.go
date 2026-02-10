package auth

import (
	"errors"
	"os"
	"strings"
	"testing"

	"golang.org/x/oauth2"
)

func TestTokenPathFor(t *testing.T) {
	tests := []struct {
		name       string
		email      string
		wantErr    bool
		wantSuffix string
	}{
		{
			name:    "should error when email is empty",
			email:   "",
			wantErr: true,
		},
		{
			name:       "should return path with tokens prefix for valid email",
			email:      "alice@example.com",
			wantSuffix: "tokens/alice@example.com.json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("XDG_CONFIG_HOME", t.TempDir())

			path, err := TokenPathFor(tt.email)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !strings.HasSuffix(path, tt.wantSuffix) {
				t.Errorf("path = %q, want suffix %q", path, tt.wantSuffix)
			}
		})
	}
}

func TestSaveAndLoadTokenRoundTrip(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	email := "roundtrip@example.com"
	original := &oauth2.Token{
		AccessToken:  "test-access",
		RefreshToken: "test-refresh",
		TokenType:    "Bearer",
	}

	if err := SaveTokenFor(email, original); err != nil {
		t.Fatalf("SaveTokenFor failed: %v", err)
	}

	loaded, err := LoadTokenFor(email)
	if err != nil {
		t.Fatalf("LoadTokenFor failed: %v", err)
	}

	if loaded.AccessToken != original.AccessToken {
		t.Errorf("AccessToken = %q, want %q", loaded.AccessToken, original.AccessToken)
	}
	if loaded.RefreshToken != original.RefreshToken {
		t.Errorf("RefreshToken = %q, want %q", loaded.RefreshToken, original.RefreshToken)
	}
	if loaded.TokenType != original.TokenType {
		t.Errorf("TokenType = %q, want %q", loaded.TokenType, original.TokenType)
	}
}

func TestLoadTokenForMissingEmail(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	_, err := LoadTokenFor("nonexistent@example.com")
	if err == nil {
		t.Fatal("expected error for missing token, got nil")
	}
	if !errors.Is(err, os.ErrNotExist) {
		t.Errorf("expected os.ErrNotExist, got: %v", err)
	}
}

func TestDeleteTokenFor(t *testing.T) {
	t.Run("should make load fail after save and delete", func(t *testing.T) {
		t.Setenv("XDG_CONFIG_HOME", t.TempDir())

		email := "delete-me@example.com"
		token := &oauth2.Token{
			AccessToken:  "to-be-deleted",
			RefreshToken: "refresh",
			TokenType:    "Bearer",
		}

		if err := SaveTokenFor(email, token); err != nil {
			t.Fatalf("SaveTokenFor failed: %v", err)
		}

		if err := DeleteTokenFor(email); err != nil {
			t.Fatalf("DeleteTokenFor failed: %v", err)
		}

		_, err := LoadTokenFor(email)
		if err == nil {
			t.Fatal("expected error after delete, got nil")
		}
		if !errors.Is(err, os.ErrNotExist) {
			t.Errorf("expected os.ErrNotExist after delete, got: %v", err)
		}
	})

	t.Run("should be no-op when deleting nonexistent token", func(t *testing.T) {
		t.Setenv("XDG_CONFIG_HOME", t.TempDir())

		if err := DeleteTokenFor("ghost@example.com"); err != nil {
			t.Fatalf("expected no error deleting nonexistent token, got: %v", err)
		}
	})
}
