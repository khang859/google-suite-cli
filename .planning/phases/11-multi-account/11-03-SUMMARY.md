---
phase: 11-multi-account
plan: 03
subsystem: cli
tags: [cobra, accounts, multi-account, flags]

requires:
  - phase: 11-01
    provides: AccountStore with CRUD operations
  - phase: 11-02
    provides: Login() saves per-account token, NewGmailService(ctx, account) resolves per-account tokens
provides:
  - accounts list/switch/remove CLI commands
  - --account global flag with GSUITE_ACCOUNT env var fallback
  - multi-account login/logout behavior
affects: [11-04]

tech-stack:
  added: []
  patterns: [account-aware subcommands, per-account token lifecycle]

key-files:
  created: [cmd/accounts.go]
  modified: [cmd/root.go, cmd/login.go, cmd/whoami.go, cmd/send.go, cmd/search.go, cmd/drafts.go, cmd/threads.go, cmd/labels.go, cmd/messages.go]

key-decisions:
  - "Included NewGmailService caller updates in Task 1 since 11-02 changed the signature without updating callers"

patterns-established:
  - "GetAccountEmail() getter pattern with env var fallback for --account flag"

issues-created: []

duration: 4min
completed: 2026-02-10
---

# Phase 11 Plan 3: CLI Commands & Account Flag Summary

**`accounts list|switch|remove` commands, `--account` global flag, and multi-account login/logout updates**

## Performance

- **Duration:** 4 min
- **Started:** 2026-02-10T00:51:16Z
- **Completed:** 2026-02-10T00:54:55Z
- **Tasks:** 3
- **Files modified:** 10

## Accomplishments
- Added `--account` global flag with `GSUITE_ACCOUNT` env var fallback via `GetAccountEmail()`
- Created `accounts` command group with `list`, `switch`, and `remove` subcommands
- Updated login to show multi-account hint when >1 accounts exist
- Updated logout to accept optional email argument and use AccountStore/DeleteTokenFor

## Task Commits

Each task was committed atomically:

1. **Task 1: Add --account global flag** - `4d9d695` (feat)
2. **Task 2: Create accounts command group** - `95563cb` (feat)
3. **Task 3: Update login/logout for multi-account** - `029bd03` (feat)

## Files Created/Modified
- `cmd/root.go` - Added accountEmail var, --account flag, GetAccountEmail() getter
- `cmd/accounts.go` - New file with accounts list/switch/remove subcommands
- `cmd/login.go` - Multi-account hint on login, optional email arg on logout
- `cmd/whoami.go` - Updated NewGmailService call to pass GetAccountEmail()
- `cmd/send.go` - Updated NewGmailService call to pass GetAccountEmail()
- `cmd/search.go` - Updated NewGmailService call to pass GetAccountEmail()
- `cmd/drafts.go` - Updated NewGmailService call to pass GetAccountEmail()
- `cmd/threads.go` - Updated NewGmailService call to pass GetAccountEmail()
- `cmd/labels.go` - Updated NewGmailService call to pass GetAccountEmail()
- `cmd/messages.go` - Updated NewGmailService call to pass GetAccountEmail()

## Decisions Made
- Included all NewGmailService caller updates in Task 1 since plan 11-02 changed the signature to `NewGmailService(ctx, account)` without updating existing callers — build would fail otherwise

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Updated all NewGmailService callers to pass account parameter**
- **Found during:** Task 1 (--account flag addition)
- **Issue:** Plan 11-02 changed NewGmailService signature to require account string but didn't update 19 call sites across 7 files
- **Fix:** Updated all callers to pass `GetAccountEmail()` as the account parameter
- **Files modified:** cmd/whoami.go, cmd/send.go, cmd/search.go, cmd/drafts.go, cmd/threads.go, cmd/labels.go, cmd/messages.go
- **Verification:** `go build -o gsuite .` succeeds
- **Committed in:** 4d9d695

---

**Total deviations:** 1 auto-fixed (blocking)
**Impact on plan:** Fix was necessary for compilation. No scope creep.

## Issues Encountered
None

## Next Phase Readiness
- All CLI commands for multi-account management are in place
- Ready for 11-04-PLAN.md (final integration/wiring)

---
*Phase: 11-multi-account*
*Completed: 2026-02-10*
