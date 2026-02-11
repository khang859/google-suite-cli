package cmd

import (
	"context"
	"fmt"
	"net/mail"
	"strings"
	"time"

	"github.com/khang/google-suite-cli/internal/auth"
	"github.com/spf13/cobra"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/googleapi"
)

func init() {
	calendarCreateCmd.RunE = runCalendarCreate
	calendarUpdateCmd.RunE = runCalendarUpdate
	calendarDeleteCmd.RunE = runCalendarDelete
	calendarRespondCmd.RunE = runCalendarRespond
}

func runCalendarCreate(cmd *cobra.Command, args []string) error {
	if err := validateCalendarCreateFlags(calendarStart, calendarEnd, calendarDuration, calendarAllDay); err != nil {
		return err
	}

	tz, err := resolveTimezone()
	if err != nil {
		return err
	}

	now := time.Now().In(tz)
	startTime, err := parseDateTime(calendarStart, tz, now)
	if err != nil {
		return fmt.Errorf("invalid --start value: %w", err)
	}

	var endTime time.Time
	switch {
	case calendarEnd != "":
		endTime, err = parseDateTime(calendarEnd, tz, now)
		if err != nil {
			return fmt.Errorf("invalid --end value: %w", err)
		}
	case calendarDuration != "":
		d, err := parseDuration(calendarDuration)
		if err != nil {
			return err
		}
		endTime = startTime.Add(d)
	case calendarAllDay:
		endTime = startTime.AddDate(0, 0, 1)
	default:
		endTime = startTime.Add(time.Hour)
	}

	event := &calendar.Event{
		Summary:     calendarSummary,
		Description: calendarDescription,
		Location:    calendarLocation,
		Start:       buildEventDateTime(startTime, calendarAllDay, calendarTimezone),
		End:         buildEventDateTime(endTime, calendarAllDay, calendarTimezone),
	}

	if calendarRrule != "" {
		event.Recurrence = []string{"RRULE:" + calendarRrule}
		if event.Start.TimeZone == "" {
			event.Start.TimeZone = tz.String()
		}
		if event.End.TimeZone == "" {
			event.End.TimeZone = tz.String()
		}
	}

	if calendarAttendees != "" {
		emails, err := validateAttendeeEmails(calendarAttendees)
		if err != nil {
			return err
		}
		attendees := make([]*calendar.EventAttendee, len(emails))
		for i, email := range emails {
			attendees[i] = &calendar.EventAttendee{Email: email}
		}
		event.Attendees = attendees
	}

	ctx := context.Background()
	service, err := auth.NewCalendarService(ctx, GetAccountEmail())
	if err != nil {
		return auth.HandleCalendarError(err, "authentication failed")
	}

	result, err := service.Events.Insert(calendarID, event).SendUpdates(calendarSendUpdates).Do()
	if err != nil {
		return auth.HandleCalendarError(err, "failed to create event")
	}

	if GetOutputFormat() == "json" {
		type createResult struct {
			ID       string `json:"id"`
			Summary  string `json:"summary"`
			HtmlLink string `json:"html_link"`
			Start    string `json:"start"`
			End      string `json:"end"`
		}
		return outputJSON(createResult{
			ID:       result.Id,
			Summary:  result.Summary,
			HtmlLink: result.HtmlLink,
			Start:    formatEventTime(result.Start, tz),
			End:      formatEventTime(result.End, tz),
		})
	}

	fmt.Printf("Event created: %s\n", result.Id)
	fmt.Printf("Link: %s\n", result.HtmlLink)
	return nil
}

func runCalendarUpdate(cmd *cobra.Command, args []string) error {
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

	event, err := service.Events.Get(calendarID, eventID).Do()
	if err != nil {
		return auth.HandleCalendarError(err, "failed to get event")
	}

	// For recurring events with --recurring-scope all, operate on the parent
	if calendarRecurringScope == "all" && event.RecurringEventId != "" {
		eventID = event.RecurringEventId
		event, err = service.Events.Get(calendarID, eventID).Do()
		if err != nil {
			return auth.HandleCalendarError(err, "failed to get recurring event")
		}
	}

	now := time.Now().In(tz)

	if cmd.Flags().Changed("summary") {
		event.Summary = calendarSummary
	}

	if cmd.Flags().Changed("start") {
		startTime, err := parseDateTime(calendarStart, tz, now)
		if err != nil {
			return fmt.Errorf("invalid --start value: %w", err)
		}
		isAllDay := event.Start != nil && event.Start.Date != ""
		event.Start = buildEventDateTime(startTime, isAllDay, calendarTimezone)
	}

	if cmd.Flags().Changed("end") {
		endTime, err := parseDateTime(calendarEnd, tz, now)
		if err != nil {
			return fmt.Errorf("invalid --end value: %w", err)
		}
		isAllDay := event.End != nil && event.End.Date != ""
		event.End = buildEventDateTime(endTime, isAllDay, calendarTimezone)
	}

	if cmd.Flags().Changed("description") {
		if calendarDescription == "" {
			event.NullFields = append(event.NullFields, "Description")
		} else {
			event.Description = calendarDescription
		}
	}

	if cmd.Flags().Changed("location") {
		if calendarLocation == "" {
			event.NullFields = append(event.NullFields, "Location")
		} else {
			event.Location = calendarLocation
		}
	}

	if cmd.Flags().Changed("timezone") {
		if event.Start != nil {
			event.Start.TimeZone = calendarTimezone
		}
		if event.End != nil {
			event.End.TimeZone = calendarTimezone
		}
	}

	if calendarAddAttendees != "" {
		emails, err := validateAttendeeEmails(calendarAddAttendees)
		if err != nil {
			return err
		}
		for _, email := range emails {
			event.Attendees = append(event.Attendees, &calendar.EventAttendee{Email: email})
		}
	}

	if calendarRemoveAttendees != "" {
		emails, err := validateAttendeeEmails(calendarRemoveAttendees)
		if err != nil {
			return err
		}
		removeSet := make(map[string]bool, len(emails))
		for _, email := range emails {
			removeSet[strings.ToLower(email)] = true
		}
		filtered := make([]*calendar.EventAttendee, 0, len(event.Attendees))
		for _, a := range event.Attendees {
			if !removeSet[strings.ToLower(a.Email)] {
				filtered = append(filtered, a)
			}
		}
		event.Attendees = filtered
	}

	event.ServerResponse = googleapi.ServerResponse{}

	result, err := service.Events.Update(calendarID, eventID, event).SendUpdates(calendarSendUpdates).Do()
	if err != nil {
		return auth.HandleCalendarError(err, "failed to update event")
	}

	if GetOutputFormat() == "json" {
		type updateResult struct {
			ID       string `json:"id"`
			Summary  string `json:"summary"`
			HtmlLink string `json:"html_link"`
		}
		return outputJSON(updateResult{
			ID:       result.Id,
			Summary:  result.Summary,
			HtmlLink: result.HtmlLink,
		})
	}

	fmt.Printf("Event updated: %s\n", result.Id)
	return nil
}

