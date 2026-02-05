package cmd

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/khang/google-suite-cli/internal/auth"
	"github.com/spf13/cobra"
	"google.golang.org/api/gmail/v1"
)

var (
	// labelsCreateCmd flags
	labelName                 string
	labelListVisibility       string
	messageListVisibility     string
)

// System label IDs that cannot be modified or deleted
var systemLabelIDs = map[string]bool{
	"INBOX":       true,
	"SENT":        true,
	"TRASH":       true,
	"SPAM":        true,
	"DRAFT":       true,
	"STARRED":     true,
	"UNREAD":      true,
	"IMPORTANT":   true,
	"CATEGORY_PERSONAL":    true,
	"CATEGORY_SOCIAL":      true,
	"CATEGORY_PROMOTIONS":  true,
	"CATEGORY_UPDATES":     true,
	"CATEGORY_FORUMS":      true,
}

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

// labelsCreateCmd represents the labels create command
var labelsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new Gmail label",
	Long: `Create a new user label in the authenticated user's Gmail account.

Required flags:
  --name, -n: The display name for the new label

Optional flags:
  --label-list-visibility: Visibility in label list (labelShow, labelShowIfUnread, labelHide)
  --message-list-visibility: Visibility in message list (show, hide)`,
	Example: `  # Create a simple label
  gsuite labels create -n "My Label"

  # Create a label with custom visibility
  gsuite labels create -n "Work" --label-list-visibility labelShow --message-list-visibility show`,
	RunE: runLabelsCreate,
}

// labelsUpdateCmd represents the labels update command
var labelsUpdateCmd = &cobra.Command{
	Use:   "update <label-id>",
	Short: "Update an existing Gmail label",
	Long: `Update a user-created label's properties.

Note: System labels (INBOX, SENT, SPAM, etc.) cannot be updated.

Args:
  label-id: The ID of the label to update (required)

Optional flags:
  --name, -n: New display name for the label
  --label-list-visibility: Visibility in label list (labelShow, labelShowIfUnread, labelHide)
  --message-list-visibility: Visibility in message list (show, hide)`,
	Example: `  # Rename a label
  gsuite labels update Label_123 -n "New Name"

  # Update visibility settings
  gsuite labels update Label_123 --label-list-visibility labelHide`,
	Args: cobra.ExactArgs(1),
	RunE: runLabelsUpdate,
}

// labelsDeleteCmd represents the labels delete command
var labelsDeleteCmd = &cobra.Command{
	Use:   "delete <label-id>",
	Short: "Delete a Gmail label",
	Long: `Delete a user-created label from the authenticated user's Gmail account.

Note: System labels (INBOX, SENT, SPAM, etc.) cannot be deleted.
Messages that have this label will not be deleted, only the label will be removed from them.

Args:
  label-id: The ID of the label to delete (required)`,
	Example: `  # Delete a label
  gsuite labels delete Label_123`,
	Args: cobra.ExactArgs(1),
	RunE: runLabelsDelete,
}

