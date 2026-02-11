package cmd

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/khang/google-suite-cli/internal/auth"
	"github.com/spf13/cobra"
	"google.golang.org/api/calendar/v3"
)

var (
	calendarID           string
	calendarMaxResults   int64
	calendarAfter        string
	calendarBefore       string
	calendarQuery        string
	calendarSingleEvents bool
	calendarOrderBy      string
	calendarTimezone     string
	calendarShowDeleted  bool

	// Write command flags (used by calendar_write.go)
	calendarSummary         string
	calendarStart           string
	calendarEnd             string
	calendarDuration        string
	calendarDescription     string
	calendarLocation        string
	calendarAttendees       string
	calendarAllDay          bool
	calendarRrule           string
	calendarSendUpdates     string
	calendarAddAttendees    string
	calendarRemoveAttendees string
	calendarRecurringScope  string
	calendarYes             bool
	calendarStatus          string
	calendarComment         string
)

var errDone = fmt.Errorf("done")

var calendarCmd = &cobra.Command{
	Use:   "calendar",
	Short: "Manage Google Calendar events",
	Long: `Commands for listing, creating, updating, and managing Google Calendar events.

Use the subcommands to interact with calendars and events for the authenticated user.`,
}

var calendarListCmd = &cobra.Command{
	Use:   "list",
	Short: "List upcoming calendar events",
	RunE:  runCalendarList,
}

var calendarGetCmd = &cobra.Command{
	Use:   "get <event-id>",
	Short: "Get details of a calendar event",
	Args:  cobra.ExactArgs(1),
	RunE:  runCalendarGet,
}

var calendarTodayCmd = &cobra.Command{
	Use:   "today",
	Short: "Show today's events",
	RunE:  runCalendarToday,
}

var calendarWeekCmd = &cobra.Command{
	Use:   "week",
	Short: "Show this week's events",
	RunE:  runCalendarWeek,
}

var calendarCalendarsCmd = &cobra.Command{
	Use:   "calendars",
	Short: "List available calendars",
	RunE:  runCalendarCalendars,
}

var calendarCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a calendar event",
	Long: `Create a new event in the specified calendar.

Required flags:
  --summary: Event title
  --start: Start time (supports various formats)

The --end flag or --duration flag specifies when the event ends.
If neither is provided, a 1-hour duration is assumed.
Use --all-day for all-day events (only date portion of --start is used).`,
	Example: `  # Create a 1-hour meeting
  gsuite calendar create --summary "Team Meeting" --start "2026-03-15 09:00"

  # Create a meeting with explicit duration
  gsuite calendar create --summary "Standup" --start "2026-03-15 09:00" --duration 30m

  # Create an all-day event
  gsuite calendar create --summary "Company Holiday" --start 2026-12-25 --all-day

  # Create a recurring weekly meeting
  gsuite calendar create --summary "1:1" --start "2026-03-15 10:00" --duration 30m --rrule "FREQ=WEEKLY;BYDAY=MO"

  # Create an event with attendees
  gsuite calendar create --summary "Review" --start "2026-03-15 14:00" --duration 1h --attendees "alice@example.com,bob@example.com" --send-updates all`,
}

var calendarUpdateCmd = &cobra.Command{
	Use:   "update <event-id>",
	Short: "Update a calendar event",
	Long: `Update an existing calendar event's properties.

Only the flags you provide will be changed; other fields remain unchanged.
Use --add-attendees and --remove-attendees to modify the attendee list.`,
	Example: `  # Change event title
  gsuite calendar update abc123 --summary "New Title"

  # Reschedule an event
  gsuite calendar update abc123 --start "2026-03-20 10:00" --end "2026-03-20 11:00"

  # Add attendees and notify them
  gsuite calendar update abc123 --add-attendees "carol@example.com" --send-updates all`,
	Args: cobra.ExactArgs(1),
}

