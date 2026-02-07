package cmd

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"os"
	"path/filepath"
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
	sendAttach  []string
)

// sendCmd represents the send command
var sendCmd = &cobra.Command{
	Use:   "send",
	Short: "Send an email message",
	Long: `Send a plain text email message via Gmail.

Composes and sends an email with the specified recipient, subject, and body.
Optionally include CC and BCC recipients. Supports file attachments via --attach.`,
	Example: `  # Send a simple email
  gsuite send --to "recipient@example.com" --subject "Hello" --body "Message content"

  # Send with short flags
  gsuite send -t "user@domain.com" -s "Test" -b "Body text"

  # Send with CC and BCC
  gsuite send -t "recipient@example.com" -s "Meeting" -b "See you there" --cc "cc@example.com" --bcc "bcc@example.com"

  # Send to multiple CC recipients
  gsuite send -t "main@example.com" -s "Update" -b "Content" --cc "one@example.com,two@example.com"

  # Send with file attachments
  gsuite send -t "user@domain.com" -s "Report" -b "See attached." --attach report.pdf --attach data.csv`,
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
	sendCmd.Flags().StringArrayVarP(&sendAttach, "attach", "a", nil, "File path to attach (can be specified multiple times)")
}

func runSend(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	service, err := auth.NewGmailService(ctx)
	if err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	// Validate attachment files exist before building message
	for _, attachPath := range sendAttach {
		if _, err := os.Stat(attachPath); err != nil {
			return fmt.Errorf("attachment file not found: %s", attachPath)
		}
	}

	var encodedMessage string
	if len(sendAttach) > 0 {
		// Build MIME multipart message with attachments
		mimeMessage, err := buildMultipartMessage(sendTo, sendSubject, sendBody, sendCc, sendBcc, sendAttach)
		if err != nil {
			return fmt.Errorf("failed to build message with attachments: %w", err)
		}
		encodedMessage = base64.URLEncoding.EncodeToString(mimeMessage)
	} else {
		// Build simple RFC 2822 formatted message (no attachments)
		message := buildSendRFC2822Message(sendTo, sendSubject, sendBody, sendCc, sendBcc)
		encodedMessage = base64.URLEncoding.EncodeToString([]byte(message))
	}

	// Create the Gmail message object
	gmailMessage := &gmail.Message{
		Raw: encodedMessage,
	}

	// Send the message
	sent, err := service.Users.Messages.Send("me", gmailMessage).Do()
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	// JSON output mode
	if GetOutputFormat() == "json" {
		type sendResult struct {
			MessageID string `json:"message_id"`
		}
		return outputJSON(sendResult{
			MessageID: sent.Id,
		})
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

// buildMultipartMessage constructs a MIME multipart message with file attachments.
func buildMultipartMessage(to, subject, body, cc, bcc string, attachPaths []string) ([]byte, error) {
	var buf bytes.Buffer

	// Create multipart writer
	writer := multipart.NewWriter(&buf)

	// Write top-level headers
	buf.Reset() // Reset to write headers before multipart content
	var headerBuf bytes.Buffer
	headerBuf.WriteString(fmt.Sprintf("To: %s\r\n", to))
	if cc != "" {
		headerBuf.WriteString(fmt.Sprintf("Cc: %s\r\n", cc))
	}
	if bcc != "" {
		headerBuf.WriteString(fmt.Sprintf("Bcc: %s\r\n", bcc))
	}
	headerBuf.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	headerBuf.WriteString("MIME-Version: 1.0\r\n")
	headerBuf.WriteString(fmt.Sprintf("Content-Type: multipart/mixed; boundary=%s\r\n", writer.Boundary()))
	headerBuf.WriteString("\r\n")

	// Write the text body part
	textHeader := make(textproto.MIMEHeader)
	textHeader.Set("Content-Type", "text/plain; charset=UTF-8")
	textPart, err := writer.CreatePart(textHeader)
	if err != nil {
		return nil, fmt.Errorf("failed to create text part: %w", err)
	}
	if _, err := textPart.Write([]byte(body)); err != nil {
		return nil, fmt.Errorf("failed to write body: %w", err)
	}

	// Write attachment parts
	for _, attachPath := range attachPaths {
		fileData, err := os.ReadFile(attachPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read attachment %s: %w", attachPath, err)
		}

		// Detect MIME type from file content
		sniffLen := 512
		if len(fileData) < sniffLen {
			sniffLen = len(fileData)
		}
		mimeType := http.DetectContentType(fileData[:sniffLen])

		// Create attachment part
		basename := filepath.Base(attachPath)
		attachHeader := make(textproto.MIMEHeader)
		attachHeader.Set("Content-Type", mimeType)
		attachHeader.Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", basename))
		attachHeader.Set("Content-Transfer-Encoding", "base64")

		attachPart, err := writer.CreatePart(attachHeader)
		if err != nil {
			return nil, fmt.Errorf("failed to create attachment part: %w", err)
		}

		// Base64 encode and write with line wrapping at 76 chars
		encoded := base64.StdEncoding.EncodeToString(fileData)
		for i := 0; i < len(encoded); i += 76 {
			end := i + 76
			if end > len(encoded) {
				end = len(encoded)
			}
			if _, err := attachPart.Write([]byte(encoded[i:end] + "\r\n")); err != nil {
				return nil, fmt.Errorf("failed to write attachment data: %w", err)
			}
		}
	}

	// Close the multipart writer
	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close multipart writer: %w", err)
	}

	// Combine headers and multipart body
	var result bytes.Buffer
	result.Write(headerBuf.Bytes())
	result.Write(buf.Bytes())

	return result.Bytes(), nil
}
