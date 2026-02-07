---
phase: 09-remove-auth-code
plan: 01
subsystem: auth
tags: [oauth2, pkce, refactor, cleanup]

# Dependency graph
requires:
  - phase: 08-device-auth-flow
    provides: device auth flow code to be removed
provides:
  - OAuth2 PKCE-only auth package (no service account, no device flow)
  - Simplified Login() function (single code path)
  - Simplified NewGmailService() (direct OAuth2 token path)
affects: [10-simplify-cli]

# Tech tracking
tech-stack:
  added: []
  patterns: [single-auth-path]

key-files:
  modified: [internal/auth/auth.go, internal/auth/oauth2.go, cmd/login.go]

key-decisions:
  - "Kept UserEmail in Config struct (unused) to avoid touching all subcommands — Phase 10 handles that"

patterns-established:
  - "Single OAuth2 PKCE auth path: no credential type detection needed"

issues-created: []

# Metrics
duration: 2min
completed: 2026-02-07
---

# Phase 9 Plan 1: Strip Auth to OAuth2-Only Summary

**Removed service account JWT auth, device flow, and credential dispatching — auth is now single-path OAuth2 PKCE**

## Performance

- **Duration:** 2 min
- **Started:** 2026-02-07T02:12:16Z
- **Completed:** 2026-02-07T02:14:19Z
- **Tasks:** 3
- **Files modified:** 3

## Accomplishments

- Removed all service account code (credentialType enum, detectCredentialType, newServiceAccountGmailService, JWTConfig usage)
- Removed DeviceAuthenticate() method from oauth2.go
- Simplified Login() to single PKCE browser flow (removed noBrowser parameter)
- Simplified NewGmailService() to direct OAuth2 token path (removed credential type dispatch)
- Removed --no-browser flag from login command

## Task Commits

Each task was committed atomically:

1. **Task 1: Strip auth.go to OAuth2-only** - `529c50b` (refactor)
2. **Task 2: Strip device flow from oauth2.go** - `ea08c20` (refactor)
3. **Task 3: Clean up login command** - `2f3d45e` (refactor)

## Files Created/Modified

- `internal/auth/auth.go` — Removed 117 lines: service account code, credential type detection, type dispatching. Login() now takes (ctx, credJSON) only.
- `internal/auth/oauth2.go` — Removed DeviceAuthenticate() method and unused os import (26 lines removed)
- `cmd/login.go` — Removed --no-browser flag, noBrowser variable, device flow references from help text

## Decisions Made

- Kept `UserEmail` in Config struct as deprecated field to avoid breaking all subcommand callers — Phase 10 handles removing `--user` flag and Config.UserEmail together

## Deviations from Plan

None — plan executed exactly as written.

## Issues Encountered

None

## Next Phase Readiness

- Auth package is now pure OAuth2 PKCE — no service account or device flow code remains
- Ready for Phase 10: Simplify CLI (remove --credentials-file, --user flags, update subcommands)

---
*Phase: 09-remove-auth-code*
*Completed: 2026-02-07*
