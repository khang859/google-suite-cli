# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Run

```bash
go build -o gsuite .              # Build binary
./gsuite --help                   # Run locally
```

Version is injected at build time via ldflags targeting `cmd.Version`. GoReleaser handles this for releases; for local dev builds, version will show "dev".

No tests exist yet. No linter is configured.

## Architecture

This is a Go CLI tool using Cobra for Gmail operations via Google service account with domain-wide delegation.

**Entry point:** `main.go` → `cmd.Execute()`

**Three packages:**
- `main` — just calls `cmd.Execute()`
- `cmd` — all Cobra commands and CLI logic
- `internal/auth` — service account authentication and Gmail service creation

### Command Structure

Root command defines persistent flags (`--credentials-file`, `--user`, `--format`, `--verbose`) and exposes getter functions (`GetCredentialsFile()`, `GetUserEmail()`, `GetOutputFormat()`, `GetVerbose()`) used by all subcommands.

Subcommands: `messages`, `threads`, `labels`, `drafts`, `send`, `search`, `whoami`, `version`. Parent commands like `messages` have nested subcommands (`list`, `get`, `modify`, `get-attachment`).

### Command Pattern

Every command function follows this flow:
1. Get global flags via getter functions from `root.go`
2. Build `auth.Config` struct
3. Call `auth.NewGmailService()` to get authenticated client
4. Execute Gmail API calls
5. Output as text (default) or JSON based on `--format` flag

### Authentication Flow

`internal/auth` handles credential loading with this priority:
1. `--credentials-file` flag
2. `GOOGLE_CREDENTIALS` env var (raw JSON)
3. `GOOGLE_APPLICATION_CREDENTIALS` env var (file path)

Uses JWT config with `gmail.GmailModifyScope`, sets `Subject` to the `--user` email for domain-wide delegation. All API calls use "me" which resolves to the impersonated user.

### Output Formatting

All commands support `--format text` (default) and `--format json`. JSON output uses inline struct types with `json` tags defined within each command's run function. The shared `outputJSON()` helper in `root.go` handles marshaling.

### Email Encoding

Messages use RFC 2822 format with base64url encoding per Gmail API spec. MIME multipart is used for attachments. Key helpers: `buildRFC2822Message()`, `buildMultipartMessage()`, `extractBody()`, `decodeBase64URL()`.

## Release Process

Pushing a version tag triggers `.github/workflows/release.yml` which runs GoReleaser to build binaries for linux/darwin/windows on amd64/arm64 and creates a GitHub release.

```bash
git tag v1.x.x && git push origin v1.x.x
```

## Important rules to follows

- Code comments should explain WHY, its pointless to have comments explaining code, code should be self documented.
- Avoid large files/functions
