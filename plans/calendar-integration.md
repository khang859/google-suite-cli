# feat: Add Google Calendar Integration

## Enhancement Summary

**Deepened on:** 2026-02-10
**Research agents used:** 8 (architecture-strategist, performance-oracle, security-sentinel, code-simplicity-reviewer, pattern-recognition-specialist, go-testing-researcher, date-time-parser-researcher, calendar-api-patterns-researcher)

### Key Changes from Original Plan
1. **Scope narrowed**: Use `CalendarEventsScope` + `CalendarReadonlyScope` instead of `CalendarScope` (principle of least privilege)
2. **`--send-updates` default changed**: `"none"` instead of `"all"` (prevents accidental email sends in automation)
3. **Flag naming fixed**: `--max-results` instead of `--max`, `--after`/`--before` instead of `--from`/`--to` (pattern consistency)
4. **File count reduced**: 3 files instead of 6 (matches codebase convention of single file per resource)
5. **JSON `omitempty` removed**: Never used in existing codebase -- always include all fields
6. **`-s` shorthand dropped**: Conflicts with `--subject` in drafts/send commands
7. **Delete safety added**: `--yes` flag required for `--recurring-scope all` delete
8. **Attendee email validation added**: `net/mail.ParseAddress()` before API calls
9. **Date parser designed for testability**: Accepts `now` and `loc` parameters, never calls `time.Now()` directly
10. **Proactive scope checking**: Detect missing calendar scope before API call, not via reactive 403

### Reviewer Consensus on Conflicts
- **Get+Update vs Patch**: Security says use Patch (avoids TOCTOU), Performance says Get+Update (2 quota units vs 3). **Decision: Use Get+Update for `update`, Patch for `respond`** (respond is a single-field change where Patch is simpler and more atomic).
- **File splitting**: Simplicity says 2 files, Architecture says 3, Pattern says match existing (1 per resource). **Decision: 3 files** (`calendar.go`, `calendar_write.go`, `calendar_time.go`) since the combined size would exceed 500 lines.
- **MVP scope**: Simplicity says drop today/week/calendars/respond. **Decision: Keep all 9 subcommands** -- they are thin wrappers that add significant usability. But `--recurring-scope following` is deferred (requires complex 4-step API dance).

---

## Overview

Add Google Calendar support to the gsuite CLI, enabling users to list, create, update, delete, and respond to calendar events alongside existing Gmail functionality. This transforms the tool from a Gmail CLI into a Google Workspace CLI.

Calendar is explicitly listed as a "What's next" area in `.planning/MILESTONES.md` (lines 76, 102, 127) and was previously "Out of Scope" in `.planning/PROJECT.md` (line 40).

## Problem Statement / Motivation

Users of the gsuite CLI currently have no way to manage their Google Calendar from the command line. Calendar management is a natural companion to email -- checking today's meetings, creating events, RSVPing to invitations -- all common workflows for power users and automation scripts. Adding calendar support fulfills the project's stated roadmap and makes the CLI significantly more useful for daily workflows.

## Proposed Solution

Add a `gsuite calendar` parent command with subcommands following the established Cobra command pattern. Extend the auth layer to support Calendar API scope alongside Gmail. No new Go module dependencies are needed -- `google.golang.org/api/calendar/v3` is already available in the existing `google.golang.org/api v0.265.0` module.

### Command Structure

```
gsuite calendar list                     # List upcoming events
gsuite calendar get <event-id>           # Get event details
gsuite calendar create                   # Create an event
gsuite calendar update <event-id>        # Update an event
gsuite calendar delete <event-id>        # Delete an event
gsuite calendar respond <event-id>       # RSVP to an event
gsuite calendar today                    # Shortcut: today's events
gsuite calendar week                     # Shortcut: this week's events
gsuite calendar calendars                # List available calendars
```

## Technical Approach

### Architecture