func runCalendarDelete(cmd *cobra.Command, args []string) error {
	eventID := args[0]

	if calendarRecurringScope == "all" && !calendarYes {
		return fmt.Errorf("this will delete ALL instances of this recurring event. Use --yes to confirm, or --recurring-scope this to delete only this instance")
	}

	ctx := context.Background()
	service, err := auth.NewCalendarService(ctx, GetAccountEmail())
	if err != nil {
		return auth.HandleCalendarError(err, "authentication failed")
	}

	// For recurring events with --recurring-scope all, operate on the parent
	if calendarRecurringScope == "all" {
		event, err := service.Events.Get(calendarID, eventID).Do()
		if err != nil {
			return auth.HandleCalendarError(err, "failed to get event")
		}
		if event.RecurringEventId != "" {
			eventID = event.RecurringEventId
		}
	}

	err = service.Events.Delete(calendarID, eventID).SendUpdates(calendarSendUpdates).Do()
	if err != nil {
		return auth.HandleCalendarError(err, "failed to delete event")
	}

	if GetOutputFormat() == "json" {
		type deleteResult struct {
			ID      string `json:"id"`
			Deleted bool   `json:"deleted"`
		}
		return outputJSON(deleteResult{
			ID:      eventID,
			Deleted: true,
		})
	}

	fmt.Printf("Event deleted: %s\n", eventID)
	return nil
}

func runCalendarRespond(cmd *cobra.Command, args []string) error {
	eventID := args[0]

	validStatuses := map[string]bool{
		"accepted":  true,
		"declined":  true,
		"tentative": true,
	}
	if !validStatuses[calendarStatus] {
		return fmt.Errorf("invalid --status %q: must be one of accepted, declined, tentative", calendarStatus)
	}

	ctx := context.Background()
	service, err := auth.NewCalendarService(ctx, GetAccountEmail())
	if err != nil {
		return auth.HandleCalendarError(err, "authentication failed")
	}

	event, err := service.Events.Get(calendarID, eventID).Do()
	if err != nil {
		return auth.HandleCalendarError(err, "failed to get event")
	}

	var selfAttendee *calendar.EventAttendee
	for _, a := range event.Attendees {
		if a.Self {
			selfAttendee = a
			break
		}
	}
	if selfAttendee == nil {
		return fmt.Errorf("you are not listed as an attendee of this event")
	}

	selfAttendee.ResponseStatus = calendarStatus
	if calendarComment != "" {
		selfAttendee.Comment = calendarComment
	}

	patchEvent := &calendar.Event{
		Attendees: event.Attendees,
	}

	result, err := service.Events.Patch(calendarID, eventID, patchEvent).
		SendUpdates(calendarSendUpdates).
		Do()
	if err != nil {
		return auth.HandleCalendarError(err, "failed to update RSVP")
	}

	if GetOutputFormat() == "json" {
		type respondResult struct {
			ID     string `json:"id"`
			Status string `json:"status"`
		}
		return outputJSON(respondResult{
			ID:     result.Id,
			Status: calendarStatus,
		})
	}

	fmt.Printf("RSVP updated: %s for %s\n", calendarStatus, result.Id)
	return nil
}

func validateCalendarCreateFlags(start, end, duration string, allDay bool) error {
	if end != "" && duration != "" {
		return fmt.Errorf("--end and --duration are mutually exclusive")
	}
	if allDay && duration != "" {
		return fmt.Errorf("--all-day and --duration cannot be combined")
	}
	return nil
}

func validateAttendeeEmails(csv string) ([]string, error) {
	parts := strings.Split(csv, ",")
	emails := make([]string, 0, len(parts))

	for _, part := range parts {
		email := strings.TrimSpace(part)
		if email == "" {
			continue
		}

		addr, err := mail.ParseAddress(email)
		if err != nil {
			// Try wrapping with angle brackets for bare emails
			addr, err = mail.ParseAddress("<" + email + ">")
			if err != nil {
				return nil, fmt.Errorf("invalid email address %q: %w", email, err)
			}
		}
		emails = append(emails, addr.Address)
	}

	return emails, nil
}
