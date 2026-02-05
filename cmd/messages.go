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
	// messagesListCmd flags
	maxResults int64
	labelIDs   string
	query      string

	// messagesModifyCmd flags
	addLabels    string
	removeLabels string
)

// messagesCmd represents the messages command group
var messagesCmd = &cobra.Command{
	Use:   "messages",
	Short: "Manage Gmail messages",
	Long: `Commands for listing, reading, and managing Gmail messages.

Use the subcommands to interact with messages in the authenticated user's mailbox.`,
}

// messagesListCmd represents the messages list command
var messagesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List messages in the mailbox",
	Long: `List messages in the authenticated user's Gmail mailbox.

Supports filtering by labels, search query, and limiting results.
Returns message ID, thread ID, and snippet for each message.`,
	Example: `  # List last 10 messages
  gsuite messages list

  # List 50 messages from inbox
  gsuite messages list -n 50 --label-ids INBOX

  # Search for messages
  gsuite messages list -q "from:example@gmail.com subject:important"

  # List unread inbox messages
  gsuite messages list --label-ids INBOX,UNREAD`,
	RunE: runMessagesList,
}

// messagesGetCmd represents the messages get command
var messagesGetCmd = &cobra.Command{
	Use:   "get <message-id>",
	Short: "Get a specific message",
	Long: `Retrieve and display a specific Gmail message by its ID.

Displays the message headers (From, To, Subject, Date) and body content.
Prefers plain text body, falls back to snippet if not available.`,
	Example: `  # Get a specific message
  gsuite messages get 18d5a1b2c3d4e5f6`,
	Args: cobra.ExactArgs(1),
	RunE: runMessagesGet,
}

// messagesModifyCmd represents the messages modify command
var messagesModifyCmd = &cobra.Command{
	Use:   "modify <message-id>",
	Short: "Modify labels on a message",
	Long: `Add or remove labels from a Gmail message.

At least one of --add-labels or --remove-labels is required.
Labels can be system labels (INBOX, UNREAD, STARRED, etc.) or user-created label IDs.`,
	Example: `  # Mark message as read (remove UNREAD label)
  gsuite messages modify <id> --remove-labels UNREAD

  # Archive message (remove INBOX label)
  gsuite messages modify <id> --remove-labels INBOX

  # Add custom label
  gsuite messages modify <id> --add-labels Label_123

  # Star a message
  gsuite messages modify <id> --add-labels STARRED

  # Add and remove labels at the same time
  gsuite messages modify <id> --add-labels STARRED --remove-labels UNREAD

  # Add multiple labels (comma-separated)
  gsuite messages modify <id> --add-labels Label_1,Label_2,STARRED`,
	Args: cobra.ExactArgs(1),
	RunE: runMessagesModify,
}

func init() {
	rootCmd.AddCommand(messagesCmd)
	messagesCmd.AddCommand(messagesListCmd)
	messagesCmd.AddCommand(messagesGetCmd)
	messagesCmd.AddCommand(messagesModifyCmd)

	// messagesListCmd flags
	messagesListCmd.Flags().Int64VarP(&maxResults, "max-results", "n", 10, "Maximum number of messages to return (max 500)")
	messagesListCmd.Flags().StringVar(&labelIDs, "label-ids", "", "Comma-separated label IDs to filter by (e.g., INBOX,UNREAD)")
	messagesListCmd.Flags().StringVarP(&query, "query", "q", "", "Gmail search query string")

	// messagesModifyCmd flags
	messagesModifyCmd.Flags().StringVar(&addLabels, "add-labels", "", "Comma-separated label IDs to add (e.g., STARRED,Label_123)")
	messagesModifyCmd.Flags().StringVar(&removeLabels, "remove-labels", "", "Comma-separated label IDs to remove (e.g., UNREAD,INBOX)")
}

func runMessagesList(cmd *cobra.Command, args []string) error {
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

	// Build the list request
	listCall := service.Users.Messages.List("me")

	// Apply max results (cap at 500)
	if maxResults > 500 {
		maxResults = 500
	}
	listCall.MaxResults(maxResults)

	// Apply label filter if provided
	if labelIDs != "" {
		labels := strings.Split(labelIDs, ",")
		listCall.LabelIds(labels...)
	}

	// Apply search query if provided
	if query != "" {
		listCall.Q(query)
	}

	// Execute the request
	resp, err := listCall.Do()
	if err != nil {
		return fmt.Errorf("Gmail API error: %w", err)
	}

	// Check if no messages found
	if len(resp.Messages) == 0 {
		fmt.Println("No messages found.")
		return nil
	}

	// Print results
	fmt.Printf("Messages (%d):\n\n", len(resp.Messages))
	for _, msg := range resp.Messages {
		// Get message details for snippet
		detail, err := service.Users.Messages.Get("me", msg.Id).Format("metadata").Do()
		if err != nil {
			fmt.Printf("ID: %s  Thread: %s  (error fetching details)\n", msg.Id, msg.ThreadId)
			continue
		}
		snippet := detail.Snippet
		if len(snippet) > 80 {
			snippet = snippet[:80] + "..."
		}
		fmt.Printf("ID: %s\nThread: %s\nSnippet: %s\n\n", msg.Id, msg.ThreadId, snippet)
	}

	// Indicate if more results are available
	if resp.NextPageToken != "" {
		fmt.Println("More results available (pagination token exists)")
	}

	return nil
}