The implementation follows the established patterns exactly:
- Auth: `NewCalendarService()` parallel to `NewGmailService()` in `internal/auth/`
- Commands: New files in `cmd/` using `RunE` + `init()` registration pattern
- Output: Text (default) and JSON via `--format` flag with inline struct types

### Key Design Decisions

1. **OAuth2 Scope**: Add `calendar.CalendarEventsScope` + `calendar.CalendarReadonlyScope` to `NewOAuth2Config()` scopes list. This covers all planned operations (CRUD events + list calendars) without granting calendar admin access (ACL, settings). Calendar commands proactively check for missing scope before attempting API calls.

    > **Research Insight (Security):** `calendar.CalendarScope` grants full admin access including ACL modification and calendar deletion. A compromised token with full scope allows an attacker to modify calendar sharing settings and delete entire calendars. The narrower `CalendarEventsScope` + `CalendarReadonlyScope` combination covers all planned commands.

2. **Shared Auth Logic**: Extract unexported `newAuthenticatedClient(ctx, account) (*http.Client, error)` from `NewGmailService()`. Both `NewGmailService()` and `NewCalendarService()` become thin wrappers that call this shared function then construct their service-specific client.

    > **Research Insight (Architecture):** The extracted function must return `*http.Client` (not token source), because that is what both `gmail.NewService()` and `calendar.NewService()` accept via `option.WithHTTPClient()`. Keep it unexported -- command code should not need to know about HTTP clients.

3. **`--calendar-id`**: Regular flag added to each subcommand individually (default: `"primary"`). This matches the existing pattern where no parent command has persistent flags -- only `rootCmd` has persistent flags.

    > **Research Insight (Pattern):** The existing codebase never uses persistent flags on parent commands (`labelsCmd`, `messagesCmd`, `draftsCmd` have zero persistent flags). Using a persistent flag on `calendarCmd` would be unprecedented. Adding the flag to each subcommand individually follows the `--max-results` pattern.

4. **Date/Time Parsing**: Build a custom parser (no external library). Accept flexible input and normalize to RFC3339. All parsing functions accept `now time.Time` and `loc *time.Location` parameters for testability. Use `time.ParseInLocation` for all formats except RFC3339 (which carries its own offset). Default timezone is system local, overridable with `--timezone`.

    > **Research Insight (Date/Time):** `time.Parse` for bare datetimes returns UTC, which is almost never what a CLI user intends. `time.ParseInLocation` is required for everything except RFC3339. Performance cost of the format chain (~750ns worst case) is 5 orders of magnitude smaller than a single API call. Libraries like `araddon/dateparse` support 100+ formats we don't need -- custom parser is simpler and gives better error messages.

5. **Recurring Events**: Use `SingleEvents(true)` by default for list/today/week. Recurring event update/delete defaults to "this instance only" (pass instance ID to API). Support `--recurring-scope this|all` flag. Defer `--recurring-scope following` (requires complex 4-step process: truncate parent RRULE, create new series).

    > **Research Insight (API):** Instance IDs follow the format `{baseEventId}_{originalStartTime}` (e.g., `abc123_20260315T160000Z`). "This and following" has no single API call -- it requires modifying the parent's RRULE to add an UNTIL clause, then creating a new recurring event from the split point.

6. **Attendee Management**: Use `--add-attendees` and `--remove-attendees` on update (not `--attendees` which could replace the entire list). Validate email addresses with `net/mail.ParseAddress()` before API calls.

    > **Research Insight (Security):** Invalid email formats could silently create events with malformed attendee entries. Combined with `--send-updates all`, this could attempt to send emails to nonsensical addresses. Validate with Go's `net/mail.ParseAddress()` before making API calls.

7. **Pagination**: Auto-paginate up to `--max-results` count using `.Pages()` with early termination counter. Set `MaxResults` on the API call to `min(maxResults, 250)` to minimize over-fetching.

    > **Research Insight (Performance):** Without early termination in the `.Pages()` callback, requesting `--max-results 25` on a calendar with 2000 events could fetch all events across multiple pages. Set the API's `MaxResults` to match `--max-results` for single-page responses. The `.Pages()` method in `calendar-gen.go:5760-5776` loops until `NextPageToken` is empty -- return a sentinel error to break early.

