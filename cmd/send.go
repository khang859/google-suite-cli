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
	// sendCmd flags
	sendTo      string
	sendSubject string
	sendBody    string
	sendCc      string
	sendBcc     string
)

// sendCmd represents the send command
var sendCmd = &cobra.Command{
	Use:   "send",
	Short: "Send an email message",
	Long: `Send a plain text email message via Gmail.

Composes and sends an email with the specified recipient, subject, and body.
Optionally include CC and BCC recipients.`,
	Example: `  # Send a simple email
  gsuite send --to "recipient@example.com" --subject "Hello" --body "Message content"

  # Send with short flags
  gsuite send -t "user@domain.com" -s "Test" -b "Body text"

  # Send with CC and BCC
  gsuite send -t "recipient@example.com" -s "Meeting" -b "See you there" --cc "cc@example.com" --bcc "bcc@example.com"

  # Send to multiple CC recipients
  gsuite send -t "main@example.com" -s "Update" -b "Content" --cc "one@example.com,two@example.com"`,
	RunE: runSend,
}

func init() {
	rootCmd.AddCommand(sendCmd)

	// Required flags
	sendCmd.Flags().StringVarP(&sendTo, "to", "t", "", "Recipient email address (required)")
	sendCmd.Flags().StringVarP(&sendSubject, "subject", "s", "", "Email subject (required)")
	sendCmd.Flags().StringVarP(&sendBody, "body", "b", "", "Plain text body content (required)")

	// Mark required flags
	sendCmd.MarkFlagRequired("to")
	sendCmd.MarkFlagRequired("subject")
	sendCmd.MarkFlagRequired("body")

	// Optional flags
	sendCmd.Flags().StringVar(&sendCc, "cc", "", "CC recipients (comma-separated)")
	sendCmd.Flags().StringVar(&sendBcc, "bcc", "", "BCC recipients (comma-separated)")
}

func runSend(cmd *cobra.Command, args []string) error {
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
	message := buildSendRFC2822Message(sendTo, sendSubject, sendBody, sendCc, sendBcc)

	// Base64url encode the message
	encodedMessage := base64.URLEncoding.EncodeToString([]byte(message))

	// Create the Gmail message object
	gmailMessage := &gmail.Message{
		Raw: encodedMessage,
	}

	// Send the message
	sent, err := service.Users.Messages.Send("me", gmailMessage).Do()
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	fmt.Printf("Message sent successfully!\nMessage ID: %s\n", sent.Id)
	return nil
}

// buildSendRFC2822Message constructs an RFC 2822 formatted email message for sending.
func buildSendRFC2822Message(to, subject, body, cc, bcc string) string {
	var builder strings.Builder

	// Write headers
	builder.WriteString(fmt.Sprintf("To: %s\r\n", to))

	if cc != "" {
		builder.WriteString(fmt.Sprintf("Cc: %s\r\n", cc))
	}

	if bcc != "" {
		builder.WriteString(fmt.Sprintf("Bcc: %s\r\n", bcc))
	}

	builder.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	builder.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")

	// Empty line separates headers from body
	builder.WriteString("\r\n")

	// Write body
	builder.WriteString(body)

	return builder.String()
}
