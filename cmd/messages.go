package cmd

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
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

	// messagesGetAttachmentCmd flags
	attachmentOutput string
)

// attachmentInfo holds information about a message attachment.
type attachmentInfo struct {
	Filename     string
	MimeType     string
	Size         int64
	AttachmentId string
}

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

// messagesGetAttachmentCmd represents the messages get-attachment command
var messagesGetAttachmentCmd = &cobra.Command{
	Use:   "get-attachment <message-id> <attachment-id>",
	Short: "Download an attachment from a message",
	Long: `Download a specific attachment from a Gmail message.

Requires the message ID and the attachment ID. The attachment ID can be found
by viewing the message with 'messages get', which displays attachment details
including their IDs.

By default, the file is saved with its original filename. Use --output to
specify a custom output path.`,
	Example: `  # Download an attachment (uses original filename)
  gsuite messages get-attachment 18d5a1b2c3d4e5f6 ANGjdJ8abc123

  # Download to a specific file
  gsuite messages get-attachment 18d5a1b2c3d4e5f6 ANGjdJ8abc123 --output ./downloads/report.pdf`,
	Args: cobra.ExactArgs(2),
	RunE: runMessagesGetAttachment,
}

func init() {
	rootCmd.AddCommand(messagesCmd)
	messagesCmd.AddCommand(messagesListCmd)
	messagesCmd.AddCommand(messagesGetCmd)
	messagesCmd.AddCommand(messagesModifyCmd)
	messagesCmd.AddCommand(messagesGetAttachmentCmd)

	// messagesListCmd flags
	messagesListCmd.Flags().Int64VarP(&maxResults, "max-results", "n", 10, "Maximum number of messages to return (max 500)")
	messagesListCmd.Flags().StringVar(&labelIDs, "label-ids", "", "Comma-separated label IDs to filter by (e.g., INBOX,UNREAD)")
	messagesListCmd.Flags().StringVarP(&query, "query", "q", "", "Gmail search query string")

	// messagesModifyCmd flags
	messagesModifyCmd.Flags().StringVar(&addLabels, "add-labels", "", "Comma-separated label IDs to add (e.g., STARRED,Label_123)")
	messagesModifyCmd.Flags().StringVar(&removeLabels, "remove-labels", "", "Comma-separated label IDs to remove (e.g., UNREAD,INBOX)")

	// messagesGetAttachmentCmd flags
	messagesGetAttachmentCmd.Flags().StringVarP(&attachmentOutput, "output", "o", "", "Output file path (defaults to attachment filename)")
}