8. **`--send-updates` default**: Default to `"none"` for safety. Users who want to send notifications must opt in with `--send-updates all`.

    > **Research Insight (Security):** For a CLI tool designed for automation, defaulting to `"all"` means a script bug or typo in attendee email could send real calendar invitations to external people. This is irreversible. The Gmail `send` command requires explicit `--to` and `--body` -- calendar notifications should follow the same intentionality.

9. **Update strategy**: Use Get+Update for `calendar update` (preserves all fields, 2 API quota units). Use Patch for `calendar respond` (single-field change, 1 API call, more atomic). Use `cmd.Flags().Changed("summary")` to determine which fields the user explicitly set.

    > **Research Insight (Performance + Security):** Get+Update introduces a TOCTOU race window but is simpler for multi-field updates. Patch avoids the race but uses 3x quota for updates. For `respond`, Patch is better: it only changes one attendee's status, requires 1 call instead of 2, and avoids sending the entire event back.

### Implementation Phases

#### Phase 1: Auth Layer Extension

**Files:**
- `internal/auth/auth.go` -- Extract `newAuthenticatedClient()`, add `NewCalendarService()`
- `internal/auth/oauth2.go` -- Add calendar scopes, add `NewCalendarService()` method on `OAuth2Config`
- `internal/auth/auth_test.go` -- Tests for scope checking

**Tasks:**
- [ ] Extract unexported `newAuthenticatedClient(ctx, account) (*http.Client, error)` from `NewGmailService()` (lines 110-147 of `auth.go`). Contains: credential loading, client cred extraction, migration check, account resolution, token loading, OAuth2 client construction.
- [ ] Refactor `NewGmailService()` to call `newAuthenticatedClient()` then `gmail.NewService(ctx, option.WithHTTPClient(client))`
- [ ] Add calendar scopes to `NewOAuth2Config()` in `oauth2.go` line 46:
  ```go
  Scopes: []string{
      gmail.GmailModifyScope,
      calendar.CalendarEventsScope,
      calendar.CalendarReadonlyScope,
  },
  ```
- [ ] Add `NewCalendarService()` method to `OAuth2Config`:
  ```go
  func (c *OAuth2Config) NewCalendarService(ctx context.Context, token *oauth2.Token) (*calendar.Service, error) {
      tokenSource := c.config.TokenSource(ctx, token)
      client := oauth2.NewClient(ctx, tokenSource)
      return calendar.NewService(ctx, option.WithHTTPClient(client))
  }
  ```
- [ ] Add top-level `NewCalendarService(ctx, account)` in `auth.go` -- thin wrapper calling `newAuthenticatedClient()` then `oauthCfg.NewCalendarService()`
- [ ] Add `isInsufficientScopeError(err error) bool` helper using `errors.As` with `*googleapi.Error`:
  ```go
  var gErr *googleapi.Error
  if errors.As(err, &gErr) && gErr.Code == 403 {
      for _, item := range gErr.Errors {
          if item.Reason == "insufficientPermissions" {
              return true
          }
      }
  }
  ```
- [ ] Add `handleCalendarError(err error, context string) error` that wraps API errors with clear messages: 401 -> "Run gsuite login", 403 -> "Run gsuite login to grant calendar access", 404 -> "not found"
- [ ] Table-driven tests for `isInsufficientScopeError` with synthetic `*googleapi.Error` values

#### Phase 2: Date/Time Parsing Utility

**Files:**
- `cmd/calendar_time.go` -- Date/time parsing and formatting helpers
- `cmd/calendar_time_test.go` -- Table-driven tests
- `cmd/calendar_time_fuzz_test.go` -- Fuzz tests