var calendarDeleteCmd = &cobra.Command{
	Use:   "delete <event-id>",
	Short: "Delete a calendar event",
	Long: `Delete a calendar event by its ID.

Use --yes to skip the confirmation prompt.
For recurring events, --recurring-scope controls whether to delete
just this instance ("this") or all instances ("all").`,
	Example: `  # Delete an event (will prompt for confirmation)
  gsuite calendar delete abc123

  # Delete without confirmation
  gsuite calendar delete abc123 --yes

  # Delete all instances of a recurring event
  gsuite calendar delete abc123 --recurring-scope all --yes`,
	Args: cobra.ExactArgs(1),
}

var calendarRespondCmd = &cobra.Command{
	Use:   "respond <event-id>",
	Short: "Respond to a calendar event invitation",
	Long: `Set your RSVP status for a calendar event.

Required flags:
  --status: One of "accepted", "declined", or "tentative"`,
	Example: `  # Accept an invitation
  gsuite calendar respond abc123 --status accepted

  # Decline with a comment
  gsuite calendar respond abc123 --status declined --comment "Out of office"

  # Tentatively accept
  gsuite calendar respond abc123 --status tentative`,
	Args: cobra.ExactArgs(1),
}

func init() {
	rootCmd.AddCommand(calendarCmd)
	calendarCmd.AddCommand(calendarListCmd)
	calendarCmd.AddCommand(calendarGetCmd)
	calendarCmd.AddCommand(calendarTodayCmd)
	calendarCmd.AddCommand(calendarWeekCmd)
	calendarCmd.AddCommand(calendarCalendarsCmd)
	calendarCmd.AddCommand(calendarCreateCmd)
	calendarCmd.AddCommand(calendarUpdateCmd)
	calendarCmd.AddCommand(calendarDeleteCmd)
	calendarCmd.AddCommand(calendarRespondCmd)

	// List flags
	calendarListCmd.Flags().StringVar(&calendarID, "calendar-id", "primary", "Calendar ID")
	calendarListCmd.Flags().Int64VarP(&calendarMaxResults, "max-results", "n", 25, "Maximum number of events")
	calendarListCmd.Flags().StringVar(&calendarAfter, "after", "", "Show events after this time")
	calendarListCmd.Flags().StringVar(&calendarBefore, "before", "", "Show events before this time")
	calendarListCmd.Flags().StringVarP(&calendarQuery, "query", "q", "", "Search query")
	calendarListCmd.Flags().BoolVar(&calendarSingleEvents, "single-events", true, "Expand recurring events")
	calendarListCmd.Flags().StringVar(&calendarOrderBy, "order-by", "startTime", "Order by: startTime or updated")
	calendarListCmd.Flags().StringVar(&calendarTimezone, "timezone", "", "IANA timezone")
	calendarListCmd.Flags().BoolVar(&calendarShowDeleted, "show-deleted", false, "Show deleted events")

	// Get flags
	calendarGetCmd.Flags().StringVar(&calendarID, "calendar-id", "primary", "Calendar ID")
	calendarGetCmd.Flags().StringVar(&calendarTimezone, "timezone", "", "IANA timezone")

	// Today flags
	calendarTodayCmd.Flags().StringVar(&calendarID, "calendar-id", "primary", "Calendar ID")
	calendarTodayCmd.Flags().Int64VarP(&calendarMaxResults, "max-results", "n", 25, "Maximum number of events")
	calendarTodayCmd.Flags().StringVar(&calendarTimezone, "timezone", "", "IANA timezone")

	// Week flags
	calendarWeekCmd.Flags().StringVar(&calendarID, "calendar-id", "primary", "Calendar ID")
	calendarWeekCmd.Flags().Int64VarP(&calendarMaxResults, "max-results", "n", 25, "Maximum number of events")
	calendarWeekCmd.Flags().StringVar(&calendarTimezone, "timezone", "", "IANA timezone")

	// Calendars flags
	calendarCalendarsCmd.Flags().Int64VarP(&calendarMaxResults, "max-results", "n", 100, "Maximum number of calendars")

	// Create flags
	calendarCreateCmd.Flags().StringVar(&calendarSummary, "summary", "", "Event title (required)")
	calendarCreateCmd.Flags().StringVar(&calendarStart, "start", "", "Start time (required)")
	calendarCreateCmd.Flags().StringVar(&calendarEnd, "end", "", "End time")
	calendarCreateCmd.Flags().StringVarP(&calendarDuration, "duration", "d", "", "Duration (e.g., 1h, 30m)")
	calendarCreateCmd.Flags().StringVar(&calendarDescription, "description", "", "Event description")
	calendarCreateCmd.Flags().StringVarP(&calendarLocation, "location", "l", "", "Event location")
	calendarCreateCmd.Flags().StringVar(&calendarAttendees, "attendees", "", "Comma-separated attendee emails")
	calendarCreateCmd.Flags().BoolVar(&calendarAllDay, "all-day", false, "Create all-day event")
	calendarCreateCmd.Flags().StringVar(&calendarRrule, "rrule", "", "Recurrence rule (e.g., FREQ=WEEKLY;BYDAY=MO,WE,FR)")
	calendarCreateCmd.Flags().StringVar(&calendarSendUpdates, "send-updates", "none", "Send notifications: all, externalOnly, none")
	calendarCreateCmd.Flags().StringVar(&calendarTimezone, "timezone", "", "IANA timezone for the event")
	calendarCreateCmd.Flags().StringVar(&calendarID, "calendar-id", "primary", "Calendar ID")
	calendarCreateCmd.MarkFlagRequired("summary")
	calendarCreateCmd.MarkFlagRequired("start")

	// Update flags
	calendarUpdateCmd.Flags().StringVar(&calendarSummary, "summary", "", "New event title")
	calendarUpdateCmd.Flags().StringVar(&calendarStart, "start", "", "New start time")
	calendarUpdateCmd.Flags().StringVar(&calendarEnd, "end", "", "New end time")
	calendarUpdateCmd.Flags().StringVar(&calendarDescription, "description", "", "New description")
	calendarUpdateCmd.Flags().StringVarP(&calendarLocation, "location", "l", "", "New location")
	calendarUpdateCmd.Flags().StringVar(&calendarAddAttendees, "add-attendees", "", "Comma-separated emails to add")
	calendarUpdateCmd.Flags().StringVar(&calendarRemoveAttendees, "remove-attendees", "", "Comma-separated emails to remove")
	calendarUpdateCmd.Flags().StringVar(&calendarSendUpdates, "send-updates", "none", "Send notifications: all, externalOnly, none")
	calendarUpdateCmd.Flags().StringVar(&calendarTimezone, "timezone", "", "IANA timezone")
	calendarUpdateCmd.Flags().StringVar(&calendarID, "calendar-id", "primary", "Calendar ID")
	calendarUpdateCmd.Flags().StringVar(&calendarRecurringScope, "recurring-scope", "this", "Recurring event scope: this, all")

	// Delete flags
	calendarDeleteCmd.Flags().StringVar(&calendarSendUpdates, "send-updates", "none", "Send notifications: all, externalOnly, none")
	calendarDeleteCmd.Flags().StringVar(&calendarRecurringScope, "recurring-scope", "this", "Recurring event scope: this, all")
	calendarDeleteCmd.Flags().BoolVar(&calendarYes, "yes", false, "Confirm destructive operations")
	calendarDeleteCmd.Flags().StringVar(&calendarID, "calendar-id", "primary", "Calendar ID")

	// Respond flags
	calendarRespondCmd.Flags().StringVar(&calendarStatus, "status", "", "Response status: accepted, declined, tentative (required)")
	calendarRespondCmd.Flags().StringVar(&calendarComment, "comment", "", "RSVP comment")
	calendarRespondCmd.Flags().StringVar(&calendarSendUpdates, "send-updates", "none", "Send notifications: all, externalOnly, none")
	calendarRespondCmd.Flags().StringVar(&calendarID, "calendar-id", "primary", "Calendar ID")
	calendarRespondCmd.MarkFlagRequired("status")
}

