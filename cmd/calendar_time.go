package cmd

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"google.golang.org/api/calendar/v3"
)

var relDaysRegexp = regexp.MustCompile(`^\+(\d+)d$`)

var dayNames = map[string]time.Weekday{
	"monday":    time.Monday,
	"tuesday":   time.Tuesday,
	"wednesday": time.Wednesday,
	"thursday":  time.Thursday,
	"friday":    time.Friday,
	"saturday":  time.Saturday,
	"sunday":    time.Sunday,
}

func parseDateTime(input string, loc *time.Location, now time.Time) (time.Time, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return time.Time{}, fmt.Errorf("empty datetime input; accepted formats: RFC3339, 2006-01-02, 2006-01-02 15:04, 2006-01-02T15:04:05, 15:04, today, tomorrow, monday-sunday, +Nd")
	}

	if t, ok := parseRelative(input, loc, now); ok {
		return t, nil
	}

	if t, err := time.Parse(time.RFC3339, input); err == nil {
		return t, nil
	}

	formats := []string{
		"2006-01-02",
		"2006-01-02 15:04",
		"2006-01-02T15:04:05",
	}
	for _, f := range formats {
		if t, err := time.ParseInLocation(f, input, loc); err == nil {
			return t, nil
		}
	}

	// Time-only: anchor to today
	if t, err := time.ParseInLocation("15:04", input, loc); err == nil {
		return time.Date(now.Year(), now.Month(), now.Day(), t.Hour(), t.Minute(), 0, 0, loc), nil
	}

	return time.Time{}, fmt.Errorf("cannot parse %q; accepted formats: RFC3339, 2006-01-02, 2006-01-02 15:04, 2006-01-02T15:04:05, 15:04, today, tomorrow, monday-sunday, +Nd", input)
}

func parseRelative(input string, loc *time.Location, now time.Time) (time.Time, bool) {
	lower := strings.ToLower(strings.TrimSpace(input))

	switch lower {
	case "today":
		return startOfDay(now, loc), true
	case "tomorrow":
		return startOfDay(now.AddDate(0, 0, 1), loc), true
	}

	if wd, ok := dayNames[lower]; ok {
		return nextWeekday(now.In(loc), wd), true
	}

	if m := relDaysRegexp.FindStringSubmatch(lower); m != nil {
		n := 0
		for _, c := range m[1] {
			n = n*10 + int(c-'0')
		}
		return startOfDay(now.AddDate(0, 0, n), loc), true
	}

	return time.Time{}, false
}

func parseDuration(input string) (time.Duration, error) {
	d, err := time.ParseDuration(input)
	if err != nil {
		return 0, fmt.Errorf("invalid duration %q: %w", input, err)
	}
	if d <= 0 {
		return 0, fmt.Errorf("duration must be positive, got %s", d)
	}
	return d, nil
}

func buildEventDateTime(t time.Time, allDay bool, tz string) *calendar.EventDateTime {
	edt := &calendar.EventDateTime{}
	if allDay {
		edt.Date = t.Format("2006-01-02")
	} else {
		edt.DateTime = t.Format(time.RFC3339)
	}
	if tz != "" {
		edt.TimeZone = tz
	}
	return edt
}

func formatEventTime(edt *calendar.EventDateTime, displayTz *time.Location) string {
	if edt == nil {
		return ""
	}
	if edt.Date != "" {
		t, err := time.Parse("2006-01-02", edt.Date)
		if err != nil {
			return edt.Date
		}
		return t.Format("Mon Jan 02, 2006 (all day)")
	}
	if edt.DateTime != "" {
		t, err := time.Parse(time.RFC3339, edt.DateTime)
		if err != nil {
			return edt.DateTime
		}
		return t.In(displayTz).Format("Mon Jan 02, 2006 03:04 PM MST")
	}
	return ""
}

func startOfDay(t time.Time, loc *time.Location) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, loc)
}

func nextWeekday(from time.Time, target time.Weekday) time.Time {
	days := int(target) - int(from.Weekday())
	if days <= 0 {
		days += 7
	}
	return startOfDay(from.AddDate(0, 0, days), from.Location())
}
