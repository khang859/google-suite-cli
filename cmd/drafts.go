package cmd

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/khang/google-suite-cli/internal/auth"
	"github.com/spf13/cobra"
	"google.golang.org/api/gmail/v1"
)

var (
	// draftsListCmd flags
	draftsMaxResults int64
)

// draftsCmd represents the drafts command group
var draftsCmd = &cobra.Command{
	Use:   "drafts",
	Short: "Manage Gmail drafts",
	Long: `Commands for managing Gmail drafts.

Use the subcommands to list, view, create, update, send, and delete drafts
in the authenticated user's mailbox.`,
}

// draftsListCmd represents the drafts list command
var draftsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List drafts in the mailbox",
	Long: `List drafts in the authenticated user's Gmail mailbox.

Returns draft ID, message ID, subject, and snippet for each draft.`,
	Example: `  # List last 10 drafts
  gsuite drafts list

  # List up to 50 drafts
  gsuite drafts list -n 50`,
	RunE: runDraftsList,
}

// draftsGetCmd represents the drafts get command
var draftsGetCmd = &cobra.Command{
	Use:   "get <draft-id>",
	Short: "Get a specific draft",
	Long: `Retrieve and display a specific Gmail draft by its ID.

Displays the draft headers (To, Subject, Date) and body content.`,
	Example: `  # Get a specific draft
  gsuite drafts get r1234567890123456789`,
	Args: cobra.ExactArgs(1),
	RunE: runDraftsGet,
}

func init() {
	rootCmd.AddCommand(draftsCmd)
	draftsCmd.AddCommand(draftsListCmd)
	draftsCmd.AddCommand(draftsGetCmd)

	// draftsListCmd flags
	draftsListCmd.Flags().Int64VarP(&draftsMaxResults, "max-results", "n", 10, "Maximum number of drafts to return (max 500)")
}

func runDraftsList(cmd *cobra.Command, args []string) error {
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
	listCall := service.Users.Drafts.List("me")

	// Apply max results (cap at 500)
	if draftsMaxResults > 500 {
		draftsMaxResults = 500
	}
	listCall.MaxResults(draftsMaxResults)

	// Execute the request
	resp, err := listCall.Do()
	if err != nil {
		return fmt.Errorf("Gmail API error: %w", err)
	}

	// Check if no drafts found
	if len(resp.Drafts) == 0 {
		fmt.Println("No drafts found.")
		return nil
	}

	// Print results
	fmt.Printf("Drafts (%d):\n\n", len(resp.Drafts))
	for _, draft := range resp.Drafts {
		// Get draft details
		detail, err := service.Users.Drafts.Get("me", draft.Id).Format("metadata").Do()
		if err != nil {
			fmt.Printf("Draft ID: %s  (error fetching details)\n", draft.Id)
			continue
		}

		// Extract subject from headers
		var subject string
		if detail.Message != nil && detail.Message.Payload != nil {
			for _, header := range detail.Message.Payload.Headers {
				if header.Name == "Subject" {
					subject = header.Value
					break
				}
			}
		}

		// Get snippet and truncate to 60 chars
		snippet := ""
		if detail.Message != nil {
			snippet = detail.Message.Snippet
			if len(snippet) > 60 {
				snippet = snippet[:60] + "..."
			}
		}

		// Get message ID
		messageID := ""
		if detail.Message != nil {
			messageID = detail.Message.Id
		}

		fmt.Printf("Draft ID: %s\n", draft.Id)
		fmt.Printf("Message ID: %s\n", messageID)
		fmt.Printf("Subject: %s\n", subject)
		fmt.Printf("Snippet: %s\n\n", snippet)
	}

	// Indicate if more results are available
	if resp.NextPageToken != "" {
		fmt.Println("More results available (pagination token exists)")
	}

	return nil
}

func runDraftsGet(cmd *cobra.Command, args []string) error {
	draftID := args[0]

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

	// Get the draft with full format
	draft, err := service.Users.Drafts.Get("me", draftID).Format("full").Do()
	if err != nil {
		return fmt.Errorf("Gmail API error: %w", err)
	}

	if draft.Message == nil || draft.Message.Payload == nil {
		return fmt.Errorf("draft not found: %s", draftID)
	}

	// Extract headers
	var to, subject, date string
	for _, header := range draft.Message.Payload.Headers {
		switch header.Name {
		case "To":
			to = header.Value
		case "Subject":
			subject = header.Value
		case "Date":
			date = header.Value
		}
	}

	// Print headers
	fmt.Printf("Draft ID: %s\n", draftID)
	fmt.Printf("To: %s\n", to)
	fmt.Printf("Subject: %s\n", subject)
	fmt.Printf("Date: %s\n", date)
	fmt.Println("---")

	// Extract body content (reuse pattern from messages.go)
	body := extractDraftBody(draft.Message)
	if body != "" {
		fmt.Println(body)
	} else {
		// Fallback to snippet
		fmt.Printf("(Snippet) %s\n", draft.Message.Snippet)
	}

	return nil
}

// extractDraftBody extracts the plain text body from a draft message.
// It recursively searches through MIME parts, preferring text/plain.
func extractDraftBody(msg *gmail.Message) string {
	if msg.Payload == nil {
		return ""
	}

	// Check if the payload itself has body data
	if msg.Payload.MimeType == "text/plain" && msg.Payload.Body != nil && msg.Payload.Body.Data != "" {
		return decodeDraftBase64URL(msg.Payload.Body.Data)
	}

	// Search through parts
	return findDraftPlainTextPart(msg.Payload.Parts)
}

// findDraftPlainTextPart recursively searches for text/plain content in MIME parts.
func findDraftPlainTextPart(parts []*gmail.MessagePart) string {
	for _, part := range parts {
		if part.MimeType == "text/plain" && part.Body != nil && part.Body.Data != "" {
			return decodeDraftBase64URL(part.Body.Data)
		}
		// Recurse into nested parts (for multipart messages)
		if len(part.Parts) > 0 {
			if content := findDraftPlainTextPart(part.Parts); content != "" {
				return content
			}
		}
	}
	return ""
}

// decodeDraftBase64URL decodes a base64url-encoded string.
func decodeDraftBase64URL(encoded string) string {
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

