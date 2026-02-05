# gsuite - Gmail CLI Tool

A command-line interface for Gmail mailbox management using service account authentication with domain-wide delegation.

## What This Tool Does

gsuite provides full access to Gmail operations including reading, sending, searching, and managing messages, threads, labels, and drafts. It's designed for automation workflows and scripting with support for both human-readable and JSON output formats.

## Prerequisites

1. **Google Cloud Project** with Gmail API enabled
2. **Service Account** with domain-wide delegation enabled
3. **Google Workspace Admin** must grant the service account access to Gmail scopes:
   - `https://www.googleapis.com/auth/gmail.modify`

### Setting Up Domain-Wide Delegation

1. Create a service account in Google Cloud Console
2. Enable domain-wide delegation for the service account
3. Download the JSON key file
4. In Google Workspace Admin Console, go to Security > API Controls > Domain-wide Delegation
5. Add the service account client ID with the required Gmail scopes

## Installation

### Quick Install

```bash
curl -sSL https://raw.githubusercontent.com/khang859/google-suite-cli/main/install.sh | sh
```

This detects your OS and architecture, downloads the latest release, and installs `gsuite` to `/usr/local/bin`.

### Download from Releases

Pre-built binaries for Linux, macOS, and Windows (amd64/arm64) are available on the [Releases](https://github.com/khang859/google-suite-cli/releases) page.

### Build from Source

```bash
go build -o gsuite .
```

## Quick Start

### Verify Authentication

Test that your credentials are working correctly:

```bash
# Using credentials file flag
./gsuite --credentials-file /path/to/service-account.json --user user@yourdomain.com whoami

# Or using environment variable
export GOOGLE_APPLICATION_CREDENTIALS=/path/to/service-account.json
./gsuite --user user@yourdomain.com whoami

# Or with JSON content in environment
export GOOGLE_CREDENTIALS='{"type":"service_account",...}'
./gsuite --user user@yourdomain.com whoami
```

Expected output:
```
Email: user@yourdomain.com
Messages Total: 12345
Threads Total: 6789
```

## Available Commands

| Command | Description |
|---------|-------------|
| `whoami` | Show authenticated user's Gmail profile |
| `version` | Show version information |

## Global Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--credentials-file` | `-c` | Path to service account JSON credentials |
| `--user` | `-u` | Email of user to impersonate (required) |
| `--verbose` | `-v` | Enable verbose output |
| `--help` | `-h` | Show help |

## Credential Loading Priority

1. `--credentials-file` flag (if provided)
2. `GOOGLE_CREDENTIALS` environment variable (JSON content)
3. `GOOGLE_APPLICATION_CREDENTIALS` environment variable (file path)

## Examples

```bash
# Show help
./gsuite --help

# Show whoami help
./gsuite whoami --help

# Check authentication (verbose)
./gsuite -v --credentials-file creds.json --user admin@company.com whoami
```

## Error Messages

| Error | Solution |
|-------|----------|
| `--user flag required` | Add `--user email@domain.com` to specify user to impersonate |
| `no credentials provided` | Set `--credentials-file` flag or `GOOGLE_CREDENTIALS` env var |
| `authentication failed` | Check credentials file exists and is valid JSON |
| `Gmail API error` | Check domain-wide delegation is configured correctly |

## License

See LICENSE file for details.
