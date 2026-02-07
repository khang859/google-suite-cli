# Project Milestones: Google Suite CLI

## v2.0 Auth Simplification (Shipped: 2026-02-07)

**Delivered:** Stripped CLI to OAuth2 PKCE-only authentication — removed service account JWT, device flow, Config struct, and legacy flags. Net -60 lines, single auth path via `auth.NewGmailService(ctx)`.

**Phases completed:** 9-10 (2 plans total)

**Key accomplishments:**
- Removed all service account JWT auth, device flow, and credential type dispatching
- Simplified auth to single-path OAuth2 PKCE (no credential type detection)
- Removed Config struct entirely — `auth.NewGmailService(ctx)` with no config
- Removed `--credentials-file` and `--user` legacy flags from CLI
- Updated all 8 subcommands to simplified auth API

**Stats:**
- 17 files changed, 513 insertions, 573 deletions (net -60 lines)
- 3,028 lines of Go (total project)
- 2 phases, 2 plans, 5 tasks
- 1 day (2026-02-07)
- 9 commits

**Git range:** `7181d49` → `d7cc85e`

**What's next:** TBD — project complete with full Gmail CLI and clean OAuth2-only auth.

---

## v1.2 Headless Login (Shipped: 2026-02-06)

**Delivered:** RFC 8628 device authorization flow for OAuth2 login on headless machines — `gsuite login --no-browser` prints a verification URL and code to stderr for cross-device authentication.

**Phases completed:** 8 (1 plan total)

**Key accomplishments:**
- DeviceAuthenticate method using golang.org/x/oauth2 built-in device flow
- `--no-browser` flag on `gsuite login` for headless environments (EC2, SSH, containers)
- Stderr output for device flow prompts (stdout stays scriptable)

**Stats:**
- 7 files changed, 265 insertions, 25 deletions
- 3,455 lines of Go (total project)
- 1 phase, 1 plan, 2 tasks
- 1 day (2026-02-05 → 2026-02-06)
- 5 commits

**Git range:** `ed42a27` (milestone start) → `cf9b6d0` (plan complete)

**What's next:** TBD — potential areas include Calendar/Drive support, batch operations, or interactive mode.

---

## v1.1 OAuth2 Support (Shipped: 2026-02-06)

**Delivered:** OAuth2 browser-based login for personal Gmail accounts — `gsuite login` triggers PKCE flow, auto-detecting credential type dispatches between service account and OAuth2 transparently.

**Phases completed:** 5-7 (3 plans total)

**Key accomplishments:**
- OAuth2 PKCE authorization flow with browser-based login and XDG-compatible token storage
- Auto-detecting credential type dispatcher (service account vs OAuth2 from JSON structure)
- `gsuite login` / `gsuite logout` commands for OAuth2 flow
- Removed 18 --user guards — all commands work with both auth methods transparently
- Secure token persistence (0600 file, 0700 directory permissions)

**Stats:**
- 18 files changed, 1,023 insertions, 130 deletions
- 3,411 lines of Go (total project)
- 3 phases, 3 plans, 7 tasks
- 1 day (2026-02-05 → 2026-02-06)
- 9 commits

**Git range:** `4d9d932` (phase 5 start) → `9ddfba0` (phase 7 end)

**What's next:** TBD — potential areas include Calendar/Drive support, batch operations, or interactive mode.

---

## v1.0 MVP (Shipped: 2026-02-05)

**Delivered:** Complete Gmail CLI with full API coverage — read, write, search, labels, threads, drafts, attachments, and JSON output.

**Phases completed:** 1-4 (11 plans total)

**Key accomplishments:**
- Go project foundation with Cobra CLI and service account auth (env var + file path)
- Full read operations: messages list/get, search, labels, threads with body decoding
- Complete write operations: send, drafts CRUD, labels CRUD, message label modification
- Attachment support: download and send-with-attachment
- JSON output format (--format json) across all commands with consistent snake_case

**Stats:**
- 11 Go files, 2,972 lines of Go
- 4 phases, 11 plans
- 2 days from init to ship (2026-02-04 → 2026-02-05)
- 41 commits

**Git range:** `d0c4972` (init) → `c8c9e2b` (phase 4 complete)

**What's next:** TBD — potential areas include Calendar/Drive support, OAuth flow, batch operations, or interactive mode.

---
