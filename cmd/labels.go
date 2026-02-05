package cmd

import (
	"context"
	"fmt"
	"sort"

	"github.com/khang/google-suite-cli/internal/auth"
	"github.com/spf13/cobra"
	"google.golang.org/api/gmail/v1"
)

// labelsCmd represents the labels parent command
var labelsCmd = &cobra.Command{
	Use:   "labels",
	Short: "Manage Gmail labels",
	Long: `Manage Gmail labels for the authenticated user.

Labels are used to categorize messages in Gmail. This command group
provides operations for listing, creating, updating, and deleting labels.`,
}

// labelsListCmd represents the labels list command
var labelsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all Gmail labels",
	Long: `List all labels for the authenticated user's Gmail account.

Labels are sorted with system labels first (like INBOX, SENT, SPAM),
followed by user-created labels in alphabetical order.

Each label shows:
  - ID: The unique identifier used in API calls
  - Name: The display name of the label
  - Type: Either "system" or "user"`,
	RunE: runLabelsList,
}

func init() {
	rootCmd.AddCommand(labelsCmd)
	labelsCmd.AddCommand(labelsListCmd)
}

func runLabelsList(cmd *cobra.Command, args []string) error {
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
		if credFile == "" {
			return fmt.Errorf("no credentials provided. Use --credentials-file or set GOOGLE_CREDENTIALS env var")
		}
		return fmt.Errorf("authentication failed: %w", err)
	}

	// List labels
	resp, err := service.Users.Labels.List("me").Do()
	if err != nil {
		return fmt.Errorf("Gmail API error: %w", err)
	}

	// Separate system and user labels
	var systemLabels, userLabels []*gmail.Label
	for _, label := range resp.Labels {
		if label.Type == "system" {
			systemLabels = append(systemLabels, label)
		} else {
			userLabels = append(userLabels, label)
		}
	}

	// Sort each group alphabetically by name
	sort.Slice(systemLabels, func(i, j int) bool {
		return systemLabels[i].Name < systemLabels[j].Name
	})
	sort.Slice(userLabels, func(i, j int) bool {
		return userLabels[i].Name < userLabels[j].Name
	})

	// Print header
	fmt.Printf("%-30s %-40s %s\n", "NAME", "ID", "TYPE")
	fmt.Printf("%-30s %-40s %s\n", "----", "--", "----")

	// Print system labels first
	for _, label := range systemLabels {
		fmt.Printf("%-30s %-40s %s\n", label.Name, label.Id, label.Type)
	}

	// Print user labels
	for _, label := range userLabels {
		fmt.Printf("%-30s %-40s %s\n", label.Name, label.Id, label.Type)
	}

	fmt.Printf("\n[Total: %d labels (%d system, %d user)]\n",
		len(resp.Labels), len(systemLabels), len(userLabels))

	return nil
}