**Tasks:**
- [ ] Implement `parseDateTime(input string, loc *time.Location, now time.Time) (time.Time, error)`:
  - Try relative/keyword parsing first (cheap string comparisons)
  - Then try format chain ordered by frequency: RFC3339, `"2006-01-02"`, `"2006-01-02 15:04"`, `"2006-01-02T15:04:05"`, `"15:04"`
  - Use `time.Parse` for RFC3339 only; `time.ParseInLocation` for all others
  - For `"15:04"` format: anchor to today's date using `now`
  - Return descriptive error listing accepted formats on failure
- [ ] Implement `parseRelative(input string, loc *time.Location, now time.Time) (time.Time, bool)`:
  - Keywords: "today" (start of day), "tomorrow" (start of next day)
  - Day names: "monday"-"sunday" (next occurrence, never today)
  - Relative: "+Nd" pattern via regex `^\+(\d+)d$`
  - Case insensitive via `strings.ToLower()`
- [ ] Implement `parseDuration(input string) (time.Duration, error)` wrapping `time.ParseDuration`, rejecting zero/negative
- [ ] Implement `buildEventDateTime(t time.Time, allDay bool, tz string) *calendar.EventDateTime`:
  - All-day: use `Date` field with `"2006-01-02"` format
  - Timed: use `DateTime` field with `time.RFC3339` format, set `TimeZone` if provided
- [ ] Implement `formatEventTime(edt *calendar.EventDateTime, displayTz *time.Location) string`:
  - All-day: `"Mon Jan 02, 2006 (all day)"`
  - Timed: `"Mon Jan 02, 2006 03:04 PM MST"` converted to `displayTz`
- [ ] Implement `startOfDay(t time.Time, loc *time.Location) time.Time` and `nextWeekday(from time.Time, target time.Weekday) time.Time` helpers
- [ ] Table-driven tests (parallel, `t.Fatalf`):
  - `TestParseDateTime`: RFC3339, date-only, date+time, ISO without TZ, time-only, today/tomorrow/+Nd, day names, empty, garbage, leap year, midnight boundary, DST spring forward
  - `TestParseDuration`: valid durations, zero, negative, overflow, bare number
  - `TestBuildEventDateTime`: timed, all-day, with timezone
  - `TestFormatEventTime`: datetime, date-only, empty, unparseable
- [ ] Fuzz tests: `FuzzParseDateTime`, `FuzzParseDuration` with seed corpus from unit test values
- [ ] DST edge case tests using `time.LoadLocation("America/Los_Angeles")` and `t.Setenv("TZ", ...)`

> **Research Insight (Testing):** The codebase convention is `tt` for test table variable (not `tc`), `t.Parallel()` at both levels, `t.Fatalf` for assertions, stdlib only, fuzz seeds from unit test values. `parseRelativeDate` MUST accept a `now` parameter -- calling `time.Now()` internally makes tests non-deterministic.

#### Phase 3: Core Commands -- List, Get, Today, Week, Calendars

**Files:**
- `cmd/calendar.go` -- Parent command, list, get, today, week, calendars subcommands, single `init()`
- `cmd/calendar_test.go` -- Tests

**Tasks:**
- [ ] Define `calendarCmd` parent command (no `RunE`, no persistent flags -- help only)
- [ ] Define flag variables with `calendar` prefix to avoid collisions (matching `drafts.go`/`threads.go` pattern):
  ```go
  var (
      calendarID          string  // --calendar-id, default "primary"
      calendarMaxResults  int64   // --max-results -n, default 25
      calendarAfter       string  // --after
      calendarBefore      string  // --before
      calendarQuery       string  // --query -q
      calendarSingleEvents bool   // --single-events, default true
      calendarOrderBy     string  // --order-by, default "startTime"
      calendarTimezone    string  // --timezone
      calendarShowDeleted bool    // --show-deleted
  )
  ```
