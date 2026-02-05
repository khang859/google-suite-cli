package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/khang/google-suite-cli/internal/auth"
	"github.com/spf13/cobra"
)

var (
	searchMaxResults int64
	searchLabelIDs   string
)

// searchCmd represents the search command
var searchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search Gmail messages using Gmail query syntax",
	Long: `Search Gmail messages using Gmail's powerful query syntax.

Examples:
  gsuite search "from:user@example.com"
  gsuite search "subject:meeting" --max-results 20
  gsuite search "is:unread" --label-ids INBOX
  gsuite search "newer_than:1d"

Query syntax supports operators like:
  from:, to:, subject:, has:attachment, is:unread, is:starred,
  newer_than:, older_than:, label:, in:, and many more.

See https://support.google.com/mail/answer/7190 for full query syntax.`,
	Args: cobra.ExactArgs(1),
	RunE: runSearch,
}

func init() {
	rootCmd.AddCommand(searchCmd)
	searchCmd.Flags().Int64VarP(&searchMaxResults, "max-results", "n", 10, "Maximum number of results (1-500)")
	searchCmd.Flags().StringVar(&searchLabelIDs, "label-ids", "", "Comma-separated label IDs to filter by")
}

func runSearch(cmd *cobra.Command, args []string) error {
	query := args[0]

	// Get credentials file and user email from root flags
	credFile := GetCredentialsFile()
	user := GetUserEmail()

	// Validate user is provided
	if user == "" {
		return fmt.Errorf("--user flag required to specify email to impersonate")
	}

	// Validate max-results
	if searchMaxResults < 1 || searchMaxResults > 500 {
		return fmt.Errorf("--max-results must be between 1 and 500")
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

	// Build the list request
	listReq := service.Users.Messages.List("me").Q(query).MaxResults(searchMaxResults)

	// Add label filter if provided
	if searchLabelIDs != "" {
		labelList := strings.Split(searchLabelIDs, ",")
		listReq = listReq.LabelIds(labelList...)
	}

	// Execute the search
	resp, err := listReq.Do()
	if err != nil {
		return fmt.Errorf("Gmail API error: %w", err)
	}

	// Handle empty results
	if len(resp.Messages) == 0 {
		fmt.Println("No messages found matching query")
		return nil
	}

	// Fetch and display each message
	for i, msg := range resp.Messages {
		// Fetch full message details
		fullMsg, err := service.Users.Messages.Get("me", msg.Id).Format("metadata").MetadataHeaders("From", "Subject", "Date").Do()
		if err != nil {
			return fmt.Errorf("failed to fetch message %s: %w", msg.Id, err)
		}

		// Extract headers
		var from, subject, date string
		for _, header := range fullMsg.Payload.Headers {
			switch header.Name {
			case "From":
				from = header.Value
			case "Subject":
				subject = header.Value
			case "Date":
				date = header.Value
			}
		}

		// Get snippet (truncate to 100 chars)
		snippet := fullMsg.Snippet
		if len(snippet) > 100 {
			snippet = snippet[:100] + "..."
		}

		// Print message info
		if i > 0 {
			fmt.Println("---")
		}
		fmt.Printf("ID: %s\n", fullMsg.Id)
		fmt.Printf("Date: %s\n", date)
		fmt.Printf("From: %s\n", from)
		fmt.Printf("Subject: %s\n", subject)
		fmt.Printf("Snippet: %s\n", snippet)
	}

	fmt.Printf("\n[Showing %d of %d estimated results]\n", len(resp.Messages), resp.ResultSizeEstimate)

	return nil
}
