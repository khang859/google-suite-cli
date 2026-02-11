package cmd

import (
	"testing"
	"time"

	"google.golang.org/api/calendar/v3"
)

// Fixed reference time: Sunday, March 15, 2026, 12:00 UTC
var refNow = time.Date(2026, 3, 15, 12, 0, 0, 0, time.UTC)

func TestParseDateTime(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		want    time.Time
		wantErr bool
	}{
		{
			name:  "RFC3339",
			input: "2026-03-15T09:00:00-07:00",
			want:  time.Date(2026, 3, 15, 9, 0, 0, 0, time.FixedZone("", -7*3600)),
		},
		{
			name:  "date only",
			input: "2026-03-15",
			want:  time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC),
		},
		{
			name:  "date and time",
			input: "2026-03-15 09:00",
			want:  time.Date(2026, 3, 15, 9, 0, 0, 0, time.UTC),
		},
		{
			name:  "ISO without timezone",
			input: "2026-03-15T09:00:00",
			want:  time.Date(2026, 3, 15, 9, 0, 0, 0, time.UTC),
		},
		{
			name:  "time only anchors to today",
			input: "09:00",
			want:  time.Date(2026, 3, 15, 9, 0, 0, 0, time.UTC),
		},
		{
			name:  "today keyword",
			input: "today",
			want:  time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC),
		},
		{
			name:  "tomorrow keyword",
			input: "tomorrow",
			want:  time.Date(2026, 3, 16, 0, 0, 0, 0, time.UTC),
		},
		{
			name:  "relative +3d",
			input: "+3d",
			want:  time.Date(2026, 3, 18, 0, 0, 0, 0, time.UTC),
		},
		{
			name:  "monday from Sunday",
			input: "monday",
			want:  time.Date(2026, 3, 16, 0, 0, 0, 0, time.UTC),
		},
		{
			name:  "friday from Sunday",
			input: "friday",
			want:  time.Date(2026, 3, 20, 0, 0, 0, 0, time.UTC),
		},
		{
			name:  "leap year date",
			input: "2024-02-29",
			want:  time.Date(2024, 2, 29, 0, 0, 0, 0, time.UTC),
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
		},
		{
			name:    "garbage input",
			input:   "not-a-date",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := parseDateTime(tt.input, time.UTC, refNow)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("parseDateTime(%q) expected error, got %v", tt.input, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("parseDateTime(%q) unexpected error: %v", tt.input, err)
			}
			if !got.Equal(tt.want) {
				t.Fatalf("parseDateTime(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseRelative(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  time.Time
		ok    bool
	}{
		{
			name:  "today",
			input: "today",
			want:  time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC),
			ok:    true,
		},
		{
			name:  "Tomorrow mixed case",
			input: "Tomorrow",
			want:  time.Date(2026, 3, 16, 0, 0, 0, 0, time.UTC),
			ok:    true,
		},
		{
			name:  "+1d",
			input: "+1d",
			want:  time.Date(2026, 3, 16, 0, 0, 0, 0, time.UTC),
			ok:    true,
		},
		{
			name:  "+7d",
			input: "+7d",
			want:  time.Date(2026, 3, 22, 0, 0, 0, 0, time.UTC),
			ok:    true,
		},
		{
			name:  "monday",
			input: "monday",
			want:  time.Date(2026, 3, 16, 0, 0, 0, 0, time.UTC),
			ok:    true,
		},
		{
			name:  "FRIDAY uppercase",
			input: "FRIDAY",
			want:  time.Date(2026, 3, 20, 0, 0, 0, 0, time.UTC),
			ok:    true,
		},
		{
			name:  "not a day",
			input: "notaday",
			ok:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, ok := parseRelative(tt.input, time.UTC, refNow)
			if ok != tt.ok {
				t.Fatalf("parseRelative(%q) ok = %v, want %v", tt.input, ok, tt.ok)
			}
			if tt.ok && !got.Equal(tt.want) {
				t.Fatalf("parseRelative(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseDuration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		want    time.Duration
		wantErr bool
	}{
		{name: "1 hour", input: "1h", want: time.Hour},
		{name: "30 minutes", input: "30m", want: 30 * time.Minute},
		{name: "1h30m", input: "1h30m", want: 90 * time.Minute},
		{name: "zero", input: "0s", wantErr: true},
		{name: "negative", input: "-1h", wantErr: true},
		{name: "invalid", input: "abc", wantErr: true},
		{name: "empty", input: "", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := parseDuration(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("parseDuration(%q) expected error, got %v", tt.input, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("parseDuration(%q) unexpected error: %v", tt.input, err)
			}
			if got != tt.want {
				t.Fatalf("parseDuration(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestBuildEventDateTime(t *testing.T) {
	t.Parallel()

	t.Run("timed event", func(t *testing.T) {
		t.Parallel()
		ts := time.Date(2026, 3, 15, 9, 0, 0, 0, time.UTC)
		edt := buildEventDateTime(ts, false, "America/Los_Angeles")
		if edt.DateTime == "" {
			t.Fatalf("expected DateTime to be set")
		}
		if edt.Date != "" {
			t.Fatalf("expected Date to be empty for timed event")
		}
		if edt.TimeZone != "America/Los_Angeles" {
			t.Fatalf("TimeZone = %q, want %q", edt.TimeZone, "America/Los_Angeles")
		}
	})

	t.Run("all-day event", func(t *testing.T) {
		t.Parallel()
		ts := time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC)
		edt := buildEventDateTime(ts, true, "")
		if edt.Date != "2026-03-15" {
			t.Fatalf("Date = %q, want %q", edt.Date, "2026-03-15")
		}
		if edt.DateTime != "" {
			t.Fatalf("expected DateTime to be empty for all-day event")
		}
		if edt.TimeZone != "" {
			t.Fatalf("expected empty TimeZone when not provided")
		}
	})

	t.Run("with timezone", func(t *testing.T) {
		t.Parallel()
		ts := time.Date(2026, 3, 15, 9, 0, 0, 0, time.UTC)
		edt := buildEventDateTime(ts, false, "Europe/London")
		if edt.TimeZone != "Europe/London" {
			t.Fatalf("TimeZone = %q, want %q", edt.TimeZone, "Europe/London")
		}
	})
}

func TestFormatEventTime(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		edt  *calendar.EventDateTime
		want string
	}{
		{
			name: "DateTime field",
			edt:  &calendar.EventDateTime{DateTime: "2026-03-15T09:00:00Z"},
			want: "Sun Mar 15, 2026 09:00 AM UTC",
		},
		{
			name: "Date field",
			edt:  &calendar.EventDateTime{Date: "2026-03-15"},
			want: "Sun Mar 15, 2026 (all day)",
		},
		{
			name: "nil input",
			edt:  nil,
			want: "",
		},
		{
			name: "empty EventDateTime",
			edt:  &calendar.EventDateTime{},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := formatEventTime(tt.edt, time.UTC)
			if got != tt.want {
				t.Fatalf("formatEventTime() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestStartOfDay(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		t    time.Time
		loc  *time.Location
		want time.Time
	}{
		{
			name: "midday UTC",
			t:    time.Date(2026, 3, 15, 14, 30, 45, 123, time.UTC),
			loc:  time.UTC,
			want: time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC),
		},
		{
			name: "already midnight",
			t:    time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC),
			loc:  time.UTC,
			want: time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := startOfDay(tt.t, tt.loc)
			if !got.Equal(tt.want) {
				t.Fatalf("startOfDay() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNextWeekday(t *testing.T) {
	t.Parallel()

	// refNow is Sunday March 15, 2026
	tests := []struct {
		name   string
		from   time.Time
		target time.Weekday
		want   time.Time
	}{
		{
			name:   "Sunday to Monday",
			from:   refNow,
			target: time.Monday,
			want:   time.Date(2026, 3, 16, 0, 0, 0, 0, time.UTC),
		},
		{
			name:   "Sunday to Friday",
			from:   refNow,
			target: time.Friday,
			want:   time.Date(2026, 3, 20, 0, 0, 0, 0, time.UTC),
		},
		{
			name:   "Sunday to Sunday wraps to next week",
			from:   refNow,
			target: time.Sunday,
			want:   time.Date(2026, 3, 22, 0, 0, 0, 0, time.UTC),
		},
		{
			name:   "Monday to Wednesday",
			from:   time.Date(2026, 3, 16, 10, 0, 0, 0, time.UTC),
			target: time.Wednesday,
			want:   time.Date(2026, 3, 18, 0, 0, 0, 0, time.UTC),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := nextWeekday(tt.from, tt.target)
			if !got.Equal(tt.want) {
				t.Fatalf("nextWeekday() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseDateTimeDST(t *testing.T) {
	t.Parallel()

	la, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		t.Fatalf("failed to load timezone: %v", err)
	}

	// March 8, 2026 is DST spring-forward in LA (2:00 AM -> 3:00 AM)
	dstNow := time.Date(2026, 3, 8, 12, 0, 0, 0, la)

	t.Run("time-only during DST transition day", func(t *testing.T) {
		t.Parallel()
		got, err := parseDateTime("09:00", la, dstNow)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := time.Date(2026, 3, 8, 9, 0, 0, 0, la)
		if !got.Equal(want) {
			t.Fatalf("got %v, want %v", got, want)
		}
	})

	t.Run("tomorrow across DST boundary", func(t *testing.T) {
		t.Parallel()
		got, err := parseDateTime("tomorrow", la, dstNow)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := time.Date(2026, 3, 9, 0, 0, 0, 0, la)
		if !got.Equal(want) {
			t.Fatalf("got %v, want %v", got, want)
		}
	})

	t.Run("date parsed in LA timezone", func(t *testing.T) {
		t.Parallel()
		got, err := parseDateTime("2026-03-08", la, dstNow)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := time.Date(2026, 3, 8, 0, 0, 0, 0, la)
		if !got.Equal(want) {
			t.Fatalf("got %v, want %v", got, want)
		}
	})
}
