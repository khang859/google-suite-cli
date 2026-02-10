# gsuite Command Reference

## Global Flags

Available on all commands:

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--account` | | | Use specific account email (overrides active account) |
| `--format` | `-f` | `text` | Output format: `text` or `json` |
| `--verbose` | `-v` | `false` | Enable verbose output |

The `--account` flag can also be set via the `GSUITE_ACCOUNT` environment variable.

## Authentication

### `gsuite login`

Authenticate with Gmail using OAuth2. Opens browser for consent flow.
You can login with multiple accounts — the most recently logged-in becomes active.

Requires credentials via `GOOGLE_CREDENTIALS` env var (raw JSON) or
`GOOGLE_APPLICATION_CREDENTIALS` env var (file path).

```bash
gsuite login
```

### `gsuite logout [email]`

Remove an authenticated account and its stored token. If no email is given,
logs out the active account. If other accounts remain, the next available
account becomes active.

```bash
gsuite logout
gsuite logout other@gmail.com
```

## Accounts

### `gsuite accounts list`

List all authenticated accounts. The active account is marked with `*`.

```bash
gsuite accounts list
gsuite accounts list -f json
```

### `gsuite accounts switch <email>`

Switch the active account. The email must be an already authenticated account.

```bash
gsuite accounts switch user@gmail.com
```

### `gsuite accounts remove <email>`

Remove an authenticated account and its stored token. If the removed account
was active, another account is set as active automatically.

```bash
gsuite accounts remove user@gmail.com
```

## Profile

### `gsuite whoami`

Show authenticated user's Gmail profile (email, message count, thread count).

```bash
gsuite whoami
gsuite whoami -f json
```

## Messages

### `gsuite messages list`

List messages in the mailbox.

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--max-results` | `-n` | `10` | Max messages to return (max 500) |
| `--label-ids` | | | Comma-separated label IDs (e.g., `INBOX,UNREAD`) |
| `--query` | `-q` | | Gmail search query string |

```bash
gsuite messages list
gsuite messages list -n 50 --label-ids INBOX
gsuite messages list -q "from:user@example.com" -f json
gsuite messages list --label-ids INBOX,UNREAD -n 20
```

### `gsuite messages get <message-id>`

Retrieve and display a message. Shows From, To, Subject, Date, body, and attachment info.

```bash
gsuite messages get 18d5a1b2c3d4e5f6
gsuite messages get 18d5a1b2c3d4e5f6 -f json
```

### `gsuite messages modify <message-id>`

Add or remove labels from a message. At least one of `--add-labels` or `--remove-labels` is required.

| Flag | Description |
|------|-------------|
| `--add-labels` | Comma-separated label IDs to add |
| `--remove-labels` | Comma-separated label IDs to remove |

```bash
# Mark as read
gsuite messages modify <id> --remove-labels UNREAD

# Archive
gsuite messages modify <id> --remove-labels INBOX

# Star
gsuite messages modify <id> --add-labels STARRED

# Add custom label and mark read
gsuite messages modify <id> --add-labels Label_123 --remove-labels UNREAD

# Add multiple labels
gsuite messages modify <id> --add-labels Label_1,Label_2,STARRED
```

### `gsuite messages get-attachment <message-id> <attachment-id>`

Download an attachment from a message.

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--output` | `-o` | original filename | Output file path |

```bash
gsuite messages get-attachment 18d5a1b2c3d4e5f6 ANGjdJ8abc123
gsuite messages get-attachment 18d5a1b2c3d4e5f6 ANGjdJ8abc123 -o ./report.pdf
```

## Threads

### `gsuite threads list`

List conversation threads.

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--max-results` | `-n` | `10` | Max threads to return (max 500) |
| `--label-ids` | | | Comma-separated label IDs |
| `--query` | `-q` | | Gmail search query |

```bash
gsuite threads list
gsuite threads list -n 20 -q "from:alice@example.com"
gsuite threads list --label-ids INBOX,UNREAD -f json
```

### `gsuite threads get <thread-id>`

Get a thread with all messages in chronological order.

```bash
gsuite threads get 18d1234567890abc
gsuite threads get 18d1234567890abc -f json
```

## Search

### `gsuite search <query>`

Search messages using Gmail query syntax.

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--max-results` | `-n` | `10` | Max results (1-500) |
| `--label-ids` | | | Comma-separated label IDs to filter by |

```bash
gsuite search "from:user@example.com"
gsuite search "subject:meeting" -n 20
gsuite search "is:unread" --label-ids INBOX
gsuite search "newer_than:1d"
gsuite search "has:attachment filename:pdf"
```

## Labels

### `gsuite labels list`

List all labels (system labels first, then user labels alphabetically).

```bash
gsuite labels list
gsuite labels list -f json
```

### `gsuite labels create`

Create a new label.

| Flag | Short | Required | Description |
|------|-------|----------|-------------|
| `--name` | `-n` | Yes | Display name for the label |
| `--label-list-visibility` | | No | `labelShow`, `labelShowIfUnread`, `labelHide` |
| `--message-list-visibility` | | No | `show`, `hide` |

```bash
gsuite labels create -n "My Label"
gsuite labels create -n "Work" --label-list-visibility labelShow
```

### `gsuite labels update <label-id>`

Update a user-created label. System labels cannot be updated.

| Flag | Short | Description |
|------|-------|-------------|
| `--name` | `-n` | New display name |
| `--label-list-visibility` | | `labelShow`, `labelShowIfUnread`, `labelHide` |
| `--message-list-visibility` | | `show`, `hide` |

```bash
gsuite labels update Label_123 -n "New Name"
gsuite labels update Label_123 --label-list-visibility labelHide
```

### `gsuite labels delete <label-id>`

Delete a user-created label. System labels cannot be deleted. Messages with this
label are not deleted — only the label is removed from them.

```bash
gsuite labels delete Label_123
```

## Drafts

### `gsuite drafts list`

List drafts.

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--max-results` | `-n` | `10` | Max drafts to return (max 500) |