- [ ] Single `init()` registering all subcommands and all flags (even from `calendar_write.go`)
- [ ] Implement `calendar list`:
  - Flags: `--calendar-id`, `--after`, `--before`, `--max-results -n` (default 25), `--query -q`, `--single-events` (default true), `--order-by` (default "startTime"), `--timezone`, `--show-deleted`
  - Default `--after`: current time (now)
  - Default `--before`: 30 days from now
  - Fields selector (no attendees for list -- reduces payload 40-60%):
    ```go
    Fields("nextPageToken", "items(id,summary,start,end,status,location,recurrence,recurringEventId)")
    ```
  - Auto-paginate with early termination:
    ```go
    call.MaxResults(min(calendarMaxResults, 250)).Pages(ctx, func(page *calendar.Events) error {
        allEvents = append(allEvents, page.Items...)
        if int64(len(allEvents)) >= calendarMaxResults {
            return errDone
        }
        return nil
    })
    ```
  - Empty results: `outputJSON([]struct{}{})` for JSON, `"No events found."` for text
- [ ] Implement `calendar get <event-id>` (`cobra.ExactArgs(1)`):
  - Full fields (include attendees): `Fields("id,summary,start,end,status,location,description,recurrence,recurringEventId,attendees(email,responseStatus,self),htmlLink,creator,organizer")`
  - Display: summary, ID, status, calendar, time range, timezone, recurrence, location, description, attendees (email + response status), HTML link
- [ ] Implement `calendar today` (thin wrapper calling list logic with `--after` = start of today, `--before` = end of today)
- [ ] Implement `calendar week` (thin wrapper: `--after` = start of this week Monday, `--before` = end of Sunday)
- [ ] Implement `calendar calendars` (`CalendarList.List`):
  - Display: ID, summary, access role, primary indicator, timezone

**Text Output Format (list):**
```
Events (5):

  Mon Mar 15, 2026  09:00 AM - 10:00 AM  Team Standup (recurring)
  Mon Mar 15, 2026  02:00 PM - 03:00 PM  1:1 with Manager
  Tue Mar 16, 2026  All Day               Company Holiday
```

**Text Output Format (get):**
```
Event: Team Standup
ID: abc123def456
Status: confirmed
Calendar: primary

When: Mon Mar 15, 2026 09:00 AM - 10:00 AM PST
Timezone: America/Los_Angeles
Recurrence: RRULE:FREQ=WEEKLY;BYDAY=MO,WE,FR

Location: Conference Room B
Description: Daily sync with the engineering team

Attendees (3):
  alice@example.com      accepted
  bob@example.com        tentative
  charlie@example.com    needsAction

Link: https://calendar.google.com/calendar/event?eid=...
```

**JSON Output Schema (list item) -- no `omitempty`, always include all fields:**
```json
{
  "id": "string",
  "summary": "string",
  "start": "RFC3339 or date string",
  "end": "RFC3339 or date string",
  "location": "string",
  "status": "string",
  "all_day": false,
  "recurring": false,
  "recurring_event_id": "string"
}
```

> **Research Insight (Pattern):** No existing JSON struct in the codebase uses `omitempty`. Always include all fields with empty defaults. This maintains a consistent JSON contract for automation consumers. Include `recurring_event_id` so JSON consumers can correlate expanded instances.

> **Research Insight (Performance):** Remove `attendees` from the list Fields selector. For an event with 50 attendees, the attendees field alone is 2-3KB. List output doesn't display attendees, so fetching them wastes bandwidth. Keep attendees only in `calendar get`.

#### Phase 4: Write Commands -- Create, Update, Delete, Respond

**Files:**
- `cmd/calendar_write.go` -- Create, update, delete, respond commands
- `cmd/calendar_write_test.go` -- Tests

