package cmd

import (
	"testing"
)

func TestValidateCalendarCreateFlags(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		start    string
		end      string
		duration string
		allDay   bool
		wantErr  bool
		errMsg   string
	}{
		{
			name:  "valid: start only",
			start: "2026-03-15 09:00",
		},
		{
			name:  "valid: start and end",
			start: "2026-03-15 09:00",
			end:   "2026-03-15 10:00",
		},
		{
			name:     "valid: start and duration",
			start:    "2026-03-15 09:00",
			duration: "1h",
		},
		{
			name:   "valid: all-day",
			start:  "2026-03-15",
			allDay: true,
		},
		{
			name:     "invalid: end and duration",
			start:    "2026-03-15 09:00",
			end:      "2026-03-15 10:00",
			duration: "1h",
			wantErr:  true,
			errMsg:   "--end and --duration are mutually exclusive",
		},
		{
			name:     "invalid: all-day and duration",
			start:    "2026-03-15",
			allDay:   true,
			duration: "1h",
			wantErr:  true,
			errMsg:   "--all-day and --duration cannot be combined",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := validateCalendarCreateFlags(tt.start, tt.end, tt.duration, tt.allDay)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				if err.Error() != tt.errMsg {
					t.Fatalf("error = %q, want %q", err.Error(), tt.errMsg)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestValidateAttendeeEmails(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		want    []string
		wantErr bool
	}{
		{
			name:  "single email",
			input: "alice@example.com",
			want:  []string{"alice@example.com"},
		},
		{
			name:  "multiple emails",
			input: "alice@example.com,bob@example.com",
			want:  []string{"alice@example.com", "bob@example.com"},
		},
		{
			name:  "emails with whitespace",
			input: " alice@example.com , bob@example.com ",
			want:  []string{"alice@example.com", "bob@example.com"},
		},
		{
			name:  "email with display name",
			input: "Alice <alice@example.com>",
			want:  []string{"alice@example.com"},
		},
		{
			name:  "skip empty parts",
			input: "alice@example.com,,bob@example.com",
			want:  []string{"alice@example.com", "bob@example.com"},
		},
		{
			name:    "invalid email",
			input:   "not-an-email",
			wantErr: true,
		},
		{
			name:    "one valid one invalid",
			input:   "alice@example.com,bad-email",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := validateAttendeeEmails(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got %v", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(got) != len(tt.want) {
				t.Fatalf("got %d emails, want %d", len(got), len(tt.want))
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Fatalf("email[%d] = %q, want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}
