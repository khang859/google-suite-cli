# Google Suite CLI

## What This Is

A Go CLI tool for full Gmail mailbox management via service account authentication. Provides complete API coverage — messages, threads, search, labels, drafts, attachments, and send — with both human-readable and JSON output modes. Designed for personal automation workflows running on an always-on AI agent.

## Core Value

Complete Gmail API coverage through a secure, scriptable command-line interface — every operation available, flexible auth options.

## Requirements

### Validated

- ✓ Full Gmail API operations (read, send, delete, search, labels, threads, drafts, attachments) — v1.0
- ✓ Service account authentication with domain-wide delegation — v1.0
- ✓ Flexible credential handling (env var OR file path, user's choice) — v1.0
- ✓ Intuitive CLI UX with consistent command structure — v1.0
- ✓ Good output formatting (human-readable and JSON for scripting) — v1.0
- ✓ Clear error messages for auth and API failures — v1.0

### Active

(None — all v1.0 requirements shipped)

### Out of Scope

- Other Google services (Calendar, Drive, etc.) — Gmail only for v1
- OAuth user flow — service account auth only for v1
- GUI/TUI — pure CLI, no interactive interface

## Context

Shipped v1.0 with 2,972 LOC across 11 Go files.
Tech stack: Go, Cobra CLI, Google Gmail API, JWT service account auth.
41 commits over 2 days. All 4 phases (11 plans) complete.

## Constraints

- **Tech stack**: Go — single binary, good for CLI tools
- **Auth**: Service account only — no OAuth flow complexity
- **Security**: Credentials via env var (preferred) or file path — flexible for different deployment scenarios

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

---
*Last updated: 2026-02-05 after v1.0 milestone*
