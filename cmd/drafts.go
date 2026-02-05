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

	// draftsCreateCmd flags
	draftTo      string
	draftSubject string
	draftBody    string
	draftCc      string
	draftBcc     string
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

// draftsCreateCmd represents the drafts create command
var draftsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new draft",
	Long: `Create a new email draft in the authenticated user's Gmail account.

Required flags:
  --to, -t: Recipient email address
  --subject, -s: Email subject
  --body, -b: Plain text body content

Optional flags:
  --cc: CC recipients (comma-separated)
  --bcc: BCC recipients (comma-separated)`,
	Example: `  # Create a simple draft
  gsuite drafts create --to "user@example.com" --subject "Hello" --body "Draft content"

  # Create a draft with CC and BCC
  gsuite drafts create -t "user@example.com" -s "Meeting" -b "Let's meet" --cc "cc@example.com"`,
	RunE: runDraftsCreate,
}

// draftsUpdateCmd represents the drafts update command
var draftsUpdateCmd = &cobra.Command{
	Use:   "update <draft-id>",
	Short: "Update an existing draft",
	Long: `Update an existing draft with new content.

Args:
  draft-id: The ID of the draft to update (required)

Optional flags (at least one required):
  --to, -t: New recipient email address
  --subject, -s: New email subject
  --body, -b: New plain text body content
  --cc: New CC recipients (comma-separated)
  --bcc: New BCC recipients (comma-separated)

Note: If a field is not provided, the existing value is preserved.`,
	Example: `  # Update draft subject
  gsuite drafts update r1234567890 --subject "Updated Subject"

  # Update multiple fields
  gsuite drafts update r1234567890 --to "new@example.com" --body "New content"`,
	Args: cobra.ExactArgs(1),
	RunE: runDraftsUpdate,
}

// draftsSendCmd represents the drafts send command
var draftsSendCmd = &cobra.Command{
	Use:   "send <draft-id>",
	Short: "Send an existing draft",
	Long: `Send an existing draft as an email message.

Args:
  draft-id: The ID of the draft to send (required)

The draft will be removed from the drafts folder after sending.`,
	Example: `  # Send a draft
  gsuite drafts send r1234567890123456789`,
	Args: cobra.ExactArgs(1),
	RunE: runDraftsSend,
}

// draftsDeleteCmd represents the drafts delete command
var draftsDeleteCmd = &cobra.Command{
	Use:   "delete <draft-id>",
	Short: "Delete a draft",
	Long: `Permanently delete a draft from the authenticated user's Gmail account.

Args:
  draft-id: The ID of the draft to delete (required)

Warning: This action cannot be undone.`,
	Example: `  # Delete a draft
  gsuite drafts delete r1234567890123456789`,
	Args: cobra.ExactArgs(1),
	RunE: runDraftsDelete,
}

