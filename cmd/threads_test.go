package cmd

import (
	"encoding/base64"
	"testing"

	"google.golang.org/api/gmail/v1"
)

func TestTruncateSnippet(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		input  string
		maxLen int
		want   string
	}{
		{
			name:   "should not truncate short string",
			input:  "hello",
			maxLen: 10,
			want:   "hello",
		},
		{
			name:   "should not truncate exact length string",
			input:  "hello",
			maxLen: 5,
			want:   "hello",
		},
		{
			name:   "should truncate long string with ellipsis",
			input:  "hello world this is long",
			maxLen: 10,
			want:   "hello w...",
		},
		{
			name:   "should return empty for empty input",
			input:  "",
			maxLen: 10,
			want:   "",
		},
		{
			name:   "should handle maxLen 3 edge case",
			input:  "abcdef",
			maxLen: 3,
			want:   "...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := truncateSnippet(tt.input, tt.maxLen)
			if got != tt.want {
				t.Errorf("truncateSnippet(%q, %d) = %q, want %q", tt.input, tt.maxLen, got, tt.want)
			}
		})
	}
}

func TestExtractMessageBody(t *testing.T) {
	t.Parallel()
	// extractMessageBody uses base64.URLEncoding (WITH padding)
	encode := func(s string) string {
		return base64.URLEncoding.EncodeToString([]byte(s))
	}

	tests := []struct {
		name    string
		payload *gmail.MessagePart
		want    string
	}{
		{
			name:    "should return empty for nil payload",
			payload: nil,
			want:    "",
		},
		{
			name: "should extract direct text/plain body",
			payload: &gmail.MessagePart{
				MimeType: "text/plain",
				Body:     &gmail.MessagePartBody{Data: encode("hello")},
			},
			want: "hello",
		},
		{
			name: "should prefer text/plain in multipart message",
			payload: &gmail.MessagePart{
				MimeType: "multipart/alternative",
				Parts: []*gmail.MessagePart{
					{
						MimeType: "text/html",
						Body:     &gmail.MessagePartBody{Data: encode("<b>html</b>")},
					},
					{
						MimeType: "text/plain",
						Body:     &gmail.MessagePartBody{Data: encode("plain text")},
					},
				},
			},
			want: "plain text",
		},
		{
			name: "should extract text/plain from nested multipart",
			payload: &gmail.MessagePart{
				MimeType: "multipart/mixed",
				Parts: []*gmail.MessagePart{
					{
						MimeType: "multipart/alternative",
						Parts: []*gmail.MessagePart{
							{
								MimeType: "text/plain",
								Body:     &gmail.MessagePartBody{Data: encode("deep nested")},
							},
						},
					},
				},
			},
			want: "deep nested",
		},
		{
			name: "should return empty when no text content exists",
			payload: &gmail.MessagePart{
				MimeType: "multipart/mixed",
				Parts: []*gmail.MessagePart{
					{
						MimeType: "image/png",
						Body:     &gmail.MessagePartBody{Data: encode("binary data")},
					},
				},
			},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := extractMessageBody(tt.payload)
			if got != tt.want {
				t.Errorf("extractMessageBody() = %q, want %q", got, tt.want)
			}
		})
	}
}
