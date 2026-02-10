---
phase: 11-multi-account
plan: 04
subsystem: auth
tags: [go, cobra, multi-account, gmail-api]

# Dependency graph
requires:
  - phase: 11-multi-account (plan 03)
    provides: GetAccountEmail() getter and --account global flag
provides:
  - All 19 command call sites wired to multi-account auth API
  - Full end-to-end --account flag support across every command
affects: []

# Tech tracking
tech-stack:
  added: []
  patterns: ["every command passes GetAccountEmail() to NewGmailService"]

key-files:
  created: []
  modified: [cmd/messages.go, cmd/threads.go, cmd/labels.go, cmd/drafts.go, cmd/search.go, cmd/send.go, cmd/whoami.go]

key-decisions:
  - "No separate commit needed — work was completed ahead of schedule in 11-03"

patterns-established:
  - "All NewGmailService calls use (ctx, GetAccountEmail()) signature"

issues-created: []

# Metrics
duration: 1 min
completed: 2026-02-10
---

# Phase 11 Plan 4: Wire --account Flag to All Commands Summary

**All 19 NewGmailService call sites across 7 command files wired to multi-account auth via GetAccountEmail()**

## Performance

- **Duration:** 1 min
- **Started:** 2026-02-10T00:56:34Z
- **Completed:** 2026-02-10T00:57:37Z
- **Tasks:** 1 (verification only — work already complete)
- **Files modified:** 7

## Accomplishments
- Verified all 19 call sites across 7 command files already pass GetAccountEmail()
- Confirmed `go build` succeeds with no errors
- Confirmed zero remaining old-style `NewGmailService(ctx)` calls
- Full multi-account support wired end-to-end

## Task Commits

Work was completed ahead of schedule during plan 11-03:

1. **Task 1: Update all command files to pass account parameter** - `4d9d695` (feat — bundled with 11-03 --account flag commit)

No additional commit needed — all 19 call sites were already updated.

## Files Created/Modified
- `cmd/messages.go` - 4 call sites updated (runMessagesList, runMessagesGet, runMessagesModify, runGetAttachment)
- `cmd/threads.go` - 2 call sites updated (runThreadsList, runThreadsGet)
- `cmd/labels.go` - 4 call sites updated (runLabelsList, runLabelsGet, runLabelsCreate, runLabelsDelete)
- `cmd/drafts.go` - 6 call sites updated (runDraftsList, runDraftsGet, runDraftsCreate, runDraftsUpdate, runDraftsSend, runDraftsDelete)
- `cmd/search.go` - 1 call site updated (runSearch)
- `cmd/send.go` - 1 call site updated (runSend)
- `cmd/whoami.go` - 1 call site updated (runWhoami)

## Decisions Made
- No separate commit created — the mechanical replacement was bundled into the 11-03 commit (`4d9d695`) when the --account flag and GetAccountEmail() were added. This was more atomic since the flag definition and its wiring belong together.

## Deviations from Plan

### Note: Work Completed Ahead of Schedule

**[Deviation] Plan 11-03 included 11-04 work**
- **Found during:** Verification phase
- **Issue:** All 19 call sites were already updated in commit `4d9d695` (plan 11-03)
- **Impact:** No code changes needed — plan was purely verification
- **Assessment:** This is a positive deviation. The 11-03 implementation correctly bundled the flag wiring with the flag definition, which is more atomic and logical.

## Issues Encountered
None

## Next Phase Readiness
- Phase 11 complete — all 4 plans executed
- v3.0 Multi-Account Support milestone is done
- Full multi-account lifecycle: login, accounts list/switch/remove, --account flag on all commands, logout

---
*Phase: 11-multi-account*
*Completed: 2026-02-10*
