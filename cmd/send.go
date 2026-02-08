package cmd

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"html"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"os"
	"path/filepath"
	"strings"

	"github.com/khang/google-suite-cli/internal/auth"
	"github.com/spf13/cobra"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	gmhtml "github.com/yuin/goldmark/renderer/html"
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
	Long: `Send an email message via Gmail.

Composes and sends an email with the specified recipient, subject, and body.
The body is sent as both plain text and HTML for best rendering across clients.
The body supports markdown formatting (bold, italic, links, lists, code, etc.)
which is rendered as HTML for recipients. Use \n in the body for line breaks.
Optionally include CC and BCC recipients. Supports file attachments via --attach.`,
	Example: `  # Send a simple email
  gsuite send --to "recipient@example.com" --subject "Hello" --body "Message content"

  # Send with line breaks
  gsuite send -t "user@domain.com" -s "Test" -b "Hi,\n\nHow are you?\nBest regards"

  # Send with markdown formatting
  gsuite send -t "user@domain.com" -s "Update" -b "**Important:** Please review the *attached* report.\n\n- Item one\n- Item two\n\nVisit [our site](https://example.com)"

  # Send with CC and BCC
  gsuite send -t "recipient@example.com" -s "Meeting" -b "See you there" --cc "cc@example.com" --bcc "bcc@example.com"

  # Send with file attachments
  gsuite send -t "user@domain.com" -s "Report" -b "See attached.\n\nThanks" --attach report.pdf --attach data.csv`,
	RunE: runSend,
}

