---
phase: 11-multi-account
plan: 02
subsystem: auth
tags: [oauth2, migration, multi-account, gmail-api]

requires:
  - phase: 11-01
    provides: AccountStore CRUD, per-account token storage (SaveTokenFor/LoadTokenFor)
provides:
  - MigrateIfNeeded transparently upgrades legacy token.json to multi-account
  - Login() saves per-account tokens and updates AccountStore
  - NewGmailService(ctx, account) resolves account from store or parameter
affects: [11-03, 11-04]

tech-stack:
  added: []
  patterns: [auto-migration on first call, account resolution via store]

key-files:
  created: [internal/auth/migrate.go]
  modified: [internal/auth/auth.go]

key-decisions:
  - "EnsureMigrated runs at start of every NewGmailService call for transparent migration"
  - "NewGmailService signature changed to accept account parameter (breaking change for callers)"

patterns-established:
  - "Account resolution: empty string = active account from store"
  - "Migration: check-then-migrate pattern with idempotent retry on failure"

issues-created: []

duration: 2min
completed: 2026-02-10
---

# Phase 11 Plan 02: Migration & Auth API Update Summary

**Transparent legacy token migration plus multi-account-aware Login() and NewGmailService() using AccountStore**

## Performance

- **Duration:** 2 min
- **Started:** 2026-02-10T00:47:14Z
- **Completed:** 2026-02-10T00:49:49Z
- **Tasks:** 3
- **Files modified:** 2 (1 created, 1 modified)

## Accomplishments
- Migration logic that transparently upgrades single-token to multi-account on first run
- Login() now saves tokens per-account and registers in AccountStore
- NewGmailService() resolves account from parameter or active account in store

## Task Commits

Each task was committed atomically:

1. **Task 1: Create migration logic** - `bb386ab` (feat)
2. **Task 2: Update Login() for multi-account** - `292e01a` (feat)
3. **Task 3: Update NewGmailService() for multi-account** - `393e8f4` (feat)

## Files Created/Modified
- `internal/auth/migrate.go` - MigrateIfNeeded and EnsureMigrated for legacy token migration
- `internal/auth/auth.go` - Login() saves per-account, NewGmailService() takes account parameter

## Decisions Made
- EnsureMigrated runs at start of every NewGmailService call for transparent migration
- NewGmailService signature changed to accept account parameter — callers updated in plan 11-04

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

None

## Next Phase Readiness
- Auth package compiles cleanly (`go build ./internal/auth/`)
- Full binary build expected to fail until plan 11-04 updates all callers in cmd/
- Ready for 11-03-PLAN.md (account management commands)

---
*Phase: 11-multi-account*
*Completed: 2026-02-10*