func init() {
	rootCmd.AddCommand(draftsCmd)
	draftsCmd.AddCommand(draftsListCmd)
	draftsCmd.AddCommand(draftsGetCmd)
	draftsCmd.AddCommand(draftsCreateCmd)
	draftsCmd.AddCommand(draftsUpdateCmd)
	draftsCmd.AddCommand(draftsSendCmd)
	draftsCmd.AddCommand(draftsDeleteCmd)

	// draftsListCmd flags
	draftsListCmd.Flags().Int64VarP(&draftsMaxResults, "max-results", "n", 10, "Maximum number of drafts to return (max 500)")

	// draftsCreateCmd flags
	draftsCreateCmd.Flags().StringVarP(&draftTo, "to", "t", "", "Recipient email address (required)")
	draftsCreateCmd.Flags().StringVarP(&draftSubject, "subject", "s", "", "Email subject (required)")
	draftsCreateCmd.Flags().StringVarP(&draftBody, "body", "b", "", "Plain text body content (required)")
	draftsCreateCmd.Flags().StringVar(&draftCc, "cc", "", "CC recipients (comma-separated)")
	draftsCreateCmd.Flags().StringVar(&draftBcc, "bcc", "", "BCC recipients (comma-separated)")
	draftsCreateCmd.MarkFlagRequired("to")
	draftsCreateCmd.MarkFlagRequired("subject")
	draftsCreateCmd.MarkFlagRequired("body")

	// draftsUpdateCmd flags (reuses same flag variables)
	draftsUpdateCmd.Flags().StringVarP(&draftTo, "to", "t", "", "New recipient email address")
	draftsUpdateCmd.Flags().StringVarP(&draftSubject, "subject", "s", "", "New email subject")
	draftsUpdateCmd.Flags().StringVarP(&draftBody, "body", "b", "", "New plain text body content")
	draftsUpdateCmd.Flags().StringVar(&draftCc, "cc", "", "New CC recipients (comma-separated)")
	draftsUpdateCmd.Flags().StringVar(&draftBcc, "bcc", "", "New BCC recipients (comma-separated)")
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
		if GetOutputFormat() == "json" {
			return outputJSON([]struct{}{})
		}
		fmt.Println("No drafts found.")
		return nil
	}

	// JSON output mode
	if GetOutputFormat() == "json" {
		type draftListItem struct {
			DraftID   string `json:"draft_id"`
			MessageID string `json:"message_id"`
			Subject   string `json:"subject"`
			Snippet   string `json:"snippet"`
		}
		var results []draftListItem
		for _, draft := range resp.Drafts {
			detail, err := service.Users.Drafts.Get("me", draft.Id).Format("metadata").Do()
			if err != nil {
				results = append(results, draftListItem{DraftID: draft.Id})
				continue
			}
			var subject, snippet, messageID string
			if detail.Message != nil && detail.Message.Payload != nil {
				for _, header := range detail.Message.Payload.Headers {
					if header.Name == "Subject" {
						subject = header.Value
						break
					}
				}
			}
			if detail.Message != nil {
				snippet = detail.Message.Snippet
				messageID = detail.Message.Id
			}
			results = append(results, draftListItem{
				DraftID:   draft.Id,
				MessageID: messageID,
				Subject:   subject,
				Snippet:   snippet,
			})
		}
		return outputJSON(results)
	}

	// Print results (text mode)
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

	// Extract body content
	body := extractDraftBody(draft.Message)

	// JSON output mode
	if GetOutputFormat() == "json" {
		type draftGetResult struct {
			DraftID string `json:"draft_id"`
			To      string `json:"to"`
			Subject string `json:"subject"`
			Date    string `json:"date"`
			Body    string `json:"body"`
		}
		bodyText := body
		if bodyText == "" {
			bodyText = draft.Message.Snippet
		}
		return outputJSON(draftGetResult{
			DraftID: draftID,
			To:      to,
			Subject: subject,
			Date:    date,
			Body:    bodyText,
		})
	}

	// Print headers (text mode)
	fmt.Printf("Draft ID: %s\n", draftID)
	fmt.Printf("To: %s\n", to)
	fmt.Printf("Subject: %s\n", subject)
	fmt.Printf("Date: %s\n", date)
	fmt.Println("---")

	// Print body content
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

func runDraftsCreate(cmd *cobra.Command, args []string) error {
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

	// Build RFC 2822 formatted message
	message := buildRFC2822Message(draftTo, draftSubject, draftBody, draftCc, draftBcc)

	// Base64url encode the message
	encodedMessage := base64.URLEncoding.EncodeToString([]byte(message))

	// Create the draft
	draft := &gmail.Draft{
		Message: &gmail.Message{
			Raw: encodedMessage,
		},
	}

	created, err := service.Users.Drafts.Create("me", draft).Do()
	if err != nil {
		return fmt.Errorf("Gmail API error: %w", err)
	}

	// JSON output mode
	if GetOutputFormat() == "json" {
		type draftCreateResult struct {
			DraftID   string `json:"draft_id"`
			MessageID string `json:"message_id"`
		}
		msgID := ""
		if created.Message != nil {
			msgID = created.Message.Id
		}
		return outputJSON(draftCreateResult{
			DraftID:   created.Id,
			MessageID: msgID,
		})
	}

	fmt.Printf("Draft created: %s\n", created.Id)
	return nil
}

