# Current Sprint

**Sprint goal:** Phase 3 — UX Scaffold (Gateway + Vue 3 app + first AFL view with editing)

## Research

### Gateway approach (ADR-008)
- [x] Research gateway options (string routing, query parsing, federation, reverse proxy)
- [x] Evaluate against current needs (single service now, multi-service later)
- [x] Update ADR-008 with decision

### Frontend component library (ADR-011)
- [x] Research current component library options (PrimeVue, alternatives)
- [x] Update ADR-011 with decision

## Tasks

### Gateway
- [x] Implement gateway service based on ADR-008 decision
- [x] GraphQL proxy routing to AFL service
- [x] CORS configuration
- [x] Health check endpoint
- [x] Add to `docker-compose.yml` / `justfile`

### Vue 3 project setup (pass 1)
- [x] Scaffold Vue 3 + TypeScript + Vite project
- [x] Configure Apollo Client pointing at gateway (:8090)
- [x] Set up Vue Router
- [x] Configure Tailwind CSS
- [x] `justfile` already configured
- [ ] Install PrimeVue unstyled (pass 2 — when first view needs it)
- [ ] Install Pinia (pass 2 — when UI state management needed)
- [ ] Install VueUse (pass 2 — when reactive utilities needed)

### AFL Match view
- [x] Match result display with player stats
- [x] Inline editing of player stats
- [x] Wire up mutations through Apollo Client

### Tests
- [x] Playwright setup
- [x] Match view read tests
- [x] Match view edit tests
