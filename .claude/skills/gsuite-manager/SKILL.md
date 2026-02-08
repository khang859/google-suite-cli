---
name: gsuite-manager
description: >-
  This skill should be used when managing Gmail through the gsuite CLI tool.
  It applies when the user asks to read, send, search, label, archive, or
  otherwise manage their Gmail — including messages, threads, labels, drafts,
  and attachments. Trigger keywords include email, Gmail, inbox, send, draft,
  label, search mail, unread, archive, attachment.
---

# Gmail Manager via gsuite CLI

Manage Gmail accounts through the `gsuite` command-line tool. This skill covers
authentication, reading mail, sending emails, organizing with labels, managing
drafts, searching, and inbox cleanup.

## Prerequisites

The `gsuite` binary must be installed and available on PATH. To verify:

```bash
gsuite --help
```

If not installed, build from source or download from releases. See the project
README for installation instructions.

## Authentication

Before any Gmail operation, verify authentication status:

```bash
gsuite whoami
```

If this fails, the user needs to authenticate. **Prefer OAuth2 for personal accounts.**

### OAuth2 Login (Personal Gmail)

Requires an OAuth2 client credentials JSON file (not a service account).

```bash
# Browser-based login (default)
gsuite login -c /path/to/oauth2-client.json

# Headless environments (SSH, containers)
gsuite login -c /path/to/oauth2-client.json --no-browser
```

After login, the token is cached at `~/.config/gsuite/token.json` and subsequent
commands work without re-authenticating.

To log out:

```bash
gsuite logout
```

### Service Account (Google Workspace)

For workspace environments with domain-wide delegation, pass credentials and user:

```bash
gsuite --credentials-file /path/to/sa.json --user user@domain.com whoami
```

Or set environment variables:

```bash
export GOOGLE_APPLICATION_CREDENTIALS=/path/to/sa.json
gsuite --user user@domain.com whoami
```

## Safety Rules

**CRITICAL: Confirm with the user before executing any destructive action.**

Destructive actions that MUST be confirmed:
- `gsuite send` — sending an email (cannot be unsent)
- `gsuite drafts send` — sending a draft (removes it from drafts)
- `gsuite drafts delete` — permanently deletes a draft
- `gsuite labels delete` — permanently deletes a label
- `gsuite messages modify` with `--remove-labels` — removing labels from messages

Safe read-only actions that do NOT need confirmation:
- `whoami`, `messages list`, `messages get`, `threads list`, `threads get`
- `search`, `labels list`, `drafts list`, `drafts get`
- `messages get-attachment` (downloads a file, low risk)

Medium-risk actions — confirm if the scope is large:
- `gsuite labels create` — creating labels
- `gsuite labels update` — renaming labels
- `gsuite drafts create` / `drafts update` — creating or editing drafts
- `gsuite messages modify` with `--add-labels` only — adding labels

## Output Format

Always use `--format json` (`-f json`) when processing results programmatically.
JSON mode provides structured output that is easier to parse and chain.

Use text mode (default) when displaying results directly to the user.

## Command Reference

For detailed command syntax, flags, and examples, read the reference file at
`references/commands.md` within this skill directory.

## Common Workflows

### Check Inbox

```bash
gsuite messages list --label-ids INBOX -n 20
```

To see only unread:

```bash
gsuite messages list --label-ids INBOX,UNREAD -n 20
```

### Read a Message

```bash
gsuite messages get <message-id>
```

### Search for Emails

```bash
gsuite search "from:boss@company.com newer_than:7d"
gsuite search "subject:invoice has:attachment"
gsuite search "is:unread in:inbox"
```

### Send an Email

Emails are sent as multipart/alternative (plain text + HTML) for best rendering.
The body supports markdown formatting (bold, italic, links, lists, code, strikethrough,
tables) which is rendered as HTML. Use `\n` for line breaks. After confirming with the user:

```bash
gsuite send --to "recipient@example.com" --subject "Subject" --body "Hello,\n\nHow are you?\nBest regards"
```

With markdown formatting:

```bash
gsuite send -t "user@example.com" -s "Update" -b "**Important:** Review the *report*.\n\n- Item one\n- Item two\n\nVisit [Google](https://google.com)"
```

With attachments:

```bash
gsuite send -t "user@example.com" -s "Report" -b "See attached.\n\nThanks" --attach report.pdf
```

### Organize with Labels

List existing labels to find IDs:

```bash
gsuite labels list
```

Create a new label:

```bash
gsuite labels create -n "Project/Alpha"
```

Apply a label to a message:

```bash
gsuite messages modify <message-id> --add-labels Label_123
```

Mark as read:

```bash
gsuite messages modify <message-id> --remove-labels UNREAD
```

Archive (remove from inbox):

```bash
gsuite messages modify <message-id> --remove-labels INBOX
```

### Find TODOs and Action Items

Search for emails that likely contain action items:

```bash
gsuite search "subject:(todo OR action OR follow up OR action required) newer_than:30d" -n 50
```

Then read individual messages to extract the actual tasks:

```bash
gsuite messages get <message-id>
```

### Clean Up Inbox

To mark multiple messages as read, archive, or label them, first search or list
to get message IDs, then modify each one. Chain operations for efficiency:

```bash
# Get unread inbox messages as JSON for processing
gsuite messages list --label-ids INBOX,UNREAD -n 50 -f json
```

Then for each message ID, apply the appropriate action.

### Manage Drafts

```bash
gsuite drafts list
gsuite drafts create -t "user@example.com" -s "Subject" -b "Body"
gsuite drafts send <draft-id>      # Confirm first!
gsuite drafts delete <draft-id>    # Confirm first!
```

### Download Attachments

First view the message to see attachment IDs:

```bash
gsuite messages get <message-id>
```

Then download:

```bash
gsuite messages get-attachment <message-id> <attachment-id> --output ./downloads/file.pdf
```

## Troubleshooting

**"no credentials provided"** — User needs to authenticate. Guide them through
OAuth2 login or service account setup.

**"authentication failed"** — Credentials are invalid or expired. Try `gsuite logout`
then `gsuite login` again.

**"Gmail API error: 403"** — Insufficient permissions. The OAuth2 scope or service
account delegation may not include the required Gmail scope.

**"Gmail API error: 404"** — Message or label not found. The ID may be wrong or
the item was already deleted.
