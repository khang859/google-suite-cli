# Roadmap: Google Suite CLI

## Overview

Build a Go CLI for complete Gmail mailbox management via service account authentication. Start with project foundation and auth, then implement read operations (list, search, threads), followed by write operations (send, drafts, delete), finishing with polish (attachments, formatting).

## Domain Expertise

None

## Phases

**Phase Numbering:**
- Integer phases (1, 2, 3): Planned milestone work
- Decimal phases (2.1, 2.2): Urgent insertions (marked with INSERTED)

Decimal phases appear between their surrounding integers in numeric order.

- [x] **Phase 1: Foundation** - Go project setup, CLI framework, service account authentication ✓
- [ ] **Phase 2: Core Read Operations** - Messages list/get, threads, search, labels read
- [ ] **Phase 3: Write Operations** - Send messages, drafts, labels management, delete
- [ ] **Phase 4: Polish** - Attachments, output formatting, error handling refinement

## Phase Details

### Phase 1: Foundation
**Goal**: Working CLI skeleton with authenticated Gmail API connection
**Depends on**: Nothing (first phase)
**Research**: Likely (Google API setup)
**Research topics**: Go Gmail API client library, service account auth with domain-wide delegation, credential loading (env var vs file path)
**Plans**: TBD

Plans:
- [x] 01-01: Project structure, Go modules, CLI framework setup ✓
- [x] 01-02: Service account authentication with flexible credential handling ✓
- [x] 01-03: Basic Gmail API connection verification ✓

### Phase 2: Core Read Operations
**Goal**: Complete read-only Gmail access (list, get, search, labels, threads)
**Depends on**: Phase 1
**Research**: Unlikely (uses patterns from Phase 1)
**Plans**: TBD

Plans:
- [ ] 02-01: Messages list and get commands
- [ ] 02-02: Search and labels list commands
- [ ] 02-03: Threads list and get commands

### Phase 3: Write Operations
**Goal**: Full write capabilities (send, drafts, labels CRUD, delete)
**Depends on**: Phase 2
**Research**: Unlikely (internal patterns)
**Plans**: TBD

Plans:
- [ ] 03-01: Send message command
- [ ] 03-02: Draft create/update/send/delete commands
- [ ] 03-03: Labels create/update/delete and message label operations

### Phase 4: Polish
**Goal**: Production-ready CLI with attachments, formatting, and robust error handling
**Depends on**: Phase 3
**Research**: Unlikely (refinement work)
**Plans**: TBD

Plans:
- [ ] 04-01: Attachment download and send-with-attachment
- [ ] 04-02: Output formatting (human-readable and JSON modes)

## Progress

**Execution Order:**
Phases execute in numeric order: 1 → 2 → 3 → 4

| Phase | Plans Complete | Status | Completed |
|-------|----------------|--------|-----------|
| 1. Foundation | 3/3 | Complete | 2026-02-04 |
| 2. Core Read Operations | 0/3 | Not started | - |
| 3. Write Operations | 0/3 | Not started | - |
| 4. Polish | 0/2 | Not started | - |