func init() {
	rootCmd.AddCommand(labelsCmd)
	labelsCmd.AddCommand(labelsListCmd)
	labelsCmd.AddCommand(labelsCreateCmd)
	labelsCmd.AddCommand(labelsUpdateCmd)
	labelsCmd.AddCommand(labelsDeleteCmd)

	// labelsCreateCmd flags
	labelsCreateCmd.Flags().StringVarP(&labelName, "name", "n", "", "Name for the new label (required)")
	labelsCreateCmd.Flags().StringVar(&labelListVisibility, "label-list-visibility", "", "Label list visibility (labelShow, labelShowIfUnread, labelHide)")
	labelsCreateCmd.Flags().StringVar(&messageListVisibility, "message-list-visibility", "", "Message list visibility (show, hide)")
	labelsCreateCmd.MarkFlagRequired("name")

	// labelsUpdateCmd flags (reuses the same flag variables)
	labelsUpdateCmd.Flags().StringVarP(&labelName, "name", "n", "", "New name for the label")
	labelsUpdateCmd.Flags().StringVar(&labelListVisibility, "label-list-visibility", "", "Label list visibility (labelShow, labelShowIfUnread, labelHide)")
	labelsUpdateCmd.Flags().StringVar(&messageListVisibility, "message-list-visibility", "", "Message list visibility (show, hide)")
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

func runLabelsCreate(cmd *cobra.Command, args []string) error {
	// Get credentials file and user email from root flags
	credFile := GetCredentialsFile()
	user := GetUserEmail()

	// Validate user is provided
	if user == "" {
		return fmt.Errorf("--user flag required to specify email to impersonate")
	}

	// Validate name is provided
	if labelName == "" {
		return fmt.Errorf("--name flag is required")
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

	// Build the label object
	label := &gmail.Label{
		Name: labelName,
	}

	// Set optional visibility settings
	if labelListVisibility != "" {
		label.LabelListVisibility = labelListVisibility
	}
	if messageListVisibility != "" {
		label.MessageListVisibility = messageListVisibility
	}

	// Create the label
	created, err := service.Users.Labels.Create("me", label).Do()
	if err != nil {
		return fmt.Errorf("Gmail API error: %w", err)
	}

	fmt.Printf("Label created: %s (%s)\n", created.Id, created.Name)
	return nil
}

func runLabelsUpdate(cmd *cobra.Command, args []string) error {
	labelID := args[0]

	// Check if it's a system label
	if systemLabelIDs[labelID] || strings.HasPrefix(labelID, "CATEGORY_") {
		return fmt.Errorf("cannot modify system label: %s", labelID)
	}

	// Get credentials file and user email from root flags
	credFile := GetCredentialsFile()
	user := GetUserEmail()

	// Validate user is provided
	if user == "" {
		return fmt.Errorf("--user flag required to specify email to impersonate")
	}

	// Check if at least one update flag is provided
	if labelName == "" && labelListVisibility == "" && messageListVisibility == "" {
		return fmt.Errorf("at least one of --name, --label-list-visibility, or --message-list-visibility is required")
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

	// First, get the existing label to check if it exists
	existing, err := service.Users.Labels.Get("me", labelID).Do()
	if err != nil {
		if strings.Contains(err.Error(), "404") || strings.Contains(err.Error(), "Not Found") {
			return fmt.Errorf("label not found: %s", labelID)
		}
		return fmt.Errorf("Gmail API error: %w", err)
	}

	// Check if it's a system label by type
	if existing.Type == "system" {
		return fmt.Errorf("cannot modify system label: %s", labelID)
	}

	// Build the label update object
	label := &gmail.Label{}

	if labelName != "" {
		label.Name = labelName
	} else {
		label.Name = existing.Name // Keep existing name
	}
	if labelListVisibility != "" {
		label.LabelListVisibility = labelListVisibility
	}
	if messageListVisibility != "" {
		label.MessageListVisibility = messageListVisibility
	}

	// Update the label
	_, err = service.Users.Labels.Update("me", labelID, label).Do()
	if err != nil {
		return fmt.Errorf("Gmail API error: %w", err)
	}

	fmt.Printf("Label updated: %s\n", labelID)
	return nil
}

func runLabelsDelete(cmd *cobra.Command, args []string) error {
	labelID := args[0]

	// Check if it's a system label
	if systemLabelIDs[labelID] || strings.HasPrefix(labelID, "CATEGORY_") {
		return fmt.Errorf("cannot delete system label: %s", labelID)
	}

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

	// First, verify the label exists and check if it's a system label
	existing, err := service.Users.Labels.Get("me", labelID).Do()
	if err != nil {
		if strings.Contains(err.Error(), "404") || strings.Contains(err.Error(), "Not Found") {
			return fmt.Errorf("label not found: %s", labelID)
		}
		return fmt.Errorf("Gmail API error: %w", err)
	}

	// Check if it's a system label by type
	if existing.Type == "system" {
		return fmt.Errorf("cannot delete system label: %s", labelID)
	}

	// Delete the label
	err = service.Users.Labels.Delete("me", labelID).Do()
	if err != nil {
		return fmt.Errorf("Gmail API error: %w", err)
	}

	fmt.Printf("Label deleted: %s\n", labelID)
	return nil
}
