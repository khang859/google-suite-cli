package cmd

import (
	"context"
	"fmt"

	"github.com/khang/google-suite-cli/internal/auth"
	"github.com/spf13/cobra"
)

// loginCmd represents the login command
var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with Gmail using OAuth2",
	Long: `Authenticate with Gmail using OAuth2.

You can login with multiple accounts. The most recently logged-in account
becomes the active account. Use 'gsuite accounts list' to see all accounts
and 'gsuite accounts switch' to change the active account.

This command opens your browser to complete the Google OAuth2 consent flow
using PKCE for security. After authentication, a token is saved locally so
subsequent commands work without needing to log in again.

Requires OAuth2 client credentials via GOOGLE_CREDENTIALS env var (raw JSON)
or GOOGLE_APPLICATION_CREDENTIALS env var (file path).`,
	Example: `  # Login (opens browser)
  gsuite login`,
	RunE: runLogin,
}

// logoutCmd represents the logout command
var logoutCmd = &cobra.Command{
	Use:   "logout [email]",
	Short: "Remove saved OAuth2 token",
	Long: `Remove an authenticated account and its stored token.

If an email argument is provided, that specific account is logged out.
If no argument is provided, the currently active account is logged out.

After logout, if other accounts remain, the next available account
becomes active. Use 'gsuite accounts list' to see remaining accounts.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runLogout,
}

func init() {
	rootCmd.AddCommand(loginCmd)
	rootCmd.AddCommand(logoutCmd)
}

func runLogin(cmd *cobra.Command, args []string) error {
	credJSON, err := auth.LoadCredentials()
	if err != nil {
		return fmt.Errorf("no credentials found: %w", err)
	}

	ctx := context.Background()

	email, err := auth.Login(ctx, credJSON)
	if err != nil {
		return err
	}

	fmt.Printf("Logged in as %s\n", email)

	store, err := auth.LoadAccountStore()
	if err == nil && len(store.List()) > 1 {
		fmt.Printf("Active account set to %s. Use 'gsuite accounts list' to see all accounts.\n", email)
	}

	return nil
}

func runLogout(cmd *cobra.Command, args []string) error {
	store, err := auth.LoadAccountStore()
	if err != nil {
		return fmt.Errorf("failed to load account store: %w", err)
	}

	var email string
	if len(args) > 0 {
		email = args[0]
	} else {
		email, err = store.GetActive()
		if err != nil {
			fmt.Println("Not logged in")
			return nil
		}
	}

	if err := store.RemoveAccount(email); err != nil {
		return err
	}

	if err := store.Save(); err != nil {
		return fmt.Errorf("failed to save account store: %w", err)
	}

	if err := auth.DeleteTokenFor(email); err != nil {
		return fmt.Errorf("failed to remove token: %w", err)
	}

	fmt.Printf("Logged out of %s\n", email)

	if len(store.List()) > 0 {
		fmt.Printf("Active account: %s\n", store.Active)
	} else {
		fmt.Println("No remaining accounts")
	}

	return nil
}
