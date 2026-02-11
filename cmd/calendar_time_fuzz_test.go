package cmd

import (
	"testing"
	"time"
)

func FuzzParseDateTime(f *testing.F) {
	f.Add("2026-03-15T09:00:00-07:00")
	f.Add("2026-03-15")
	f.Add("2026-03-15 09:00")
	f.Add("2026-03-15T09:00:00")
	f.Add("09:00")
	f.Add("today")
	f.Add("tomorrow")
	f.Add("+3d")
	f.Add("monday")
	f.Add("friday")
	f.Add("")
	f.Add("not-a-date")
	f.Add("2024-02-29")

	f.Fuzz(func(t *testing.T, input string) {
		now := time.Date(2026, 3, 15, 12, 0, 0, 0, time.UTC)
		// Should never panic
		parseDateTime(input, time.UTC, now)
	})
}

func FuzzParseDuration(f *testing.F) {
	f.Add("1h")
	f.Add("30m")
	f.Add("1h30m")
	f.Add("0s")
	f.Add("-1h")
	f.Add("abc")
	f.Add("")

	f.Fuzz(func(t *testing.T, input string) {
		// Should never panic
		parseDuration(input)
	})
}