**`calendar create` flags:**
- `--summary` (required) -- Event title (no `-s` shorthand -- conflicts with `--subject` in drafts/send)
- `--start` (required) -- Start time (flexible format)
- `--end` -- End time (required unless `--duration` or `--all-day`)
- `--duration, -d` -- Alternative to `--end` (e.g., "1h", "30m")
- `--description` -- Event description
- `--location, -l` -- Event location
- `--attendees` -- Comma-separated attendee emails (validated with `net/mail.ParseAddress`)
- `--all-day` -- Create all-day event (auto-sets end = start + 1 day if no `--end`)
- `--rrule` -- Raw RRULE string (e.g., "FREQ=WEEKLY;BYDAY=MO,WE,FR"), passed through to API
- `--send-updates` -- Default `"none"`, options: `"all"`, `"externalOnly"`, `"none"`
- `--timezone` -- IANA timezone for the event
- `--calendar-id` -- Calendar ID (default: "primary")

**Tasks:**
- [ ] Implement `calendar create`:
  - Validate: `--end` and `--duration` are mutually exclusive
  - Validate: `--all-day` with time component in `--start` warns and strips time
  - Validate: `--all-day` cannot combine with `--duration`
  - For `--all-day` without `--end`: auto-set end = start + 1 day (exclusive end date per Calendar API)
  - Validate attendee emails with `net/mail.ParseAddress()`
  - Use `SendUpdates()` (not deprecated `SendNotifications()`)
  - For recurring events with `--rrule`: set `TimeZone` on `EventDateTime` (required for recurring events)
  - Output created event ID and HTML link
- [ ] Implement `calendar update <event-id>` (`cobra.ExactArgs(1)`):
  - Flags: `--summary`, `--start`, `--end`, `--description`, `--location`, `--add-attendees`, `--remove-attendees`, `--send-updates` (default "none"), `--timezone`, `--calendar-id`
  - `--recurring-scope` flag: `"this"` (default), `"all"` (deferred: `"following"`)
  - **Get+Update pattern**: Fetch existing event, mutate only fields where `cmd.Flags().Changed("flagname")` is true, send back. This avoids zeroing out unset fields.
  - When `--recurring-scope all`: use `event.RecurringEventId` (parent ID) instead of instance ID
  - For clearing a field (e.g., empty description): use `event.NullFields = append(event.NullFields, "Description")`
  - Clear `event.ServerResponse` before sending update (contains server-managed metadata)
- [ ] Implement `calendar delete <event-id>` (`cobra.ExactArgs(1)`):
  - `--send-updates` flag (default: `"none"`)
  - `--recurring-scope` flag: `"this"` (default), `"all"`
  - **Safety: When `--recurring-scope all`, require `--yes` flag**. Without it, print warning and exit non-zero:
    ```
    Warning: This will delete ALL instances of this recurring event.
    Use --yes to confirm, or --recurring-scope this to delete only this instance.
    ```
  - Print confirmation: `"Event deleted: <event-id>"`
- [ ] Implement `calendar respond <event-id>` (`cobra.ExactArgs(1)`):
  - `--status` (required): "accepted", "declined", "tentative"
  - `--comment` -- Optional RSVP comment
  - `--send-updates` (default: "none"), `--calendar-id`
  - **Use Patch (not Get+Update)**: Simpler, more atomic, 1 API call
  - Fetch event to find self (match `Self: true` in attendees), modify response status, Patch back only the attendees list
  - Handle case where user is not in attendees list: `"You are not listed as an attendee of this event"`
- [ ] Extract `validateCalendarCreateFlags(start, end, duration string, allDay bool) error` as a pure testable function
- [ ] Table-driven tests for flag validation (mutually exclusive flags, required flags, email validation)

> **Research Insight (API):** Use `errors.As(err, &gErr)` (not direct type assertion) for error checking -- the generated `Do()` methods wrap errors via `gensupport.WrapError()`. The `googleapi.Error` type implements `Unwrap()`, so `errors.As` traverses the chain correctly. Use `ForceSendFields` when you need to send a zero-value field, `NullFields` when you need to clear a field to null.

#### Phase 5: Documentation & Cleanup

**Files:**
- `cmd/root.go` -- Update root command description
- `CLAUDE.md` -- Add calendar to architecture docs and command list
- `README.md` -- Add calendar commands to documentation
- `skills/gsuite-manager/SKILL.md` -- Add calendar workflows and safety rules
- `.planning/PROJECT.md` -- Move calendar from "Out of Scope" to "Active"

