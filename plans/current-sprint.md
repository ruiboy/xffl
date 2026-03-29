# Current Sprint

**Sprint goal:** Phase 4 — AFL Frontend (remaining views + UX polish)

## Tasks

### 1. AI plans & prompt improvements
- [x] Review and improve `plans/` structure and content
- [x] Review and improve `ai/prompts/system-prompt.md`
- [x] Ensure sprint/roadmap alignment

### 2. AFL frontend page discovery
- [x] Interview about required pages, navigation, and user flows
- [x] Document page inventory and key interactions
- [x] Confirm scope before building

### 3. Build AFL pages
- [x] Add `aflLatestRound` backend query + `season` field on `AFLRound`
- [x] Implement pages identified in step 2 (Home, Round, Match read-only, Admin Match)
- [x] Playwright e2e tests for each page (16 tests)

### 4. Apply UX style
- [x] Switch to light theme across all components and views
- [x] Add AFL club logos (18 teams, stored as static assets)
- [x] Add logos to MatchSummary, LadderTable, MatchView, AdminMatchView
- [ ] Leverage PrimeVue unstyled + Tailwind per ADR-011 (deferred)