func resolveTimezone() (*time.Location, error) {
	if calendarTimezone != "" {
		loc, err := time.LoadLocation(calendarTimezone)
		if err != nil {
			return nil, fmt.Errorf("invalid timezone %q: %w", calendarTimezone, err)
		}
		return loc, nil
	}
	return time.Now().Location(), nil
}

func listCalendarEvents(cmd *cobra.Command, calID string, timeMin, timeMax time.Time, maxResults int64, query string, singleEvents bool, orderBy string, tz *time.Location, showDeleted bool) error {
	ctx := context.Background()

	service, err := auth.NewCalendarService(ctx, GetAccountEmail())
	if err != nil {
		return auth.HandleCalendarError(err, "authentication failed")
	}

	call := service.Events.List(calID).
		TimeMin(timeMin.Format(time.RFC3339)).
		TimeMax(timeMax.Format(time.RFC3339)).
		SingleEvents(singleEvents).
		ShowDeleted(showDeleted).
		Fields("items(id,summary,start,end,location,status,recurringEventId),nextPageToken")

	if orderBy != "" {
		call = call.OrderBy(orderBy)
	}
	if query != "" {
		call = call.Q(query)
	}
	if calendarTimezone != "" {
		call = call.TimeZone(calendarTimezone)
	}

	call.MaxResults(min(maxResults, 250))

	var allEvents []*calendar.Event
	err = call.Pages(ctx, func(page *calendar.Events) error {
		allEvents = append(allEvents, page.Items...)
		if int64(len(allEvents)) >= maxResults {
			return errDone
		}
		return nil
	})
	if err != nil && err != errDone {
		return auth.HandleCalendarError(err, "failed to list events")
	}
	if int64(len(allEvents)) > maxResults {
		allEvents = allEvents[:maxResults]
	}

	if GetOutputFormat() == "json" {
		type eventListItem struct {
			ID               string `json:"id"`
			Summary          string `json:"summary"`
			Start            string `json:"start"`
			End              string `json:"end"`
			Location         string `json:"location"`
			Status           string `json:"status"`
			AllDay           bool   `json:"all_day"`
			Recurring        bool   `json:"recurring"`
			RecurringEventID string `json:"recurring_event_id"`
		}

		if len(allEvents) == 0 {
			return outputJSON([]struct{}{})
		}

		items := make([]eventListItem, len(allEvents))
		for i, ev := range allEvents {
			isAllDay := ev.Start != nil && ev.Start.Date != ""
			items[i] = eventListItem{
				ID:               ev.Id,
				Summary:          ev.Summary,
				Start:            formatEventTime(ev.Start, tz),
				End:              formatEventTime(ev.End, tz),
				Location:         ev.Location,
				Status:           ev.Status,
				AllDay:           isAllDay,
				Recurring:        ev.RecurringEventId != "",
				RecurringEventID: ev.RecurringEventId,
			}
		}
		return outputJSON(items)
	}

	if len(allEvents) == 0 {
		fmt.Println("No events found.")
		return nil
	}

	printEventTable(allEvents, tz)
	return nil
}

