package cmd

import (
	"encoding/base64"
	"testing"

	"google.golang.org/api/gmail/v1"
)

func TestDecodeBase64URL(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		encoded string
		want    string
	}{
		{
			name:    "should decode valid padded base64url",
			encoded: base64.URLEncoding.EncodeToString([]byte("hello world")),
			want:    "hello world",
		},
		{
			name:    "should decode valid raw base64url without padding",
			encoded: base64.RawURLEncoding.EncodeToString([]byte("test")),
			want:    "test",
		},
		{
			name:    "should return empty for empty input",
			encoded: "",
			want:    "",
		},
		{
			name:    "should return empty for invalid base64",
			encoded: "!!!not-valid-base64!!!",
			want:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := decodeBase64URL(tt.encoded)
			if got != tt.want {
				t.Errorf("decodeBase64URL(%q) = %q, want %q", tt.encoded, got, tt.want)
			}
		})
	}
}

func TestExtractBody(t *testing.T) {
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
					Body:     &gmail.MessagePartBody{Data: encode("hello")},
				},
			},
			want: "hello",
		},
		{
			name: "should prefer text/plain in multipart message",
			msg: &gmail.Message{
				Payload: &gmail.MessagePart{
					MimeType: "multipart/alternative",
					Parts: []*gmail.MessagePart{
						{
							MimeType: "text/html",
							Body:     &gmail.MessagePartBody{Data: encode("<b>hello</b>")},
						},
						{
							MimeType: "text/plain",
							Body:     &gmail.MessagePartBody{Data: encode("plain body")},
						},
					},
				},
			},
			want: "plain body",
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
									Body:     &gmail.MessagePartBody{Data: encode("nested plain")},
								},
							},
						},
					},
				},
			},
			want: "nested plain",
		},
		{
			name: "should return empty when no text/plain exists",
			msg: &gmail.Message{
				Payload: &gmail.MessagePart{
					MimeType: "multipart/alternative",
					Parts: []*gmail.MessagePart{
						{
							MimeType: "text/html",
							Body:     &gmail.MessagePartBody{Data: encode("<p>html only</p>")},
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
			got := extractBody(tt.msg)
			if got != tt.want {
				t.Errorf("extractBody() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFindPlainTextPart(t *testing.T) {
	t.Parallel()
	encode := func(s string) string {
		return base64.URLEncoding.EncodeToString([]byte(s))
	}

	tests := []struct {
		name  string
		parts []*gmail.MessagePart
		want  string
	}{
		{
			name:  "should return empty for nil parts",
			parts: nil,
			want:  "",
		},
		{
			name: "should find text/plain in flat list",
			parts: []*gmail.MessagePart{
				{
					MimeType: "text/html",
					Body:     &gmail.MessagePartBody{Data: encode("<b>hi</b>")},
				},
				{
					MimeType: "text/plain",
					Body:     &gmail.MessagePartBody{Data: encode("found it")},
				},
			},
			want: "found it",
		},
		{
			name: "should find text/plain in nested parts",
			parts: []*gmail.MessagePart{
				{
					MimeType: "multipart/alternative",
					Parts: []*gmail.MessagePart{
						{
							MimeType: "text/plain",
							Body:     &gmail.MessagePartBody{Data: encode("nested text")},
						},
					},
				},
			},
			want: "nested text",
		},
		{
			name: "should return empty when no text/plain found",
			parts: []*gmail.MessagePart{
				{
					MimeType: "text/html",
					Body:     &gmail.MessagePartBody{Data: encode("<p>only html</p>")},
				},
			},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := findPlainTextPart(tt.parts)
			if got != tt.want {
				t.Errorf("findPlainTextPart() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFindAttachments(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		parts     []*gmail.MessagePart
		wantCount int
		wantNames []string
	}{
		{
			name:      "should return empty for nil parts",
			parts:     nil,
			wantCount: 0,
			wantNames: nil,
		},
		{
			name: "should return empty for text-only parts",
			parts: []*gmail.MessagePart{
				{
					MimeType: "text/plain",
					Body:     &gmail.MessagePartBody{Data: "dGVzdA=="},
				},
				{
					MimeType: "text/html",
					Body:     &gmail.MessagePartBody{Data: "dGVzdA=="},
				},
			},
			wantCount: 0,
			wantNames: nil,
		},
		{
			name: "should find single attachment",
			parts: []*gmail.MessagePart{
				{
					MimeType: "text/plain",
					Body:     &gmail.MessagePartBody{Data: "dGVzdA=="},
				},
				{
					Filename: "report.pdf",
					MimeType: "application/pdf",
					Body: &gmail.MessagePartBody{
						Size:         1024,
						AttachmentId: "att-123",
					},
				},
			},
			wantCount: 1,
			wantNames: []string{"report.pdf"},
		},
		{
			name: "should find nested attachment",
			parts: []*gmail.MessagePart{
				{
					MimeType: "multipart/mixed",
					Parts: []*gmail.MessagePart{
						{
							Filename: "nested.txt",
							MimeType: "text/plain",
							Body: &gmail.MessagePartBody{
								Size:         256,
								AttachmentId: "att-nested",
							},
						},
					},
				},
			},
			wantCount: 1,
			wantNames: []string{"nested.txt"},
		},
		{
			name: "should find multiple attachments",
			parts: []*gmail.MessagePart{
				{
					Filename: "image.png",
					MimeType: "image/png",
					Body: &gmail.MessagePartBody{
						Size:         2048,
						AttachmentId: "att-1",
					},
				},
				{
					Filename: "doc.pdf",
					MimeType: "application/pdf",
					Body: &gmail.MessagePartBody{
						Size:         4096,
						AttachmentId: "att-2",
					},
				},
			},
			wantCount: 2,
			wantNames: []string{"image.png", "doc.pdf"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := findAttachments(tt.parts)
			if len(got) != tt.wantCount {
				t.Fatalf("findAttachments() returned %d attachments, want %d", len(got), tt.wantCount)
			}
			for i, name := range tt.wantNames {
				if got[i].Filename != name {
					t.Errorf("attachment[%d].Filename = %q, want %q", i, got[i].Filename, name)
				}
			}
		})
	}
}
