# gsuite - Google Workspace CLI Tool

A command-line interface for Google Workspace management. Authenticate with your account via OAuth2 and manage Gmail messages, threads, labels, drafts, and Google Calendar events from the terminal.

## Installation

### Quick Install

```bash
curl -sSL https://raw.githubusercontent.com/khang859/google-suite-cli/main/install.sh | sh
```

### Download from Releases

Pre-built binaries for Linux, macOS, and Windows (amd64/arm64) are available on the [Releases](https://github.com/khang859/google-suite-cli/releases) page.

### Build from Source

```bash
go build -o gsuite .
```

## Prerequisites

1. **Google Cloud Project** with Gmail API and Calendar API enabled
2. **OAuth2 Client Credentials** (Desktop or Web application type)
3. Set credentials via environment variable:
   - `GOOGLE_CREDENTIALS` — raw JSON content
   - `GOOGLE_APPLICATION_CREDENTIALS` — path to JSON file

## Quick Start

```bash
# Login (opens browser for OAuth2 consent)
gsuite login

# Verify authentication
gsuite whoami

# List recent messages
gsuite messages list

# Send an email
gsuite send --to "user@example.com" --subject "Hello" --body "Message content"

# Search messages
gsuite search "from:user@example.com is:unread"

# View today's calendar events
gsuite calendar today

# Create a meeting
gsuite calendar create --summary "Team Standup" --start "2026-03-15 09:00" --duration 30m
```

## Multi-Account Support

Login with multiple Gmail accounts and switch between them.

```bash
# Login with first account
gsuite login

# Login with another account
gsuite login

# List all accounts (* marks active)
gsuite accounts list

# Switch active account
gsuite accounts switch other@gmail.com

# Run a command as a specific account
gsuite --account other@gmail.com messages list

# Remove an account
gsuite accounts remove old@gmail.com

# Logout active account
gsuite logout
```

The `--account` flag (or `GSUITE_ACCOUNT` env var) can be passed to any command to override the active account for that invocation.

## Available Commands

| Command | Description |
|---------|-------------|
| `login` | Authenticate with Gmail via OAuth2 (opens browser) |
| `logout [email]` | Remove saved token (active account or specified email) |
| `accounts list` | List all authenticated accounts |
| `accounts switch <email>` | Switch the active account |
| `accounts remove <email>` | Remove an authenticated account |
| `whoami` | Show authenticated user's Gmail profile |
| `messages list` | List messages with optional filters |
| `messages get <id>` | Get a specific message |
| `messages modify <id>` | Add/remove labels on a message |
| `messages get-attachment <msg-id> <att-id>` | Download an attachment |
| `threads list` | List conversation threads |
| `threads get <id>` | Get a thread with all messages |
| `labels list` | List all Gmail labels |
| `labels create` | Create a new label |
| `labels update <id>` | Update a label |
| `labels delete <id>` | Delete a label |
| `drafts list` | List drafts |
| `drafts get <id>` | Get a specific draft |
| `drafts create` | Create a new draft |
| `drafts update <id>` | Update an existing draft |
| `drafts send <id>` | Send a draft |
| `drafts delete <id>` | Delete a draft |
| `send` | Send an email (supports markdown, attachments) |
| `search <query>` | Search messages using Gmail query syntax |
| `calendar list` | List upcoming calendar events |
| `calendar get <id>` | Get event details including attendees |
| `calendar create` | Create a calendar event |
| `calendar update <id>` | Update an existing event |
| `calendar delete <id>` | Delete a calendar event |
| `calendar respond <id>` | RSVP to an event invitation |
| `calendar today` | Show today's events |
| `calendar week` | Show this week's events (Mon-Sun) |
| `calendar calendars` | List available calendars |
| `version` | Show version information |
| `install-skill` | Install the Claude Code skill for Gmail management |

## Global Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--account` | | Use a specific account email |
| `--format` | `-f` | Output format: `text` (default) or `json` |
| `--verbose` | `-v` | Enable verbose output |
| `--help` | `-h` | Show help |

## Credential Loading Priority

1. `GOOGLE_CREDENTIALS` environment variable (JSON content)
2. `GOOGLE_APPLICATION_CREDENTIALS` environment variable (file path)

## Examples

```bash
# List 50 inbox messages
gsuite messages list -n 50 --label-ids INBOX

# Get a message
gsuite messages get 18d5a1b2c3d4e5f6

# Mark as read
gsuite messages modify 18d5a1b2c3d4e5f6 --remove-labels UNREAD

# Send with markdown and attachments
gsuite send -t "user@example.com" -s "Report" -b "**Summary:**\n\n- Item one\n- Item two" --attach report.pdf

# Create and send a draft
gsuite drafts create -t "user@example.com" -s "Hello" -b "Draft content"
gsuite drafts send r1234567890

# List threads with search
gsuite threads list -q "from:alice@example.com" -n 20

# Manage labels
gsuite labels list
gsuite labels create -n "My Label"

# JSON output for scripting
gsuite messages list -f json
gsuite search "is:unread" -f json

# Calendar: list upcoming events
gsuite calendar list --after today --before +7d

# Calendar: create a recurring meeting with attendees
gsuite calendar create --summary "Weekly 1:1" --start "2026-03-15 10:00" \
  --duration 30m --rrule "FREQ=WEEKLY;BYDAY=MO" \
  --attendees "alice@example.com" --send-updates all

# Calendar: RSVP to an event
gsuite calendar respond abc123def456 --status accepted

# Calendar: delete a recurring event (all instances)
gsuite calendar delete abc123def456 --recurring-scope all --yes
```

## License

See LICENSE file for details.