func runDraftsUpdate(cmd *cobra.Command, args []string) error {
	draftID := args[0]

	// Get credentials file and user email from root flags
	credFile := GetCredentialsFile()
	user := GetUserEmail()

	// Validate user is provided
	if user == "" {
		return fmt.Errorf("--user flag required to specify email to impersonate")
	}

	// Check if at least one field is provided
	if draftTo == "" && draftSubject == "" && draftBody == "" && draftCc == "" && draftBcc == "" {
		return fmt.Errorf("at least one of --to, --subject, --body, --cc, or --bcc is required")
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

	// Get existing draft to preserve unmodified fields
	existing, err := service.Users.Drafts.Get("me", draftID).Format("full").Do()
	if err != nil {
		return fmt.Errorf("draft not found: %s", draftID)
	}

	// Extract existing values from headers
	var existingTo, existingSubject, existingCc, existingBcc string
	if existing.Message != nil && existing.Message.Payload != nil {
		for _, header := range existing.Message.Payload.Headers {
			switch header.Name {
			case "To":
				existingTo = header.Value
			case "Subject":
				existingSubject = header.Value
			case "Cc":
				existingCc = header.Value
			case "Bcc":
				existingBcc = header.Value
			}
		}
	}

	// Extract existing body
	existingBody := extractDraftBody(existing.Message)

	// Use new values if provided, otherwise keep existing
	finalTo := existingTo
	if draftTo != "" {
		finalTo = draftTo
	}
	finalSubject := existingSubject
	if draftSubject != "" {
		finalSubject = draftSubject
	}
	finalBody := existingBody
	if draftBody != "" {
		finalBody = draftBody
	}
	finalCc := existingCc
	if draftCc != "" {
		finalCc = draftCc
	}
	finalBcc := existingBcc
	if draftBcc != "" {
		finalBcc = draftBcc
	}

	// Build updated RFC 2822 message
	message := buildRFC2822Message(finalTo, finalSubject, finalBody, finalCc, finalBcc)

	// Base64url encode the message
	encodedMessage := base64.URLEncoding.EncodeToString([]byte(message))

	// Update the draft
	draft := &gmail.Draft{
		Message: &gmail.Message{
			Raw: encodedMessage,
		},
	}

	updated, err := service.Users.Drafts.Update("me", draftID, draft).Do()
	if err != nil {
		return fmt.Errorf("Gmail API error: %w", err)
	}

	// JSON output mode
	if GetOutputFormat() == "json" {
		type draftUpdateResult struct {
			DraftID   string `json:"draft_id"`
			MessageID string `json:"message_id"`
		}
		msgID := ""
		if updated.Message != nil {
			msgID = updated.Message.Id
		}
		return outputJSON(draftUpdateResult{
			DraftID:   updated.Id,
			MessageID: msgID,
		})
	}

	fmt.Printf("Draft updated: %s\n", draftID)
	return nil
}

func runDraftsSend(cmd *cobra.Command, args []string) error {
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

	// Send the draft
	draft := &gmail.Draft{
		Id: draftID,
	}

	sent, err := service.Users.Drafts.Send("me", draft).Do()
	if err != nil {
		return fmt.Errorf("Gmail API error: %w", err)
	}

	// JSON output mode
	if GetOutputFormat() == "json" {
		type draftSendResult struct {
			MessageID string `json:"message_id"`
		}
		return outputJSON(draftSendResult{
			MessageID: sent.Id,
		})
	}

	fmt.Printf("Draft sent as message: %s\n", sent.Id)
	return nil
}

func runDraftsDelete(cmd *cobra.Command, args []string) error {
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

	// Delete the draft
	err = service.Users.Drafts.Delete("me", draftID).Do()
	if err != nil {
		return fmt.Errorf("Gmail API error: %w", err)
	}

	// JSON output mode
	if GetOutputFormat() == "json" {
		type draftDeleteResult struct {
			DraftID string `json:"draft_id"`
			Deleted bool   `json:"deleted"`
		}
		return outputJSON(draftDeleteResult{
			DraftID: draftID,
			Deleted: true,
		})
	}

	fmt.Printf("Draft deleted: %s\n", draftID)
	return nil
}

// buildRFC2822Message builds an RFC 2822 formatted email message.
func buildRFC2822Message(to, subject, body, cc, bcc string) string {
	var msg string

	msg += fmt.Sprintf("To: %s\r\n", to)
	if cc != "" {
		msg += fmt.Sprintf("Cc: %s\r\n", cc)
	}
	if bcc != "" {
		msg += fmt.Sprintf("Bcc: %s\r\n", bcc)
	}
	msg += fmt.Sprintf("Subject: %s\r\n", subject)
	msg += "MIME-Version: 1.0\r\n"
	msg += "Content-Type: text/plain; charset=\"UTF-8\"\r\n"
	msg += "\r\n"
	msg += body

	return msg
}

