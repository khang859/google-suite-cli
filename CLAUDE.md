# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Run

```bash
go build -o gsuite .              # Build binary
./gsuite --help                   # Run locally
go test ./... -race               # Run all tests
```

Version is injected at build time via ldflags targeting `cmd.Version`. GoReleaser handles this for releases; for local dev builds, version will show "dev".

## Architecture

This is a Go CLI tool using Cobra for Google Workspace operations (Gmail and Calendar) via OAuth2 PKCE authentication.

**Entry point:** `main.go` → `cmd.Execute()`

**Three packages:**
- `main` — just calls `cmd.Execute()`
- `cmd` — all Cobra commands and CLI logic
- `internal/auth` — OAuth2 PKCE authentication, Gmail and Calendar service creation

### Command Structure

Root command defines persistent flags (`--account`, `--format`, `--verbose`) and exposes getter functions (`GetAccountEmail()`, `GetOutputFormat()`, `GetVerbose()`) used by all subcommands.

Subcommands: `messages`, `threads`, `labels`, `drafts`, `send`, `search`, `calendar`, `whoami`, `version`, `login`, `logout`, `accounts`. Parent commands like `messages` and `calendar` have nested subcommands.

Calendar commands: `list`, `get`, `create`, `update`, `delete`, `respond`, `today`, `week`, `calendars`.

### Command Pattern

Every command function follows this flow:
1. Get global flags via getter functions from `root.go`
2. Call `auth.NewGmailService()` or `auth.NewCalendarService()` to get authenticated client
3. Execute API calls
4. Output as text (default) or JSON based on `--format` flag

### Authentication Flow

`internal/auth` handles credential loading with this priority:
1. `GOOGLE_CREDENTIALS` env var (raw JSON)
2. `GOOGLE_APPLICATION_CREDENTIALS` env var (file path)

Uses OAuth2 PKCE flow with scopes: `gmail.GmailModifyScope`, `calendar.CalendarEventsScope`, `calendar.CalendarReadonlyScope`. Shared auth logic is in `newAuthenticatedClient()` which both `NewGmailService()` and `NewCalendarService()` use.

### Output Formatting

All commands support `--format text` (default) and `--format json`. JSON output uses inline struct types with `json` tags defined within each command's run function (never `omitempty`). The shared `outputJSON()` helper in `root.go` handles marshaling.

### Calendar Date/Time Parsing

`cmd/calendar_time.go` provides flexible date/time parsing: RFC3339, date-only, date+time, time-only, relative (today/tomorrow/+Nd/weekday names). All functions accept `now time.Time` and `loc *time.Location` for testability — never call `time.Now()` directly.

### Email Encoding

Messages use RFC 2822 format with base64url encoding per Gmail API spec. MIME multipart is used for attachments. Key helpers: `buildRFC2822Message()`, `buildMultipartMessage()`, `extractBody()`, `decodeBase64URL()`.

## Release Process

Pushing a version tag triggers `.github/workflows/release.yml` which runs GoReleaser to build binaries for linux/darwin/windows on amd64/arm64 and creates a GitHub release.

```bash
git tag v1.x.x && git push origin v1.x.x
```

## Important rules to follows

- Code comments should explain WHY, its pointless to have comments explaining code, code should be self documented.
- Avoid large files/functions