func printEventTable(events []*calendar.Event, tz *time.Location) {
	fmt.Printf("%-12s %-20s %s\n", "DATE", "TIME", "SUMMARY")
	fmt.Printf("%-12s %-20s %s\n", "----", "----", "-------")

	for _, ev := range events {
		date, timeRange := formatEventTableRow(ev, tz)
		summary := ev.Summary
		if ev.RecurringEventId != "" {
			summary += " (recurring)"
		}
		fmt.Printf("%-12s %-20s %s\n", date, timeRange, summary)
	}

	fmt.Printf("\n[%d event(s)]\n", len(events))
}

func formatEventTableRow(ev *calendar.Event, tz *time.Location) (string, string) {
	if ev.Start == nil {
		return "", ""
	}

	if ev.Start.Date != "" {
		t, err := time.Parse("2006-01-02", ev.Start.Date)
		if err != nil {
			return ev.Start.Date, "all day"
		}
		return t.Format("2006-01-02"), "all day"
	}

	if ev.Start.DateTime != "" {
		startT, err := time.Parse(time.RFC3339, ev.Start.DateTime)
		if err != nil {
			return "", ev.Start.DateTime
		}
		startT = startT.In(tz)
		date := startT.Format("2006-01-02")
		startStr := startT.Format("03:04 PM")

		endStr := ""
		if ev.End != nil && ev.End.DateTime != "" {
			endT, err := time.Parse(time.RFC3339, ev.End.DateTime)
			if err == nil {
				endStr = endT.In(tz).Format("03:04 PM")
			}
		}

		if endStr != "" {
			return date, startStr + " - " + endStr
		}
		return date, startStr
	}

	return "", ""
}

