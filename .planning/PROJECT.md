# Google Suite CLI

## What This Is

A Go CLI tool for full Gmail mailbox management via service account authentication. Designed for personal automation workflows running on an always-on AI agent, with credentials handled securely through environment variables.

## Core Value

Complete Gmail API coverage through a secure, scriptable command-line interface — every operation available, credentials never touch disk.

## Requirements

### Validated

(None yet — ship to validate)

### Active

- [ ] Full Gmail API operations (read, send, delete, search, labels, threads, drafts, attachments)
- [ ] Service account authentication with domain-wide delegation
- [ ] Environment variable based credential handling (creds never stored on disk)
- [ ] Intuitive CLI UX with consistent command structure
- [ ] Good output formatting (human-readable and JSON for scripting)
- [ ] Clear error messages for auth and API failures

### Out of Scope

- Other Google services (Calendar, Drive, etc.) — Gmail only for v1
- OAuth user flow — service account auth only for v1
- GUI/TUI — pure CLI, no interactive interface
- Encrypted config files — using env vars instead

## Context

- Runs on an always-on AI agent where security is critical
- Personal automation use case (scripts and workflows)
- Service account will impersonate user via domain-wide delegation
- Go chosen for single binary distribution and fast startup

## Constraints

- **Tech stack**: Go — single binary, good for CLI tools
- **Auth**: Service account only — no OAuth flow complexity
- **Security**: Credentials via environment variables only — never written to disk

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| Go over TypeScript/Python | Single binary distribution, fast startup, good for CLI | — Pending |
| Env vars over encrypted config | Simpler, agent environment already secured, no key management | — Pending |
| Gmail-first scope | Focused v1, expand to other services later | — Pending |

---
*Last updated: 2026-02-04 after initialization*