**Tasks:**
- [ ] Update root command Short/Long descriptions: "Gmail CLI tool" -> "Google Workspace CLI tool"
- [ ] Update `CLAUDE.md`:
  - Architecture section: add calendar to command list
  - Auth section: mention both Gmail and Calendar scopes, `NewCalendarService()`
  - Command pattern: note Calendar follows same pattern
- [ ] Update `README.md` with calendar command documentation and examples
- [ ] Update gsuite-manager skill:
  - Add calendar operations to command reference
  - Add safety rules: `calendar delete` is destructive, `calendar create --send-updates all` sends emails
  - Add common workflows: check today's events, create meeting, RSVP
- [ ] Update `.planning/PROJECT.md`: move Calendar from "Out of Scope" to active feature

## Acceptance Criteria

### Functional Requirements
- [ ] `gsuite calendar list` lists upcoming events with date range filters
- [ ] `gsuite calendar get <id>` shows full event details including attendees
- [ ] `gsuite calendar create` creates timed, all-day, and recurring events
- [ ] `gsuite calendar update <id>` modifies event fields (only changed fields)
- [ ] `gsuite calendar delete <id>` deletes events (with `--yes` safety for recurring-all)
- [ ] `gsuite calendar respond <id>` updates RSVP status via Patch
- [ ] `gsuite calendar today` shows today's events
- [ ] `gsuite calendar week` shows this week's events (Monday-Sunday)
- [ ] `gsuite calendar calendars` lists available calendars
- [ ] All commands support `--format json` and `--format text`
- [ ] All commands support `--calendar-id` flag (default "primary")
- [ ] All commands support `--account` flag for multi-account
- [ ] Flexible date/time input parsing works correctly (RFC3339, date-only, date+time, relative)
- [ ] Existing Gmail commands continue to work unchanged
- [ ] `--send-updates` defaults to `"none"` (opt-in notification sending)

### Auth Requirements
- [ ] `gsuite login` requests Gmail + CalendarEvents + CalendarReadonly scopes
- [ ] Existing tokens without calendar scope produce proactive error: "Calendar permission not granted. Run 'gsuite login' to re-authenticate with calendar access."
- [ ] `NewCalendarService()` uses shared `newAuthenticatedClient()` -- no code duplication
- [ ] Error detection uses `errors.As` with `*googleapi.Error`, not string matching

### Quality Gates
- [ ] Table-driven unit tests for all date/time parsing functions (with `t.Parallel()`)
- [ ] Fuzz tests for `parseDateTime` and `parseDuration`
- [ ] DST edge case tests using `time.LoadLocation("America/Los_Angeles")`
- [ ] Unit tests for flag validation (mutually exclusive, required, email format)
- [ ] Unit tests for `isInsufficientScopeError` with synthetic errors
- [ ] `go build ./...` succeeds
- [ ] `go test ./... -race` passes
- [ ] No large files (target ~400 lines per file, split if exceeding 500)

### Security Requirements
- [ ] OAuth2 scopes use `CalendarEventsScope` + `CalendarReadonlyScope` (not full `CalendarScope`)
- [ ] `--send-updates` defaults to `"none"`
- [ ] `calendar delete --recurring-scope all` requires `--yes` flag
- [ ] Attendee email addresses validated with `net/mail.ParseAddress()` before API calls
- [ ] `--calendar-id` flag validated (non-empty)

## Dependencies & Prerequisites

- Google Cloud project must have Calendar API enabled
- OAuth2 client must have Calendar scope authorized in Google Cloud Console
- Existing users must re-authenticate after upgrade (`gsuite login`)
- No new Go module dependencies (`google.golang.org/api/calendar/v3` is part of existing `google.golang.org/api v0.265.0`)

## Risk Analysis & Mitigation

