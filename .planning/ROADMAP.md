# Roadmap: Google Suite CLI

## Overview

Build a Go CLI for complete Gmail mailbox management via service account authentication. Start with project foundation and auth, then implement read operations (list, search, threads), followed by write operations (send, drafts, delete), finishing with polish (attachments, formatting).

## Domain Expertise

None

## Milestones

- ✅ **v1.0 MVP** — Phases 1-4 (shipped 2026-02-05)
- ✅ **v1.1 OAuth2 Support** — Phases 5-7 (shipped 2026-02-06)
- ✅ **v1.2 Headless Login** — Phase 8 (shipped 2026-02-06)
- 🚧 **v2.0 Auth Simplification** — Phases 9-10 (in progress)

## Completed Milestones

- ✅ [v1.0 MVP](milestones/v1.0-ROADMAP.md) (Phases 1-4) — SHIPPED 2026-02-05
- ✅ [v1.1 OAuth2 Support](milestones/v1.1-ROADMAP.md) (Phases 5-7) — SHIPPED 2026-02-06
- ✅ [v1.2 Headless Login](milestones/v1.2-ROADMAP.md) (Phase 8) — SHIPPED 2026-02-06

## Phases

<details>
<summary>✅ v1.0 MVP (Phases 1-4) — SHIPPED 2026-02-05</summary>

- [x] Phase 1: Foundation (3/3 plans) — completed 2026-02-04
- [x] Phase 2: Core Read Operations (3/3 plans) — completed 2026-02-05
- [x] Phase 3: Write Operations (3/3 plans) — completed 2026-02-05
- [x] Phase 4: Polish (2/2 plans) — completed 2026-02-05

</details>

<details>
<summary>✅ v1.1 OAuth2 Support (Phases 5-7) — SHIPPED 2026-02-06</summary>

- [x] Phase 5: OAuth2 Core (1/1 plan) — completed 2026-02-05
- [x] Phase 6: Auth Dispatcher (1/1 plan) — completed 2026-02-06
- [x] Phase 7: CLI Integration (1/1 plan) — completed 2026-02-06

</details>

<details>
<summary>✅ v1.2 Headless Login (Phase 8) — SHIPPED 2026-02-06</summary>

- [x] Phase 8: Device Authorization Flow (1/1 plan) — completed 2026-02-06

</details>

### 🚧 v2.0 Auth Simplification (In Progress)

**Milestone Goal:** Strip auth down to OAuth2 PKCE only — remove service account and device flow code, simplify CLI flags

#### Phase 9: Remove Auth Code

**Goal**: Strip service account JWT auth and device flow code from internal/auth, clean up login command
**Depends on**: Previous milestone complete
**Research**: Unlikely (internal cleanup, removing existing code)
**Plans**: TBD

Plans:
- [ ] 09-01: TBD (run /gsd:plan-phase 9 to break down)

#### Phase 10: Simplify CLI

**Goal**: Remove --credentials-file and --user flags, update all subcommands, add clear error messages for missing login
**Depends on**: Phase 9
**Research**: Unlikely (internal cleanup, simplifying flags)
**Plans**: TBD

Plans:
- [ ] 10-01: TBD (run /gsd:plan-phase 10 to break down)

## Progress

| Phase | Milestone | Plans Complete | Status | Completed |
|-------|-----------|----------------|--------|-----------|
| 1. Foundation | v1.0 | 3/3 | Complete | 2026-02-04 |
| 2. Core Read Operations | v1.0 | 3/3 | Complete | 2026-02-05 |
| 3. Write Operations | v1.0 | 3/3 | Complete | 2026-02-05 |
| 4. Polish | v1.0 | 2/2 | Complete | 2026-02-05 |
| 5. OAuth2 Core | v1.1 | 1/1 | Complete | 2026-02-05 |
| 6. Auth Dispatcher | v1.1 | 1/1 | Complete | 2026-02-06 |
| 7. CLI Integration | v1.1 | 1/1 | Complete | 2026-02-06 |
| 8. Device Authorization Flow | v1.2 | 1/1 | Complete | 2026-02-06 |
| 9. Remove Auth Code | v2.0 | 0/? | Not started | - |
| 10. Simplify CLI | v2.0 | 0/? | Not started | - |
