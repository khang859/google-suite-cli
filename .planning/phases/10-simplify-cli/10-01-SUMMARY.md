---
phase: 10-simplify-cli
plan: 01
subsystem: auth
tags: [go, cobra, oauth2, cli-flags, refactor]

# Dependency graph
requires:
  - phase: 09-remove-auth-code
    provides: OAuth2 PKCE-only auth, no service account or device flow code
provides:
  - Config-free auth API (NewGmailService takes only ctx)
  - Clean CLI with only --verbose and --format global flags
  - Clear "run gsuite login" error messages
affects: []

# Tech tracking
tech-stack:
  added: []
  patterns: [simplified auth API without Config struct]

key-files:
  created: []
  modified: [internal/auth/auth.go, cmd/root.go, cmd/login.go, cmd/whoami.go, cmd/messages.go, cmd/threads.go, cmd/labels.go, cmd/drafts.go, cmd/search.go, cmd/send.go]

key-decisions:
  - "Removed Config struct entirely rather than simplifying it — no subcommand needs to pass credentials"

patterns-established:
  - "Auth pattern: auth.NewGmailService(ctx) with no config — all credential resolution internal to auth package"

issues-created: []

# Metrics
duration: 4min
completed: 2026-02-07
---

# Phase 10 Plan 01: Simplify CLI Summary

**Removed Config struct and legacy flags, all subcommands now use auth.NewGmailService(ctx) directly — net -349 lines**

## Performance

- **Duration:** 4 min
- **Started:** 2026-02-07T02:48:38Z
- **Completed:** 2026-02-07T02:52:45Z
- **Tasks:** 2
- **Files modified:** 10

## Accomplishments
- Removed auth.Config struct entirely from internal/auth package
- Removed --credentials-file and --user persistent flags from root command
- Updated all 8 subcommand files to use simplified auth.NewGmailService(ctx) API
- Updated help text across login, whoami, and root commands to reflect OAuth2-only flow

## Task Commits

Each task was committed atomically:

1. **Task 1: Simplify auth package and root command flags** - `87e4b27` (refactor)
2. **Task 2: Update all subcommands to simplified auth API** - `53eb7a5` (refactor)

**Plan metadata:** (next commit) (docs: complete plan)

## Files Created/Modified
- `internal/auth/auth.go` - Removed Config struct, simplified LoadCredentials() and NewGmailService() signatures
- `cmd/root.go` - Removed --credentials-file and --user flags, removed getter functions, updated description
- `cmd/login.go` - Uses auth.LoadCredentials() directly, updated help text
- `cmd/whoami.go` - Uses auth.NewGmailService(ctx), removed domain-wide delegation references
- `cmd/messages.go` - Updated 4 run functions to simplified auth API
- `cmd/threads.go` - Updated 2 run functions to simplified auth API
- `cmd/labels.go` - Updated 4 run functions to simplified auth API
- `cmd/drafts.go` - Updated 6 run functions to simplified auth API
- `cmd/search.go` - Updated runSearch to simplified auth API
- `cmd/send.go` - Updated runSend to simplified auth API

## Decisions Made
- Removed Config struct entirely rather than simplifying — no subcommand needs to pass credentials since auth is always OAuth2 PKCE via env vars

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered
None

## Next Phase Readiness
- Phase 10 complete — this was the last plan in the last phase
- v2.0 Auth Simplification milestone is complete
- CLI now has clean OAuth2-only auth with simplified API

---
*Phase: 10-simplify-cli*
*Completed: 2026-02-07*
