# Google Suite CLI

## What This Is

A Go CLI tool for full Gmail mailbox management via OAuth2 PKCE authentication with multi-account support. Provides complete API coverage — messages, threads, search, labels, drafts, attachments, and send — with both human-readable and JSON output modes. Supports multiple Gmail accounts with per-command account targeting.

## Core Value

Complete Gmail API coverage through a secure, scriptable command-line interface — every operation available, simple OAuth2 auth, multiple accounts.

## Requirements

### Validated

- ✓ Full Gmail API operations (read, send, delete, search, labels, threads, drafts, attachments) — v1.0
- ✓ Service account authentication with domain-wide delegation — v1.0
- ✓ Flexible credential handling (env var OR file path, user's choice) — v1.0
- ✓ Intuitive CLI UX with consistent command structure — v1.0
- ✓ Good output formatting (human-readable and JSON for scripting) — v1.0
- ✓ Clear error messages for auth and API failures — v1.0
- ✓ OAuth2 browser-based login for personal Gmail (PKCE flow) — v1.1
- ✓ Auto-detecting credential type dispatcher (service account vs OAuth2) — v1.1
- ✓ Token persistence with secure XDG-compatible storage — v1.1
- ✓ Login/logout commands for OAuth2 flow management — v1.1
- ✓ Device authorization flow for headless OAuth2 login (--no-browser) — v1.2
- ✓ Simplified auth to OAuth2 PKCE-only (removed service account, device flow) — v2.0
- ✓ Config-free auth API: `auth.NewGmailService(ctx)` with no struct — v2.0
- ✓ Clean CLI with only `--verbose` and `--format` global flags — v2.0
- ✓ Multi-account support with per-account token storage — v3.0
- ✓ Account management commands (list, switch, remove) — v3.0
- ✓ `--account` global flag for per-command account targeting — v3.0
- ✓ Transparent auto-migration from single-token to multi-account — v3.0

### Active

(None — all v1.0–v3.0 requirements shipped)

### Out of Scope

- Other Google services (Calendar, Drive, etc.) — Gmail only for now
- GUI/TUI — pure CLI, no interactive interface
- Token refresh UI — silent refresh via oauth2 library, no user interaction needed

## Context

Shipped v3.0 with 3,681 LOC across 15 Go files.
Tech stack: Go, Cobra CLI, Google Gmail API, OAuth2 PKCE.
76 commits over 6 days. 11 phases (21 plans) across 5 milestones complete.
Single auth path: OAuth2 PKCE browser flow with multi-account support.

## Constraints

- **Tech stack**: Go — single binary, good for CLI tools
- **Auth**: OAuth2 PKCE only — browser-based login, per-account tokens stored in XDG config dir
- **Security**: Credentials via env var (preferred) or file path; tokens stored with 0600 permissions

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| Go over TypeScript/Python | Single binary distribution, fast startup, good for CLI | ✓ Good |
| Support both env var and file auth | Flexibility for different deployments (agent vs local dev) | ✓ Good |
| Gmail-first scope | Focused v1, expand to other services later | ✓ Good |
| Cobra CLI framework | Industry standard, good subcommand support | ✓ Good |
| GmailModifyScope | Full read/write access for all operations | ✓ Good |
| text/plain over HTML for body display | Cleaner CLI output | ✓ Good |
| snake_case JSON keys | Standard JSON convention, consistent parsing | ✓ Good |
| Local struct types for JSON | Co-located with producing code, no leaky abstractions | ✓ Good |
| OAuth2 PKCE for personal Gmail | Secure public client auth, no client secret exposure | ✓ Good |
| Auto-detect credential type from JSON | Transparent auth — user doesn't need to specify mode | ✓ Good |
| XDG-compatible token storage | Standard path (~/.config/gsuite/), secure permissions | ✓ Good |
| auth.Login() encapsulates full flow | Clean CLI layer, single function call for entire auth sequence | ✓ Good |
| Device flow output to stderr | Keep stdout scriptable, device prompts go to stderr | ✓ Good |
| golang.org/x/oauth2 device flow | Built-in support, no custom implementation needed | ✓ Good |
| Strip to OAuth2 PKCE-only | Simplicity over flexibility — one auth path reduces complexity | ✓ Good |
| Remove Config struct entirely | No subcommand needs credentials — auth is internal to package | ✓ Good |
| Per-account token files | Simpler file locking, easier to debug, can delete one without parsing | ✓ Good |
| Email as account identifier | Emails are unique, discoverable; avoids "what did I name that profile?" | ✓ Good |
| accounts.json manifest | Single source of truth for active account; lightweight metadata | ✓ Good |
| --account flag (not --profile) | Gmail-specific, clearer intent than generic "profile" | ✓ Good |
| Auto-migration from single-token | Existing users shouldn't have to do anything; it just works | ✓ Good |

---
*Last updated: 2026-02-10 after v3.0 milestone*