func init() {
	rootCmd.AddCommand(sendCmd)

	// Required flags
	sendCmd.Flags().StringVarP(&sendTo, "to", "t", "", "Recipient email address (required)")
	sendCmd.Flags().StringVarP(&sendSubject, "subject", "s", "", "Email subject (required)")
	sendCmd.Flags().StringVarP(&sendBody, "body", "b", "", "Body content with markdown support (required)")

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

	body := interpretEscapes(sendBody)

	var rawMessage []byte
	var buildErr error
	if len(sendAttach) > 0 {
		rawMessage, buildErr = buildMultipartMessage(sendTo, sendSubject, body, sendCc, sendBcc, sendAttach)
	} else {
		rawMessage, buildErr = buildSendRFC2822Message(sendTo, sendSubject, body, sendCc, sendBcc)
	}
	if buildErr != nil {
		return fmt.Errorf("failed to build message: %w", buildErr)
	}
	encodedMessage := base64.URLEncoding.EncodeToString(rawMessage)

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

// buildSendRFC2822Message constructs an RFC 2822 message with multipart/alternative body.
func buildSendRFC2822Message(to, subject, body, cc, bcc string) ([]byte, error) {
	altBody, boundary, err := buildAlternativeBody(body)
	if err != nil {
		return nil, err
	}

	var header bytes.Buffer
	header.WriteString(fmt.Sprintf("To: %s\r\n", to))
	if cc != "" {
		header.WriteString(fmt.Sprintf("Cc: %s\r\n", cc))
	}
	if bcc != "" {
		header.WriteString(fmt.Sprintf("Bcc: %s\r\n", bcc))
	}
	header.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	header.WriteString("MIME-Version: 1.0\r\n")
	header.WriteString(fmt.Sprintf("Content-Type: multipart/alternative; boundary=%s\r\n", boundary))
	header.WriteString("\r\n")

	var result bytes.Buffer
	result.Write(header.Bytes())
	result.Write(altBody)

	return result.Bytes(), nil
}

// buildMultipartMessage constructs a MIME multipart/mixed message with
// a multipart/alternative body (text + HTML) and file attachments.
func buildMultipartMessage(to, subject, body, cc, bcc string, attachPaths []string) ([]byte, error) {
	var buf bytes.Buffer
	mixedWriter := multipart.NewWriter(&buf)

	// Write top-level headers
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
	headerBuf.WriteString(fmt.Sprintf("Content-Type: multipart/mixed; boundary=%s\r\n", mixedWriter.Boundary()))
	headerBuf.WriteString("\r\n")

	// Nest multipart/alternative as the first part of multipart/mixed
	altBody, altBoundary, err := buildAlternativeBody(body)
	if err != nil {
		return nil, err
	}

	altHeader := make(textproto.MIMEHeader)
	altHeader.Set("Content-Type", fmt.Sprintf("multipart/alternative; boundary=%s", altBoundary))
	altPart, err := mixedWriter.CreatePart(altHeader)
	if err != nil {
		return nil, fmt.Errorf("failed to create alternative part: %w", err)
	}
	if _, err := altPart.Write(altBody); err != nil {
		return nil, fmt.Errorf("failed to write alternative body: %w", err)
	}

	// Write attachment parts
	for _, attachPath := range attachPaths {
		fileData, err := os.ReadFile(attachPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read attachment %s: %w", attachPath, err)
		}

		sniffLen := 512
		if len(fileData) < sniffLen {
			sniffLen = len(fileData)
		}
		mimeType := http.DetectContentType(fileData[:sniffLen])

		basename := filepath.Base(attachPath)
		attachHeader := make(textproto.MIMEHeader)
		attachHeader.Set("Content-Type", mimeType)
		attachHeader.Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", basename))
		attachHeader.Set("Content-Transfer-Encoding", "base64")

		attachPart, err := mixedWriter.CreatePart(attachHeader)
		if err != nil {
			return nil, fmt.Errorf("failed to create attachment part: %w", err)
		}

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

	if err := mixedWriter.Close(); err != nil {
		return nil, fmt.Errorf("failed to close multipart writer: %w", err)
	}

	var result bytes.Buffer
	result.Write(headerBuf.Bytes())
	result.Write(buf.Bytes())

	return result.Bytes(), nil
}

// interpretEscapes converts literal \n, \t, and \\ sequences to their real characters.
// Bash double-quoted strings don't interpret \n, so users typing --body "Hello\nWorld"
// get literal backslash-n. This function fixes that.
func interpretEscapes(s string) string {
	var b strings.Builder
	b.Grow(len(s))

	for i := 0; i < len(s); i++ {
		if s[i] == '\\' && i+1 < len(s) {
			switch s[i+1] {
			case 'n':
				b.WriteByte('\n')
				i++
			case 't':
				b.WriteByte('\t')
				i++
			case '\\':
				b.WriteByte('\\')
				i++
			default:
				b.WriteByte(s[i])
			}
		} else {
			b.WriteByte(s[i])
		}
	}

	return b.String()
}

// plainTextToHTML renders markdown-formatted text into an HTML document.
// Supports GFM extensions: bold, italic, strikethrough, links, lists, code, tables.
func plainTextToHTML(text string) string {
	md := goldmark.New(
		goldmark.WithExtensions(extension.GFM),
		goldmark.WithRendererOptions(gmhtml.WithHardWraps()),
	)
	var buf bytes.Buffer
	if err := md.Convert([]byte(text), &buf); err != nil {
		escaped := html.EscapeString(text)
		return "<!DOCTYPE html><html><body>" + strings.ReplaceAll(escaped, "\n", "<br>\n") + "</body></html>"
	}
	return "<!DOCTYPE html><html><body>" + buf.String() + "</body></html>"
}

// buildAlternativeBody returns the raw bytes and boundary of a multipart/alternative
// containing text/plain and text/html parts.
func buildAlternativeBody(body string) ([]byte, string, error) {
	var buf bytes.Buffer
	altWriter := multipart.NewWriter(&buf)

	plainHeader := make(textproto.MIMEHeader)
	plainHeader.Set("Content-Type", "text/plain; charset=UTF-8")
	plainPart, err := altWriter.CreatePart(plainHeader)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create text/plain part: %w", err)
	}
	if _, err := plainPart.Write([]byte(body)); err != nil {
		return nil, "", fmt.Errorf("failed to write text/plain body: %w", err)
	}

	htmlHeader := make(textproto.MIMEHeader)
	htmlHeader.Set("Content-Type", "text/html; charset=UTF-8")
	htmlPart, err := altWriter.CreatePart(htmlHeader)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create text/html part: %w", err)
	}
	if _, err := htmlPart.Write([]byte(plainTextToHTML(body))); err != nil {
		return nil, "", fmt.Errorf("failed to write text/html body: %w", err)
	}

	boundary := altWriter.Boundary()
	if err := altWriter.Close(); err != nil {
		return nil, "", fmt.Errorf("failed to close alternative writer: %w", err)
	}

	return buf.Bytes(), boundary, nil
}