func runMessagesList(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	service, err := auth.NewGmailService(ctx, GetAccountEmail())
	if err != nil {
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
		if GetOutputFormat() == "json" {
			return outputJSON([]struct{}{})
		}
		fmt.Println("No messages found.")
		return nil
	}

	// JSON output mode
	if GetOutputFormat() == "json" {
		type messageListItem struct {
			ID       string `json:"id"`
			ThreadID string `json:"thread_id"`
			Snippet  string `json:"snippet"`
		}
		var results []messageListItem
		for _, msg := range resp.Messages {
			detail, err := service.Users.Messages.Get("me", msg.Id).Format("metadata").Do()
			if err != nil {
				results = append(results, messageListItem{ID: msg.Id, ThreadID: msg.ThreadId})
				continue
			}
			results = append(results, messageListItem{
				ID:       msg.Id,
				ThreadID: msg.ThreadId,
				Snippet:  detail.Snippet,
			})
		}
		return outputJSON(results)
	}

	// Print results (text mode)
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

	ctx := context.Background()

	service, err := auth.NewGmailService(ctx, GetAccountEmail())
	if err != nil {
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

	// Extract body content
	body := extractBody(msg)
	if body == "" {
		body = msg.Snippet
	}

	// Extract attachments
	attachments := findAttachments(msg.Payload.Parts)

	// JSON output mode
	if GetOutputFormat() == "json" {
		type attachmentJSON struct {
			Filename     string `json:"filename"`
			MimeType     string `json:"mime_type"`
			Size         int64  `json:"size"`
			AttachmentID string `json:"attachment_id"`
		}
		type messageGetResult struct {
			From        string           `json:"from"`
			To          string           `json:"to"`
			Subject     string           `json:"subject"`
			Date        string           `json:"date"`
			Body        string           `json:"body"`
			Snippet     string           `json:"snippet"`
			Labels      []string         `json:"labels"`
			Attachments []attachmentJSON `json:"attachments"`
		}
		result := messageGetResult{
			From:    from,
			To:      to,
			Subject: subject,
			Date:    date,
			Body:    body,
			Snippet: msg.Snippet,
			Labels:  msg.LabelIds,
		}
		if result.Labels == nil {
			result.Labels = []string{}
		}
		for _, att := range attachments {
			result.Attachments = append(result.Attachments, attachmentJSON{
				Filename:     att.Filename,
				MimeType:     att.MimeType,
				Size:         att.Size,
				AttachmentID: att.AttachmentId,
			})
		}
		if result.Attachments == nil {
			result.Attachments = []attachmentJSON{}
		}
		return outputJSON(result)
	}

	// Print headers (text mode)
	fmt.Printf("From: %s\n", from)
	fmt.Printf("To: %s\n", to)
	fmt.Printf("Subject: %s\n", subject)
	fmt.Printf("Date: %s\n", date)
	fmt.Println("---")

	// Print body content
	if body != "" {
		fmt.Println(body)
	} else {
		fmt.Printf("(Snippet) %s\n", msg.Snippet)
	}

	// Display attachment info if present
	if len(attachments) > 0 {
		fmt.Println("---")
		for _, att := range attachments {
			fmt.Printf("Attachment: %s (%s, %d bytes, ID: %s)\n", att.Filename, att.MimeType, att.Size, att.AttachmentId)
		}
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

	ctx := context.Background()

	service, err := auth.NewGmailService(ctx, GetAccountEmail())
	if err != nil {
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

	// JSON output mode
	if GetOutputFormat() == "json" {
		type modifyResult struct {
			MessageID     string   `json:"message_id"`
			LabelsAdded   []string `json:"labels_added"`
			LabelsRemoved []string `json:"labels_removed"`
		}
		result := modifyResult{
			MessageID:     messageID,
			LabelsAdded:   addLabelsList,
			LabelsRemoved: removeLabelsList,
		}
		if result.LabelsAdded == nil {
			result.LabelsAdded = []string{}
		}
		if result.LabelsRemoved == nil {
			result.LabelsRemoved = []string{}
		}
		return outputJSON(result)
	}

	// Print success message with details (text mode)
	fmt.Printf("Message modified: %s\n", messageID)
	if len(addLabelsList) > 0 {
		fmt.Printf("  Labels added: %s\n", strings.Join(addLabelsList, ", "))
	}
	if len(removeLabelsList) > 0 {
		fmt.Printf("  Labels removed: %s\n", strings.Join(removeLabelsList, ", "))
	}

	return nil
}

// findAttachments recursively searches through MIME parts and returns info about parts
// that have a non-empty Filename (i.e., actual file attachments).
func findAttachments(parts []*gmail.MessagePart) []attachmentInfo {
	var attachments []attachmentInfo
	for _, part := range parts {
		if part.Filename != "" && part.Body != nil {
			attachments = append(attachments, attachmentInfo{
				Filename:     part.Filename,
				MimeType:     part.MimeType,
				Size:         part.Body.Size,
				AttachmentId: part.Body.AttachmentId,
			})
		}
		// Recurse into nested parts
		if len(part.Parts) > 0 {
			attachments = append(attachments, findAttachments(part.Parts)...)
		}
	}
	return attachments
}

func runMessagesGetAttachment(cmd *cobra.Command, args []string) error {
	messageID := args[0]
	attachmentID := args[1]

	ctx := context.Background()

	service, err := auth.NewGmailService(ctx, GetAccountEmail())
	if err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	// Get the attachment data
	att, err := service.Users.Messages.Attachments.Get("me", messageID, attachmentID).Do()
	if err != nil {
		return fmt.Errorf("Gmail API error: %w", err)
	}

	// Decode the attachment data (base64url encoded)
	decoded, err := base64.URLEncoding.DecodeString(att.Data)
	if err != nil {
		// Try without padding (RawURLEncoding)
		decoded, err = base64.RawURLEncoding.DecodeString(att.Data)
		if err != nil {
			return fmt.Errorf("failed to decode attachment data: %w", err)
		}
	}

	// Determine output filename
	outputPath := attachmentOutput
	if outputPath == "" {
		// Get the filename from the message metadata
		msg, err := service.Users.Messages.Get("me", messageID).Format("full").Do()
		if err == nil {
			attachments := findAttachments(msg.Payload.Parts)
			for _, a := range attachments {
				if a.AttachmentId == attachmentID {
					outputPath = a.Filename
					break
				}
			}
		}
		// Fallback if filename not found
		if outputPath == "" {
			outputPath = fmt.Sprintf("attachment_%s", attachmentID)
		}
	}

	// Write the attachment to file
	err = os.WriteFile(outputPath, decoded, 0644)
	if err != nil {
		return fmt.Errorf("failed to write attachment file: %w", err)
	}

	// JSON output mode
	if GetOutputFormat() == "json" {
		type attachmentResult struct {
			FilePath string `json:"file_path"`
			Size     int64  `json:"size"`
		}
		return outputJSON(attachmentResult{
			FilePath: outputPath,
			Size:     int64(len(decoded)),
		})
	}

	fmt.Printf("Attachment saved: %s (%d bytes)\n", outputPath, len(decoded))
	return nil
}
