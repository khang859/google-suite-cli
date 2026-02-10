package auth

import (
	"testing"
	"time"
)

func TestAddAccount(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		store      AccountStore
		email      string
		wantErr    bool
		wantActive string
		wantCount  int
	}{
		{
			name:    "should error when email is empty",
			store:   AccountStore{},
			email:   "",
			wantErr: true,
		},
		{
			name:       "should add new account and set active",
			store:      AccountStore{},
			email:      "alice@example.com",
			wantActive: "alice@example.com",
			wantCount:  1,
		},
		{
			name: "should only update active when account exists",
			store: AccountStore{
				Active: "alice@example.com",
				Accounts: []AccountEntry{
					{Email: "alice@example.com", AddedAt: time.Now()},
				},
			},
			email:      "alice@example.com",
			wantActive: "alice@example.com",
			wantCount:  1,
		},
		{
			name: "should set second account as active",
			store: AccountStore{
				Active: "alice@example.com",
				Accounts: []AccountEntry{
					{Email: "alice@example.com", AddedAt: time.Now()},
				},
			},
			email:      "bob@example.com",
			wantActive: "bob@example.com",
			wantCount:  2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			store := tt.store

			err := store.AddAccount(tt.email)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if store.Active != tt.wantActive {
				t.Errorf("active = %q, want %q", store.Active, tt.wantActive)
			}
			if len(store.Accounts) != tt.wantCount {
				t.Errorf("account count = %d, want %d", len(store.Accounts), tt.wantCount)
			}
		})
	}
}

func TestRemoveAccount(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		store      AccountStore
		email      string
		wantErr    bool
		wantActive string
		wantCount  int
	}{
		{
			name:    "should error when email is empty",
			store:   AccountStore{},
			email:   "",
			wantErr: true,
		},
		{
			name: "should error when account not found",
			store: AccountStore{
				Active: "alice@example.com",
				Accounts: []AccountEntry{
					{Email: "alice@example.com"},
				},
			},
			email:   "nobody@example.com",
			wantErr: true,
		},
		{
			name: "should adjust active to first remaining when removing active",
			store: AccountStore{
				Active: "alice@example.com",
				Accounts: []AccountEntry{
					{Email: "alice@example.com"},
					{Email: "bob@example.com"},
				},
			},
			email:      "alice@example.com",
			wantActive: "bob@example.com",
			wantCount:  1,
		},
		{
			name: "should clear active when removing last account",
			store: AccountStore{
				Active: "alice@example.com",
				Accounts: []AccountEntry{
					{Email: "alice@example.com"},
				},
			},
			email:      "alice@example.com",
			wantActive: "",
			wantCount:  0,
		},
		{
			name: "should match case-insensitively",
			store: AccountStore{
				Active: "Alice@Example.com",
				Accounts: []AccountEntry{
					{Email: "Alice@Example.com"},
					{Email: "bob@example.com"},
				},
			},
			email:      "alice@example.com",
			wantActive: "bob@example.com",
			wantCount:  1,
		},
		{
			name: "should preserve active when removing inactive account",
			store: AccountStore{
				Active: "alice@example.com",
				Accounts: []AccountEntry{
					{Email: "alice@example.com"},
					{Email: "bob@example.com"},
				},
			},
			email:      "bob@example.com",
			wantActive: "alice@example.com",
			wantCount:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			store := tt.store

			err := store.RemoveAccount(tt.email)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if store.Active != tt.wantActive {
				t.Errorf("active = %q, want %q", store.Active, tt.wantActive)
			}
			if len(store.Accounts) != tt.wantCount {
				t.Errorf("account count = %d, want %d", len(store.Accounts), tt.wantCount)
			}
		})
	}
}

