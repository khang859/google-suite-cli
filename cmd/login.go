package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/khang/google-suite-cli/internal/auth"
	"github.com/spf13/cobra"
)

// noBrowser controls whether to use the device authorization flow instead of the browser flow.
var noBrowser bool

// loginCmd represents the login command
var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with Gmail using OAuth2",
	Long: `Authenticate with Gmail using OAuth2.

By default, this command opens your browser to complete the Google OAuth2 consent
flow. Use --no-browser for headless environments (SSH, EC2, containers) which
uses the RFC 8628 device authorization flow instead — it displays a URL and code
that you can enter on any device with a browser.

After authentication, a token is saved locally so subsequent commands work
without needing to log in again.

Requires OAuth2 client credentials (not a service account). Provide credentials
via --credentials-file flag or GOOGLE_CREDENTIALS / GOOGLE_APPLICATION_CREDENTIALS
environment variables.`,
	Example: `  # Login with default credentials (opens browser)
  gsuite login

  # Login with a specific credentials file
  gsuite login -c /path/to/oauth2-client.json

  # Login on headless environments (SSH, EC2, containers)
  gsuite login --no-browser`,
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
	loginCmd.Flags().BoolVar(&noBrowser, "no-browser", false, "Use device authorization flow for headless environments")
	rootCmd.AddCommand(loginCmd)
	rootCmd.AddCommand(logoutCmd)
}

func runLogin(cmd *cobra.Command, args []string) error {
	credFile := GetCredentialsFile()

	cfg := auth.Config{
		CredentialsFile: credFile,
	}

	credJSON, err := auth.LoadCredentials(cfg)
	if err != nil {
		return fmt.Errorf("no credentials found: %w", err)
	}

	ctx := context.Background()

	email, err := auth.Login(ctx, credJSON, noBrowser)
	if err != nil {
		return err
	}

	fmt.Printf("Logged in as %s\n", email)
	return nil
}

func runLogout(cmd *cobra.Command, args []string) error {
	tokenPath, err := auth.TokenPath()
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