```bash
gsuite drafts list
gsuite drafts list -n 50 -f json
```

### `gsuite drafts get <draft-id>`

Retrieve and display a draft.

```bash
gsuite drafts get r1234567890123456789
```

### `gsuite drafts create`

Create a new draft.

| Flag | Short | Required | Description |
|------|-------|----------|-------------|
| `--to` | `-t` | Yes | Recipient email |
| `--subject` | `-s` | Yes | Subject line |
| `--body` | `-b` | Yes | Plain text body |
| `--cc` | | No | CC recipients (comma-separated) |
| `--bcc` | | No | BCC recipients (comma-separated) |

```bash
gsuite drafts create -t "user@example.com" -s "Hello" -b "Draft content"
gsuite drafts create -t "user@example.com" -s "Meeting" -b "Let's meet" --cc "cc@example.com"
```

### `gsuite drafts update <draft-id>`

Update an existing draft. Unmodified fields are preserved.

| Flag | Short | Description |
|------|-------|-------------|
| `--to` | `-t` | New recipient |
| `--subject` | `-s` | New subject |
| `--body` | `-b` | New body |
| `--cc` | | New CC recipients |
| `--bcc` | | New BCC recipients |

```bash
gsuite drafts update r1234567890 --subject "Updated Subject"
gsuite drafts update r1234567890 -t "new@example.com" -b "New content"
```

### `gsuite drafts send <draft-id>`

Send a draft. The draft is removed from the drafts folder after sending.

```bash
gsuite drafts send r1234567890123456789
```

### `gsuite drafts delete <draft-id>`

Permanently delete a draft. Cannot be undone.

```bash
gsuite drafts delete r1234567890123456789
```

## Send

### `gsuite send`

Send an email as multipart/alternative (plain text + HTML) for best rendering
across email clients. The body supports markdown formatting (bold, italic, links,
lists, code, strikethrough, tables) which is rendered as HTML. Use `\n` for line
breaks. Supports attachments.

| Flag | Short | Required | Description |
|------|-------|----------|-------------|
| `--to` | `-t` | Yes | Recipient email |
| `--subject` | `-s` | Yes | Subject line |
| `--body` | `-b` | Yes | Body content with markdown support (`\n` for line breaks) |
| `--cc` | | No | CC recipients (comma-separated) |
| `--bcc` | | No | BCC recipients (comma-separated) |
| `--attach` | `-a` | No | File to attach (repeatable) |

```bash
gsuite send -t "user@example.com" -s "Hello" -b "Hi,\n\nHow are you?\nBest regards"
gsuite send -t "user@example.com" -s "Update" -b "**Bold** and *italic*\n\n- Item one\n- Item two\n\nVisit [Google](https://google.com)"
gsuite send -t "user@example.com" -s "Report" -b "See attached.\n\nThanks" --attach report.pdf --attach data.csv
```

## Gmail Search Query Syntax

The `search` command and `messages list -q` / `threads list -q` all accept Gmail
query syntax. Common operators:

| Operator | Example | Description |
|----------|---------|-------------|
| `from:` | `from:user@example.com` | Messages from a sender |
| `to:` | `to:user@example.com` | Messages to a recipient |
| `subject:` | `subject:meeting` | Subject contains word |
| `is:unread` | `is:unread` | Unread messages |
| `is:starred` | `is:starred` | Starred messages |
| `is:read` | `is:read` | Read messages |
| `has:attachment` | `has:attachment` | Has attachments |
| `filename:` | `filename:pdf` | Attachment filename/type |
| `newer_than:` | `newer_than:7d` | Newer than N days/months/years |
| `older_than:` | `older_than:1y` | Older than N days/months/years |
| `after:` | `after:2024/01/01` | After a date |
| `before:` | `before:2024/12/31` | Before a date |
| `label:` | `label:work` | Has a specific label |
| `in:` | `in:inbox` | In a specific folder |
| `OR` | `from:a OR from:b` | Match either condition |
| `-` | `-from:spam@example.com` | Exclude matches |
| `()` | `(from:a OR from:b) subject:hi` | Group conditions |

## System Label IDs

These are the built-in Gmail labels. Use these IDs with `--label-ids`, `--add-labels`, and `--remove-labels`:

| Label ID | Description |
|----------|-------------|
| `INBOX` | Inbox |
| `SENT` | Sent mail |
| `TRASH` | Trash |
| `SPAM` | Spam |
| `DRAFT` | Drafts |
| `STARRED` | Starred |
| `UNREAD` | Unread |
| `IMPORTANT` | Important |
| `CATEGORY_PERSONAL` | Primary/Personal category |
| `CATEGORY_SOCIAL` | Social category |
| `CATEGORY_PROMOTIONS` | Promotions category |
| `CATEGORY_UPDATES` | Updates category |
| `CATEGORY_FORUMS` | Forums category |