func TestSetActive(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		store      AccountStore
		email      string
		wantErr    bool
		wantActive string
	}{
		{
			name:    "should error when email is empty",
			store:   AccountStore{},
			email:   "",
			wantErr: true,
		},
		{
			name: "should error when account not found",
			store: AccountStore{
				Accounts: []AccountEntry{
					{Email: "alice@example.com"},
				},
			},
			email:   "nobody@example.com",
			wantErr: true,
		},
		{
			name: "should set active successfully",
			store: AccountStore{
				Active: "alice@example.com",
				Accounts: []AccountEntry{
					{Email: "alice@example.com"},
					{Email: "bob@example.com"},
				},
			},
			email:      "bob@example.com",
			wantActive: "bob@example.com",
		},
		{
			name: "should match case-insensitively",
			store: AccountStore{
				Accounts: []AccountEntry{
					{Email: "Alice@Example.com"},
				},
			},
			email:      "alice@example.com",
			wantActive: "alice@example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			store := tt.store

			err := store.SetActive(tt.email)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if store.Active != tt.wantActive {
				t.Errorf("active = %q, want %q", store.Active, tt.wantActive)
			}
		})
	}
}

func TestGetActive(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		store     AccountStore
		wantEmail string
		wantErr   bool
	}{
		{
			name: "should return active when set",
			store: AccountStore{
				Active: "alice@example.com",
				Accounts: []AccountEntry{
					{Email: "alice@example.com"},
					{Email: "bob@example.com"},
				},
			},
			wantEmail: "alice@example.com",
		},
		{
			name: "should fall back to single account",
			store: AccountStore{
				Accounts: []AccountEntry{
					{Email: "only@example.com"},
				},
			},
			wantEmail: "only@example.com",
		},
		{
			name:    "should error when store is empty",
			store:   AccountStore{},
			wantErr: true,
		},
		{
			name: "should error when multiple accounts have no active",
			store: AccountStore{
				Accounts: []AccountEntry{
					{Email: "alice@example.com"},
					{Email: "bob@example.com"},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := tt.store.GetActive()

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.wantEmail {
				t.Errorf("GetActive() = %q, want %q", got, tt.wantEmail)
			}
		})
	}
}

func TestHasAccount(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		store AccountStore
		email string
		want  bool
	}{
		{
			name: "should find exact match",
			store: AccountStore{
				Accounts: []AccountEntry{
					{Email: "alice@example.com"},
				},
			},
			email: "alice@example.com",
			want:  true,
		},
		{
			name: "should match case-insensitively",
			store: AccountStore{
				Accounts: []AccountEntry{
					{Email: "Alice@Example.com"},
				},
			},
			email: "alice@example.com",
			want:  true,
		},
		{
			name: "should return false when not found",
			store: AccountStore{
				Accounts: []AccountEntry{
					{Email: "alice@example.com"},
				},
			},
			email: "nobody@example.com",
			want:  false,
		},
		{
			name:  "should return false when store is empty",
			store: AccountStore{},
			email: "alice@example.com",
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := tt.store.HasAccount(tt.email)
			if got != tt.want {
				t.Errorf("HasAccount(%q) = %v, want %v", tt.email, got, tt.want)
			}
		})
	}
}

func TestList(t *testing.T) {
	t.Parallel()

	t.Run("should return correct entries", func(t *testing.T) {
		t.Parallel()

		now := time.Now().UTC()
		store := AccountStore{
			Accounts: []AccountEntry{
				{Email: "alice@example.com", AddedAt: now},
				{Email: "bob@example.com", AddedAt: now},
			},
		}

		result := store.List()

		if len(result) != 2 {
			t.Fatalf("List() returned %d entries, want 2", len(result))
		}
		if result[0].Email != "alice@example.com" {
			t.Errorf("result[0].Email = %q, want %q", result[0].Email, "alice@example.com")
		}
		if result[1].Email != "bob@example.com" {
			t.Errorf("result[1].Email = %q, want %q", result[1].Email, "bob@example.com")
		}
	})

	t.Run("should return copy that does not affect original", func(t *testing.T) {
		t.Parallel()

		store := AccountStore{
			Accounts: []AccountEntry{
				{Email: "alice@example.com"},
				{Email: "bob@example.com"},
			},
		}

		result := store.List()
		result[0].Email = "modified@example.com"

		if store.Accounts[0].Email != "alice@example.com" {
			t.Errorf("modifying returned slice affected original: got %q, want %q",
				store.Accounts[0].Email, "alice@example.com")
		}
	})
}
