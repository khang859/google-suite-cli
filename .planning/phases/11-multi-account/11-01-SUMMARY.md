---
phase: 11-multi-account
plan: 01
subsystem: auth
tags: [oauth2, multi-account, token-storage, accounts-json]

# Dependency graph
requires: []
provides:
  - AccountStore type with full CRUD for managing multiple Gmail accounts
  - Per-account token storage under ~/.config/gsuite/tokens/
  - Legacy token functions preserved for migration
affects: [11-02-migration, 11-03-commands, 11-04-integration]

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Per-account token files at tokens/<email>.json"
    - "AccountStore manifest at accounts.json with active field"
    - "Case-insensitive email comparisons via strings.EqualFold"

key-files:
  created: [internal/auth/accounts.go]
  modified: [internal/auth/token.go, internal/auth/auth.go, cmd/login.go]

key-decisions:
  - "Legacy functions renamed (not deleted) to preserve backward compatibility until migration plan"
  - "Callers updated to use legacy names so existing behavior is unchanged"

patterns-established:
  - "AccountStore CRUD pattern: Load → mutate → Save"
  - "Per-account token path: TokenPathFor(email) → tokens/<email>.json"

issues-created: []

# Metrics
duration: 1min
completed: 2026-02-10
---

# Phase 11 Plan 1: Account Store & Per-Account Token Storage Summary

**AccountStore type with full CRUD and per-account token storage under ~/.config/gsuite/tokens/, with legacy functions preserved for migration**

## Performance

- **Duration:** 1 min
- **Started:** 2026-02-10T00:43:56Z
- **Completed:** 2026-02-10T00:45:28Z
- **Tasks:** 2
- **Files modified:** 4

## Accomplishments
- AccountStore type with Add, Remove, SetActive, GetActive, HasAccount, List operations
- Per-account token storage with SaveTokenFor/LoadTokenFor/DeleteTokenFor
- Legacy TokenPath/SaveToken/LoadToken renamed and preserved for migration
- All existing commands still compile and behave identically

## Task Commits

Each task was committed atomically:

1. **Task 1: Create AccountStore type with CRUD operations** - `c87b4e1` (feat)
2. **Task 2: Refactor token storage for per-account tokens** - `eb6c385` (feat)

## Files Created/Modified
- `internal/auth/accounts.go` - AccountStore type, CRUD operations, accounts.json I/O
- `internal/auth/token.go` - Legacy renames + new per-account TokenPathFor/SaveTokenFor/LoadTokenFor/DeleteTokenFor
- `internal/auth/auth.go` - Updated to use saveLegacyToken/LoadLegacyToken
- `cmd/login.go` - Updated to use LegacyTokenPath

## Decisions Made
- Legacy functions renamed (not deleted) to preserve backward compat until migration plan (11-02) handles the switchover
- Callers updated to legacy names immediately so the build never breaks

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Updated callers of renamed token functions**
- **Found during:** Task 2 (token refactor)
- **Issue:** Renaming TokenPath→LegacyTokenPath, SaveToken→saveLegacyToken, LoadToken→LoadLegacyToken broke auth.go and login.go
- **Fix:** Updated all three callers to use the new legacy function names
- **Files modified:** internal/auth/auth.go, cmd/login.go
- **Verification:** `go build` succeeds without errors
- **Committed in:** eb6c385 (part of Task 2 commit)

---

**Total deviations:** 1 auto-fixed (1 blocking)
**Impact on plan:** Essential fix — callers must use renamed functions. No scope creep.

## Issues Encountered
None

## Next Phase Readiness
- Account storage foundation complete, ready for migration logic (11-02)
- Legacy functions preserved as the bridge — 11-02 will wire up the migration path
- No blockers

---
*Phase: 11-multi-account*
*Completed: 2026-02-10*
