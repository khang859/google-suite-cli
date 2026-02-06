---
phase: 06-auth-dispatcher
plan: 01
subsystem: auth
tags: [oauth2, service-account, json-detection, credential-dispatch]

# Dependency graph
requires:
  - phase: 05-oauth2-core
    provides: OAuth2Config, PKCE flow, token persistence (LoadToken/SaveToken)
provides:
  - Auto-detecting credential type dispatcher in NewGmailService
  - detectCredentialType function for service_account vs OAuth2 client JSON
  - extractOAuth2ClientCreds for parsing installed/web client credentials
affects: [07-cli-integration]

# Tech tracking
tech-stack:
  added: []
  patterns: [credential-type-dispatch, dual-auth-path]

key-files:
  created: []
  modified: [internal/auth/auth.go]

key-decisions:
  - "Single file refactor — all detection and dispatch logic in auth.go alongside existing LoadCredentials"
  - "UserEmail validation moved from NewGmailService top-level to service account path only"
  - "OAuth2 path returns actionable error when no token cached instead of auto-triggering browser flow"

patterns-established:
  - "Credential dispatch: JSON key detection routes to service account or OAuth2 flow transparently"
  - "Private helpers: newServiceAccountGmailService and newOAuth2GmailService encapsulate each auth path"

issues-created: []

# Metrics
duration: 1min
completed: 2026-02-06
---

# Phase 6 Plan 1: Auth Dispatcher Summary

**Auto-detecting credential type dispatcher that routes NewGmailService to service account JWT or OAuth2 token flow based on JSON structure**

## Performance

- **Duration:** 1 min
- **Started:** 2026-02-06T00:00:12Z
- **Completed:** 2026-02-06T00:01:17Z
- **Tasks:** 2
- **Files modified:** 1

## Accomplishments
- Credential type detection: parses JSON to identify service_account, installed, or web credential formats
- OAuth2 client credential extraction from installed/web JSON keys
- NewGmailService transparently dispatches between service account and OAuth2 flows
- Service account path preserved identically in private helper function
- OAuth2 path loads cached token and returns clear "run gsuite login" error when missing

## Task Commits

Each task was committed atomically:

1. **Task 1 + Task 2: Credential detection + auth dispatcher** - `f27e8e4` (feat) — Combined since both tasks modify the same file with interdependent changes

**Plan metadata:** (pending)

## Files Created/Modified
- `internal/auth/auth.go` - Added detectCredentialType, extractOAuth2ClientCreds, refactored NewGmailService to dispatch, extracted newServiceAccountGmailService and newOAuth2GmailService helpers

## Decisions Made
- Combined both tasks into single commit since they modify the same file and Task 2 directly uses Task 1's functions
- Moved UserEmail validation into service account path only (OAuth2 doesn't need it)
- OAuth2 path returns descriptive error rather than auto-triggering login flow (that's Phase 7's responsibility)

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered
None

## Next Phase Readiness
- Auth dispatcher complete — NewGmailService transparently handles both credential types
- Phase 7 (CLI Integration) can now add `login` command and remove `--user` guards
- No blockers or concerns

---
*Phase: 06-auth-dispatcher*
*Completed: 2026-02-06*
