# Roadmap: Google Suite CLI

## Overview

Build a Go CLI for complete Gmail mailbox management via service account authentication. Start with project foundation and auth, then implement read operations (list, search, threads), followed by write operations (send, drafts, delete), finishing with polish (attachments, formatting).

## Domain Expertise

None

## Milestones

- ✅ **v1.0 MVP** — Phases 1-4 (shipped 2026-02-05)
- 🚧 **v1.1 OAuth2 Support** — Phases 5-7 (in progress)

## Completed Milestones

- ✅ [v1.0 MVP](milestones/v1.0-ROADMAP.md) (Phases 1-4) — SHIPPED 2026-02-05

## Phases

<details>
<summary>✅ v1.0 MVP (Phases 1-4) — SHIPPED 2026-02-05</summary>

- [x] Phase 1: Foundation (3/3 plans) — completed 2026-02-04
- [x] Phase 2: Core Read Operations (3/3 plans) — completed 2026-02-05
- [x] Phase 3: Write Operations (3/3 plans) — completed 2026-02-05
- [x] Phase 4: Polish (2/2 plans) — completed 2026-02-05

</details>

### 🚧 v1.1 OAuth2 Support (In Progress)

**Milestone Goal:** Add OAuth2 browser-based login flow for personal Gmail accounts alongside existing service account auth

#### Phase 5: OAuth2 Core — Complete

**Goal**: Implement token storage, browser opener, and OAuth2 authorization code flow with PKCE
**Depends on**: v1.0 complete
**Research**: Unlikely (Go stdlib oauth2 already in project)
**Plans**: 1

Plans:
- [x] 05-01: OAuth2 PKCE flow + token storage — completed 2026-02-05

#### Phase 6: Auth Dispatcher — Complete

**Goal**: Refactor auth.go to auto-detect credential type from JSON and branch between service account / OAuth flows
**Depends on**: Phase 5
**Research**: Unlikely (internal refactor)
**Plans**: 1

Plans:
- [x] 06-01: Credential type detection + auth dispatcher — completed 2026-02-06

#### Phase 7: CLI Integration

**Goal**: Add `login` command, remove `--user` guards from 7 command files, update help text
**Depends on**: Phase 6
**Research**: Unlikely (internal patterns)
**Plans**: TBD

Plans:
- [ ] 07-01: TBD (run /gsd:plan-phase 7 to break down)

## Progress

| Phase | Milestone | Plans Complete | Status | Completed |
|-------|-----------|----------------|--------|-----------|
| 1. Foundation | v1.0 | 3/3 | Complete | 2026-02-04 |
| 2. Core Read Operations | v1.0 | 3/3 | Complete | 2026-02-05 |
| 3. Write Operations | v1.0 | 3/3 | Complete | 2026-02-05 |
| 4. Polish | v1.0 | 2/2 | Complete | 2026-02-05 |
| 5. OAuth2 Core | v1.1 | 1/1 | Complete | 2026-02-05 |
| 6. Auth Dispatcher | v1.1 | 1/1 | Complete | 2026-02-06 |
| 7. CLI Integration | v1.1 | 0/? | Not started | - |