| Risk | Impact | Mitigation |
|------|--------|-----------|
| Existing users get 403 on calendar commands | High | Proactive scope check in `NewCalendarService()` with clear error message before any API call |
| Recurring event delete affects entire series | High | Default `--recurring-scope this`, require `--yes` for `--recurring-scope all` |
| Unintended email sends to attendees | High | Default `--send-updates "none"` -- users must opt in |
| Attendee replacement on update | High | Use `--add-attendees`/`--remove-attendees` instead of `--attendees` |
| Get+Update race condition on update | Medium | Document the TOCTOU window; use Patch for `respond` where atomicity matters |
| Timezone confusion | Medium | Default to system local, always display timezone in output, require `TimeZone` on recurring events |
| Auth code duplication | Medium | Extract shared `newAuthenticatedClient()` helper |
| Token path traversal (pre-existing) | Low | Add path sanitization to `TokenPathFor()` -- reject `/`, `\`, `..` in email |

## Future Considerations (Explicitly Deferred)

- `--recurring-scope following` (requires complex 4-step API process: truncate parent RRULE + create new series)
- Google Meet link auto-generation (`--meet` flag via `conferenceData`)
- Event reminders (`--reminder` flag)
- Free/busy queries (`gsuite calendar freebusy`)
- Event attachments
- Event colors
- Convenience recurrence flags (`--repeat weekly --until ...`)
- iCalendar (.ics) import/export
- Calendar watch/sync with sync tokens
- QuickAdd command (`gsuite calendar quick "Lunch tomorrow at noon"`)
- Incremental scope strategy (per-service scope request instead of all-at-once)
- Backporting auto-pagination to existing Gmail commands for consistency

## References & Research

### Internal References
- Auth pattern: `internal/auth/auth.go:110-147` (NewGmailService)
- OAuth2 scopes: `internal/auth/oauth2.go:46` (current scope list)
- Command pattern: `cmd/labels.go` (cleanest CRUD example, 387 lines, 4 subcommands)
- Output formatting: `cmd/root.go:66-73` (outputJSON helper)
- Flag pattern: `cmd/messages.go:133-155` (init registration)
- Flag naming: `cmd/drafts.go` uses prefixed variables (`draftsMaxResults`)
- Modify pattern: `cmd/messages.go:145-146` (`--add-labels`/`--remove-labels`)
- Testing pattern: `cmd/fuzz_test.go` (existing fuzz tests with seed corpus)
- Testing conventions: `docs/test-best-practices.md`
- Roadmap: `.planning/MILESTONES.md:76,102,127` (calendar mentioned as future)

### External References
- [Google Calendar API v3 Reference](https://developers.google.com/workspace/calendar/api/v3/reference)
- [Google Calendar API Guides](https://developers.google.com/workspace/calendar/api/guides)
- [Go Calendar v3 Package](https://pkg.go.dev/google.golang.org/api/calendar/v3)
- [Google Calendar API Scopes](https://developers.google.com/workspace/calendar/api/auth)
- [Google Calendar Recurring Events](https://developers.google.com/workspace/calendar/api/guides/recurringevents)
- [Google Calendar API Error Handling](https://developers.google.com/workspace/calendar/api/guides/errors)
- [Google Calendar API Performance Tips](https://developers.google.com/workspace/calendar/api/guides/performance)
- [Go time.ParseInLocation (DST handling)](https://github.com/golang/go/issues/63345)
- [gcalcli (Python CLI reference)](https://github.com/insanum/gcalcli)

### API Source Code References
- `calendar-gen.go:5760-5776` -- `.Pages()` implementation (pagination loop)
- `calendar-gen.go:1356-1366` -- `Event.ForceSendFields` / `NullFields`
- `calendar-gen.go:1634-1636` -- `EventAttendee.Self` field
- `calendar-gen.go:6296` -- `EventsUpdateCall.SendUpdates()` method
- `googleapi/googleapi.go:66-92` -- `googleapi.Error` struct with `Code`, `Errors[].Reason`