func runCalendarList(cmd *cobra.Command, args []string) error {
	tz, err := resolveTimezone()
	if err != nil {
		return err
	}

	now := time.Now().In(tz)

	timeMin := now
	if calendarAfter != "" {
		t, err := parseDateTime(calendarAfter, tz, now)
		if err != nil {
			return fmt.Errorf("invalid --after value: %w", err)
		}
		timeMin = t
	}

	timeMax := now.AddDate(0, 0, 30)
	if calendarBefore != "" {
		t, err := parseDateTime(calendarBefore, tz, now)
		if err != nil {
			return fmt.Errorf("invalid --before value: %w", err)
		}
		timeMax = t
	}

	return listCalendarEvents(cmd, calendarID, timeMin, timeMax, calendarMaxResults, calendarQuery, calendarSingleEvents, calendarOrderBy, tz, calendarShowDeleted)
}

func runCalendarGet(cmd *cobra.Command, args []string) error {
	eventID := args[0]

	tz, err := resolveTimezone()
	if err != nil {
		return err
	}

	ctx := context.Background()
	service, err := auth.NewCalendarService(ctx, GetAccountEmail())
	if err != nil {
		return auth.HandleCalendarError(err, "authentication failed")
	}

	ev, err := service.Events.Get(calendarID, eventID).Do()
	if err != nil {
		return auth.HandleCalendarError(err, "failed to get event")
	}

	if GetOutputFormat() == "json" {
		type attendeeItem struct {
			Email          string `json:"email"`
			DisplayName    string `json:"display_name"`
			ResponseStatus string `json:"response_status"`
			Organizer      bool   `json:"organizer"`
			Self           bool   `json:"self"`
		}
		type eventDetail struct {
			ID               string         `json:"id"`
			Summary          string         `json:"summary"`
			Start            string         `json:"start"`
			End              string         `json:"end"`
			Status           string         `json:"status"`
			Location         string         `json:"location"`
			Description      string         `json:"description"`
			Recurrence       []string       `json:"recurrence"`
			RecurringEventID string         `json:"recurring_event_id"`
			Attendees        []attendeeItem `json:"attendees"`
			HtmlLink         string         `json:"html_link"`
			Creator          string         `json:"creator"`
			Organizer        string         `json:"organizer"`
		}

		detail := eventDetail{
			ID:               ev.Id,
			Summary:          ev.Summary,
			Start:            formatEventTime(ev.Start, tz),
			End:              formatEventTime(ev.End, tz),
			Status:           ev.Status,
			Location:         ev.Location,
			Description:      ev.Description,
			Recurrence:       ev.Recurrence,
			RecurringEventID: ev.RecurringEventId,
			HtmlLink:         ev.HtmlLink,
		}

		if ev.Creator != nil {
			detail.Creator = ev.Creator.Email
		}
		if ev.Organizer != nil {
			detail.Organizer = ev.Organizer.Email
		}

		if len(ev.Attendees) > 0 {
			detail.Attendees = make([]attendeeItem, len(ev.Attendees))
			for i, a := range ev.Attendees {
				detail.Attendees[i] = attendeeItem{
					Email:          a.Email,
					DisplayName:    a.DisplayName,
					ResponseStatus: a.ResponseStatus,
					Organizer:      a.Organizer,
					Self:           a.Self,
				}
			}
		} else {
			detail.Attendees = []attendeeItem{}
		}

		if detail.Recurrence == nil {
			detail.Recurrence = []string{}
		}

		return outputJSON(detail)
	}

	// Text output
	fmt.Printf("Event: %s\n", ev.Summary)
	fmt.Printf("ID:    %s\n", ev.Id)
	fmt.Printf("Start: %s\n", formatEventTime(ev.Start, tz))
	fmt.Printf("End:   %s\n", formatEventTime(ev.End, tz))
	fmt.Printf("Status: %s\n", ev.Status)

	if ev.Location != "" {
		fmt.Printf("Location: %s\n", ev.Location)
	}
	if ev.Description != "" {
		fmt.Printf("Description: %s\n", ev.Description)
	}
	if ev.Creator != nil {
		fmt.Printf("Creator: %s\n", ev.Creator.Email)
	}
	if ev.Organizer != nil {
		fmt.Printf("Organizer: %s\n", ev.Organizer.Email)
	}
	if len(ev.Recurrence) > 0 {
		fmt.Printf("Recurrence: %s\n", strings.Join(ev.Recurrence, ", "))
	}
	if ev.RecurringEventId != "" {
		fmt.Printf("Recurring Event ID: %s\n", ev.RecurringEventId)
	}
	if ev.HtmlLink != "" {
		fmt.Printf("Link: %s\n", ev.HtmlLink)
	}

	if len(ev.Attendees) > 0 {
		fmt.Printf("\nAttendees:\n")
		for _, a := range ev.Attendees {
			name := a.Email
			if a.DisplayName != "" {
				name = fmt.Sprintf("%s <%s>", a.DisplayName, a.Email)
			}
			status := a.ResponseStatus
			if a.Organizer {
				status += " (organizer)"
			}
			if a.Self {
				status += " (you)"
			}
			fmt.Printf("  - %s [%s]\n", name, status)
		}
	}

	return nil
}

