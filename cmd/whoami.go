package cmd

import (
	"context"
	"fmt"

	"github.com/khang/google-suite-cli/internal/auth"
	"github.com/spf13/cobra"
)

// whoamiCmd represents the whoami command
var whoamiCmd = &cobra.Command{
	Use:   "whoami",
	Short: "Show authenticated user's Gmail profile",
	Long: `Show the Gmail profile of the authenticated user.

This command uses the Gmail API to retrieve profile information for the
user being impersonated via domain-wide delegation. It displays the email
address and message/thread counts.

This is useful for verifying that authentication is working correctly.`,
	RunE: runWhoami,
}

func init() {
	rootCmd.AddCommand(whoamiCmd)
}

func runWhoami(cmd *cobra.Command, args []string) error {
	// Get credentials file and user email from root flags
	credFile := GetCredentialsFile()
	user := GetUserEmail()

	// Validate user is provided
	if user == "" {
		return fmt.Errorf("--user flag required to specify email to impersonate")
	}

	// Create auth config
	cfg := auth.Config{
		CredentialsFile: credFile,
		UserEmail:       user,
	}

	// Create context
	ctx := context.Background()

	// Create Gmail service
	service, err := auth.NewGmailService(ctx, cfg)
	if err != nil {
		// Check if it's a credentials error
		if credFile == "" {
			return fmt.Errorf("no credentials provided. Use --credentials-file or set GOOGLE_CREDENTIALS env var")
		}
		return fmt.Errorf("authentication failed: %w", err)
	}

	// Get user profile
	profile, err := service.Users.GetProfile("me").Do()
	if err != nil {
		return fmt.Errorf("Gmail API error: %w", err)
	}

	// JSON output mode
	if GetOutputFormat() == "json" {
		type whoamiResult struct {
			Email         string `json:"email"`
			MessagesTotal int64  `json:"messages_total"`
			ThreadsTotal  int64  `json:"threads_total"`
		}
		return outputJSON(whoamiResult{
			Email:         profile.EmailAddress,
			MessagesTotal: profile.MessagesTotal,
			ThreadsTotal:  profile.ThreadsTotal,
		})
	}

	// Print profile information (text mode)
	fmt.Printf("Email: %s\n", profile.EmailAddress)
	fmt.Printf("Messages Total: %d\n", profile.MessagesTotal)
	fmt.Printf("Threads Total: %d\n", profile.ThreadsTotal)

	return nil
}
