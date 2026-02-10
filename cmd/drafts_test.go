package cmd

import (
	"encoding/base64"
	"testing"

	"google.golang.org/api/gmail/v1"
)

func TestDecodeDraftBase64URL(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		encoded string
		want    string
	}{
		{
			name:    "should decode valid encoded input",
			encoded: base64.URLEncoding.EncodeToString([]byte("draft content")),
			want:    "draft content",
		},
		{
			name:    "should decode valid raw encoded input without padding",
			encoded: base64.RawURLEncoding.EncodeToString([]byte("raw draft")),
			want:    "raw draft",
		},
		{
			name:    "should return empty for invalid encoding",
			encoded: "###invalid###",
			want:    "",
		},
		{
			name:    "should return empty for empty input",
			encoded: "",
			want:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := decodeDraftBase64URL(tt.encoded)
			if got != tt.want {
				t.Errorf("decodeDraftBase64URL(%q) = %q, want %q", tt.encoded, got, tt.want)
			}
		})
	}
}

func TestExtractDraftBody(t *testing.T) {
	t.Parallel()
	encode := func(s string) string {
		return base64.URLEncoding.EncodeToString([]byte(s))
	}

	tests := []struct {
		name string
		msg  *gmail.Message
		want string
	}{
		{
			name: "should return empty for nil payload",
			msg:  &gmail.Message{Payload: nil},
			want: "",
		},
		{
			name: "should extract direct text/plain body",
			msg: &gmail.Message{
				Payload: &gmail.MessagePart{
					MimeType: "text/plain",
					Body:     &gmail.MessagePartBody{Data: encode("draft body")},
				},
			},
			want: "draft body",
		},
		{
			name: "should prefer text/plain in multipart message",
			msg: &gmail.Message{
				Payload: &gmail.MessagePart{
					MimeType: "multipart/alternative",
					Parts: []*gmail.MessagePart{
						{
							MimeType: "text/html",
							Body:     &gmail.MessagePartBody{Data: encode("<p>html</p>")},
						},
						{
							MimeType: "text/plain",
							Body:     &gmail.MessagePartBody{Data: encode("plain draft")},
						},
					},
				},
			},
			want: "plain draft",
		},
		{
			name: "should extract text/plain from nested multipart",
			msg: &gmail.Message{
				Payload: &gmail.MessagePart{
					MimeType: "multipart/mixed",
					Parts: []*gmail.MessagePart{
						{
							MimeType: "multipart/alternative",
							Parts: []*gmail.MessagePart{
								{
									MimeType: "text/plain",
									Body:     &gmail.MessagePartBody{Data: encode("nested draft")},
								},
							},
						},
					},
				},
			},
			want: "nested draft",
		},
		{
			name: "should return empty when no text/plain exists",
			msg: &gmail.Message{
				Payload: &gmail.MessagePart{
					MimeType: "multipart/alternative",
					Parts: []*gmail.MessagePart{
						{
							MimeType: "text/html",
							Body:     &gmail.MessagePartBody{Data: encode("<p>only html</p>")},
						},
					},
				},
			},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := extractDraftBody(tt.msg)
			if got != tt.want {
				t.Errorf("extractDraftBody() = %q, want %q", got, tt.want)
			}
		})
	}
}
