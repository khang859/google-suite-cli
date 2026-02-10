package cmd

import (
	"strings"
	"testing"
)

func TestInterpretEscapes(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "should convert backslash-n to newline",
			input: `hello\nworld`,
			want:  "hello\nworld",
		},
		{
			name:  "should convert backslash-t to tab",
			input: `col1\tcol2`,
			want:  "col1\tcol2",
		},
		{
			name:  "should convert double backslash to single",
			input: `path\\to\\file`,
			want:  `path\to\file`,
		},
		{
			name:  "should pass through text without escapes",
			input: "plain text",
			want:  "plain text",
		},
		{
			name:  "should preserve unrecognized escape sequences",
			input: `hello\xworld`,
			want:  `hello\xworld`,
		},
		{
			name:  "should preserve trailing backslash",
			input: `trailing\`,
			want:  `trailing\`,
		},
		{
			name:  "should handle multiple escape types",
			input: `a\nb\tc`,
			want:  "a\nb\tc",
		},
		{
			name:  "should return empty for empty input",
			input: "",
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := interpretEscapes(tt.input)
			if got != tt.want {
				t.Errorf("interpretEscapes(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestPlainTextToHTML(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		input    string
		contains []string
	}{
		{
			name:  "should wrap simple text in html structure",
			input: "hello",
			contains: []string{
				"<!DOCTYPE html>",
				"<html>",
				"<body>",
				"</body>",
				"</html>",
				"hello",
			},
		},
		{
			name:  "should render bold markdown as strong tag",
			input: "**bold**",
			contains: []string{
				"<strong>bold</strong>",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := plainTextToHTML(tt.input)
			for _, want := range tt.contains {
				if !strings.Contains(got, want) {
					t.Errorf("plainTextToHTML(%q) = %q, want it to contain %q", tt.input, got, want)
				}
			}
		})
	}
}

func TestBuildSendRFC2822Message(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		to          string
		subject     string
		body        string
		cc          string
		bcc         string
		contains    []string
		notContains []string
	}{
		{
			name:    "should build basic multipart message structure",
			to:      "recipient@example.com",
			subject: "Test Subject",
			body:    "Hello there",
			cc:      "",
			bcc:     "",
			contains: []string{
				"multipart/alternative",
				"text/plain",
				"text/html",
				"To: recipient@example.com",
				"Subject: Test Subject",
			},
			notContains: []string{
				"Cc:",
				"Bcc:",
			},
		},
		{
			name:    "should include CC and BCC headers when provided",
			to:      "to@example.com",
			subject: "With CC",
			body:    "body text",
			cc:      "cc@example.com",
			bcc:     "bcc@example.com",
			contains: []string{
				"Cc: cc@example.com",
				"Bcc: bcc@example.com",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			raw, err := buildSendRFC2822Message(tt.to, tt.subject, tt.body, tt.cc, tt.bcc)
			if err != nil {
				t.Fatalf("buildSendRFC2822Message() error = %v", err)
			}
			msg := string(raw)
			for _, want := range tt.contains {
				if !strings.Contains(msg, want) {
					t.Errorf("message missing %q", want)
				}
			}
			for _, notWant := range tt.notContains {
				if strings.Contains(msg, notWant) {
					t.Errorf("message should not contain %q", notWant)
				}
			}
		})
	}
}

func TestBuildAlternativeBody(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		body string
	}{
		{
			name: "should build alternative body for basic text",
			body: "Hello world",
		},
		{
			name: "should build alternative body with markdown",
			body: "**bold** and _italic_",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			raw, boundary, err := buildAlternativeBody(tt.body)
			if err != nil {
				t.Fatalf("buildAlternativeBody() error = %v", err)
			}
			if len(raw) == 0 {
				t.Error("buildAlternativeBody() returned empty bytes")
			}
			if boundary == "" {
				t.Error("buildAlternativeBody() returned empty boundary")
			}
			content := string(raw)
			if !strings.Contains(content, "text/plain") {
				t.Error("output missing text/plain content type")
			}
			if !strings.Contains(content, "text/html") {
				t.Error("output missing text/html content type")
			}
			if !strings.Contains(content, tt.body) {
				t.Errorf("output missing body text %q", tt.body)
			}
		})
	}
}

func TestBuildRFC2822Message(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		to          string
		subject     string
		body        string
		cc          string
		bcc         string
		contains    []string
		notContains []string
	}{
		{
			name:    "should build basic plain text message",
			to:      "user@example.com",
			subject: "Hello",
			body:    "Message body",
			cc:      "",
			bcc:     "",
			contains: []string{
				"To: user@example.com",
				"Subject: Hello",
				"Content-Type: text/plain",
				"Message body",
			},
			notContains: []string{
				"Cc:",
				"Bcc:",
			},
		},
		{
			name:    "should include CC and BCC headers when provided",
			to:      "to@example.com",
			subject: "Subject",
			body:    "Body",
			cc:      "cc@example.com",
			bcc:     "bcc@example.com",
			contains: []string{
				"Cc: cc@example.com",
				"Bcc: bcc@example.com",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := buildRFC2822Message(tt.to, tt.subject, tt.body, tt.cc, tt.bcc)
			for _, want := range tt.contains {
				if !strings.Contains(got, want) {
					t.Errorf("buildRFC2822Message() missing %q in output:\n%s", want, got)
				}
			}
			for _, notWant := range tt.notContains {
				if strings.Contains(got, notWant) {
					t.Errorf("buildRFC2822Message() should not contain %q in output:\n%s", notWant, got)
				}
			}
		})
	}
}
