package cmd

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/khang/google-suite-cli/internal/auth"
	"github.com/spf13/cobra"
	"google.golang.org/api/gmail/v1"
)

var (
	// threads list flags
	threadsMaxResults int64
	threadsLabelIDs   string
	threadsQuery      string
)

// threadsCmd represents the threads parent command
var threadsCmd = &cobra.Command{
	Use:   "threads",
	Short: "Manage Gmail conversation threads",
	Long: `Manage Gmail conversation threads.

Threads are collections of messages grouped by Gmail into conversations.
Use subcommands to list threads or view individual thread contents.`,
}

// threadsListCmd represents the threads list subcommand
var threadsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List Gmail conversation threads",
	Long: `List Gmail conversation threads with optional filtering.

Displays thread ID, snippet preview, and message count for each thread.
Use --query for Gmail search syntax (same as web interface).

Examples:
  gsuite threads list
  gsuite threads list -n 20
  gsuite threads list -q "from:alice@example.com"
  gsuite threads list --label-ids "INBOX,UNREAD"`,
	RunE: runThreadsList,
}

// threadsGetCmd represents the threads get subcommand
var threadsGetCmd = &cobra.Command{
	Use:   "get <thread-id>",
	Short: "Get a Gmail thread with all messages",
	Long: `Get a Gmail thread and display all messages in the conversation.

Shows messages in chronological order (oldest first) with headers and body content.

Example:
  gsuite threads get 18d1234567890abc`,
	Args: cobra.ExactArgs(1),
	RunE: runThreadsGet,
}

func init() {
	rootCmd.AddCommand(threadsCmd)
	threadsCmd.AddCommand(threadsListCmd)
	threadsCmd.AddCommand(threadsGetCmd)

	// threads list flags
	threadsListCmd.Flags().Int64VarP(&threadsMaxResults, "max-results", "n", 10, "Maximum number of threads to return (max 500)")
	threadsListCmd.Flags().StringVar(&threadsLabelIDs, "label-ids", "", "Comma-separated list of label IDs to filter by")
	threadsListCmd.Flags().StringVarP(&threadsQuery, "query", "q", "", "Gmail search query (same syntax as web interface)")
}

func runThreadsList(cmd *cobra.Command, args []string) error {
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

	// Build threads list request
	listCall := service.Users.Threads.List("me")

	// Apply max results (cap at 500)
	if threadsMaxResults > 500 {
		threadsMaxResults = 500
	}
	listCall = listCall.MaxResults(threadsMaxResults)

	// Apply label IDs filter
	if threadsLabelIDs != "" {
		labels := strings.Split(threadsLabelIDs, ",")
		for i := range labels {
			labels[i] = strings.TrimSpace(labels[i])
		}
		listCall = listCall.LabelIds(labels...)
	}

	// Apply search query
	if threadsQuery != "" {
		listCall = listCall.Q(threadsQuery)
	}

	// Execute request
	result, err := listCall.Do()
	if err != nil {
		return fmt.Errorf("Gmail API error: %w", err)
	}

	// Print results
	if len(result.Threads) == 0 {
		fmt.Println("No threads found.")
		return nil
	}

	for _, thread := range result.Threads {
		// Get full thread to access message count
		fullThread, err := service.Users.Threads.Get("me", thread.Id).Format("minimal").Do()
		if err != nil {
			// If we can't get full thread, just print what we have
			fmt.Printf("Thread: %s\n", thread.Id)
			fmt.Printf("  Snippet: %s\n", truncateSnippet(thread.Snippet, 80))
			fmt.Println()
			continue
		}

		fmt.Printf("Thread: %s\n", thread.Id)
		fmt.Printf("  Messages: %d\n", len(fullThread.Messages))
		fmt.Printf("  Snippet: %s\n", truncateSnippet(thread.Snippet, 80))
		fmt.Println()
	}

	// Indicate if more results available
	if result.NextPageToken != "" {
		fmt.Println("More results available. Use pagination to see more.")
	}

	return nil
}

func runThreadsGet(cmd *cobra.Command, args []string) error {
	threadID := args[0]

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

	// Get thread with full message details
	thread, err := service.Users.Threads.Get("me", threadID).Format("full").Do()
	if err != nil {
		return fmt.Errorf("Gmail API error: %w", err)
	}

	fmt.Printf("Thread: %s (%d messages)\n", thread.Id, len(thread.Messages))
	fmt.Println(strings.Repeat("=", 60))

	// Display messages in chronological order (oldest first)
	for i, msg := range thread.Messages {
		if i > 0 {
			fmt.Println(strings.Repeat("-", 60))
		}

		// Extract headers
		headers := make(map[string]string)
		if msg.Payload != nil {
			for _, h := range msg.Payload.Headers {
				headers[strings.ToLower(h.Name)] = h.Value
			}
		}

		// Print headers
		if from := headers["from"]; from != "" {
			fmt.Printf("From: %s\n", from)
		}
		if to := headers["to"]; to != "" {
			fmt.Printf("To: %s\n", to)
		}
		if date := headers["date"]; date != "" {
			fmt.Printf("Date: %s\n", date)
		}
		if subject := headers["subject"]; subject != "" {
			fmt.Printf("Subject: %s\n", subject)
		}
		fmt.Println()

		// Extract and print body
		body := extractMessageBody(msg.Payload)
		if body != "" {
			fmt.Println(body)
		} else {
			fmt.Println("[No text content]")
		}
		fmt.Println()
	}

	return nil
}

// extractMessageBody extracts plain text body from message payload
func extractMessageBody(payload *gmail.MessagePart) string {
	if payload == nil {
		return ""
	}

	// If this part has text/plain body, decode and return it
	if payload.MimeType == "text/plain" && payload.Body != nil && payload.Body.Data != "" {
		decoded, err := base64.URLEncoding.DecodeString(payload.Body.Data)
		if err != nil {
			return ""
		}
		return string(decoded)
	}

	// Check for multipart messages
	if strings.HasPrefix(payload.MimeType, "multipart/") && len(payload.Parts) > 0 {
		// First try to find text/plain
		for _, part := range payload.Parts {
			if part.MimeType == "text/plain" {
				body := extractMessageBody(part)
				if body != "" {
					return body
				}
			}
		}
		// Recurse into nested multipart
		for _, part := range payload.Parts {
			body := extractMessageBody(part)
			if body != "" {
				return body
			}
		}
	}

	return ""
}

// truncateSnippet truncates a snippet to maxLen characters
func truncateSnippet(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
