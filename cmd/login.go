package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/khang/google-suite-cli/internal/auth"
	"github.com/spf13/cobra"
)

// loginCmd represents the login command
var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with Gmail using OAuth2",
	Long: `Authenticate with Gmail using OAuth2.

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
	Use:   "logout",
	Short: "Remove saved OAuth2 token",
	Long: `Remove the locally saved OAuth2 token, effectively logging out.

The token file is stored at ~/.config/gsuite/token.json.
After logout, you will need to run 'gsuite login' again to authenticate.`,
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
	return nil
}

func runLogout(cmd *cobra.Command, args []string) error {
	tokenPath, err := auth.LegacyTokenPath()
	if err != nil {
		return fmt.Errorf("failed to determine token path: %w", err)
	}

	if _, err := os.Stat(tokenPath); os.IsNotExist(err) {
		fmt.Println("Not logged in")
		return nil
	}

	if err := os.Remove(tokenPath); err != nil {
		return fmt.Errorf("failed to remove token: %w", err)
	}

	fmt.Println("Logged out — token removed")
	return nil
}