func runMessagesGet(cmd *cobra.Command, args []string) error {
	messageID := args[0]

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

	// Get the message with full format
	msg, err := service.Users.Messages.Get("me", messageID).Format("full").Do()
	if err != nil {
		return fmt.Errorf("Gmail API error: %w", err)
	}

	// Extract headers
	var from, to, subject, date string
	for _, header := range msg.Payload.Headers {
		switch header.Name {
		case "From":
			from = header.Value
		case "To":
			to = header.Value
		case "Subject":
			subject = header.Value
		case "Date":
			date = header.Value
		}
	}

	// Print headers
	fmt.Printf("From: %s\n", from)
	fmt.Printf("To: %s\n", to)
	fmt.Printf("Subject: %s\n", subject)
	fmt.Printf("Date: %s\n", date)
	fmt.Println("---")

	// Extract body content
	body := extractBody(msg)
	if body != "" {
		fmt.Println(body)
	} else {
		// Fallback to snippet
		fmt.Printf("(Snippet) %s\n", msg.Snippet)
	}

	return nil
}

// extractBody extracts the plain text body from a message.
// It recursively searches through MIME parts, preferring text/plain.
func extractBody(msg *gmail.Message) string {
	if msg.Payload == nil {
		return ""
	}

	// Check if the payload itself has body data
	if msg.Payload.MimeType == "text/plain" && msg.Payload.Body != nil && msg.Payload.Body.Data != "" {
		return decodeBase64URL(msg.Payload.Body.Data)
	}

	// Search through parts
	return findPlainTextPart(msg.Payload.Parts)
}

// findPlainTextPart recursively searches for text/plain content in MIME parts.
func findPlainTextPart(parts []*gmail.MessagePart) string {
	for _, part := range parts {
		if part.MimeType == "text/plain" && part.Body != nil && part.Body.Data != "" {
			return decodeBase64URL(part.Body.Data)
		}
		// Recurse into nested parts (for multipart messages)
		if len(part.Parts) > 0 {
			if content := findPlainTextPart(part.Parts); content != "" {
				return content
			}
		}
	}
	return ""
}

// decodeBase64URL decodes a base64url-encoded string.
func decodeBase64URL(encoded string) string {
	decoded, err := base64.URLEncoding.DecodeString(encoded)
	if err != nil {
		// Try with padding
		decoded, err = base64.RawURLEncoding.DecodeString(encoded)
		if err != nil {
			return ""
		}
	}
	return string(decoded)
}

func runMessagesModify(cmd *cobra.Command, args []string) error {
	messageID := args[0]

	// Validate that at least one label flag is provided
	if addLabels == "" && removeLabels == "" {
		return fmt.Errorf("at least one of --add-labels or --remove-labels required")
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

	// Build the modify request
	modifyReq := &gmail.ModifyMessageRequest{}

	// Parse and set labels to add
	var addLabelsList []string
	if addLabels != "" {
		addLabelsList = strings.Split(addLabels, ",")
		// Trim whitespace from each label
		for i, label := range addLabelsList {
			addLabelsList[i] = strings.TrimSpace(label)
		}
		modifyReq.AddLabelIds = addLabelsList
	}

	// Parse and set labels to remove
	var removeLabelsList []string
	if removeLabels != "" {
		removeLabelsList = strings.Split(removeLabels, ",")
		// Trim whitespace from each label
		for i, label := range removeLabelsList {
			removeLabelsList[i] = strings.TrimSpace(label)
		}
		modifyReq.RemoveLabelIds = removeLabelsList
	}

	// Execute the modify request
	_, err = service.Users.Messages.Modify("me", messageID, modifyReq).Do()
	if err != nil {
		if strings.Contains(err.Error(), "404") || strings.Contains(err.Error(), "Not Found") {
			return fmt.Errorf("message not found: %s", messageID)
		}
		if strings.Contains(err.Error(), "Invalid label") {
			return fmt.Errorf("invalid label ID in request: %w", err)
		}
		return fmt.Errorf("Gmail API error: %w", err)
	}

	// Print success message with details
	fmt.Printf("Message modified: %s\n", messageID)
	if len(addLabelsList) > 0 {
		fmt.Printf("  Labels added: %s\n", strings.Join(addLabelsList, ", "))
	}
	if len(removeLabelsList) > 0 {
		fmt.Printf("  Labels removed: %s\n", strings.Join(removeLabelsList, ", "))
	}

	return nil
}