func runCalendarToday(cmd *cobra.Command, args []string) error {
	tz, err := resolveTimezone()
	if err != nil {
		return err
	}

	now := time.Now().In(tz)
	dayStart := startOfDay(now, tz)
	dayEnd := dayStart.AddDate(0, 0, 1)

	return listCalendarEvents(cmd, calendarID, dayStart, dayEnd, calendarMaxResults, "", true, "startTime", tz, false)
}

func runCalendarWeek(cmd *cobra.Command, args []string) error {
	tz, err := resolveTimezone()
	if err != nil {
		return err
	}

	now := time.Now().In(tz)
	// Find Monday of the current week
	daysFromMonday := (int(now.Weekday()) - int(time.Monday) + 7) % 7
	weekStart := startOfDay(now.AddDate(0, 0, -daysFromMonday), tz)
	weekEnd := weekStart.AddDate(0, 0, 7)

	return listCalendarEvents(cmd, calendarID, weekStart, weekEnd, calendarMaxResults, "", true, "startTime", tz, false)
}

func runCalendarCalendars(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	service, err := auth.NewCalendarService(ctx, GetAccountEmail())
	if err != nil {
		return auth.HandleCalendarError(err, "authentication failed")
	}

	resp, err := service.CalendarList.List().
		MaxResults(calendarMaxResults).
		Do()
	if err != nil {
		return auth.HandleCalendarError(err, "failed to list calendars")
	}

	if GetOutputFormat() == "json" {
		type calendarItem struct {
			ID         string `json:"id"`
			Summary    string `json:"summary"`
			AccessRole string `json:"access_role"`
			Primary    bool   `json:"primary"`
			Timezone   string `json:"timezone"`
		}

		if len(resp.Items) == 0 {
			return outputJSON([]struct{}{})
		}

		items := make([]calendarItem, len(resp.Items))
		for i, cal := range resp.Items {
			items[i] = calendarItem{
				ID:         cal.Id,
				Summary:    cal.Summary,
				AccessRole: cal.AccessRole,
				Primary:    cal.Primary,
				Timezone:   cal.TimeZone,
			}
		}
		return outputJSON(items)
	}

	if len(resp.Items) == 0 {
		fmt.Println("No calendars found.")
		return nil
	}

	fmt.Printf("%-40s %-30s %-15s %s\n", "ID", "NAME", "ROLE", "TIMEZONE")
	fmt.Printf("%-40s %-30s %-15s %s\n", "--", "----", "----", "--------")

	for _, cal := range resp.Items {
		name := cal.Summary
		if cal.Primary {
			name += " (primary)"
		}
		fmt.Printf("%-40s %-30s %-15s %s\n", cal.Id, name, cal.AccessRole, cal.TimeZone)
	}

	fmt.Printf("\n[%d calendar(s)]\n", len(resp.Items))
	return nil
}
