# Current Sprint

**Sprint goal:** Phase 3 — UX Scaffold (Gateway + Vue 3 app + first AFL view with editing)

## Research

### Gateway approach (ADR-008)
- [ ] Research gateway options (string routing, query parsing, federation, reverse proxy)
- [ ] Evaluate against current needs (single service now, multi-service later)
- [ ] Update ADR-008 with decision

### Frontend component library (ADR-011)
- [ ] Research current component library options (PrimeVue, alternatives)
- [ ] Update ADR-011 with decision

## Tasks

### Gateway
- [ ] Implement gateway service based on ADR-008 decision
- [ ] GraphQL proxy routing to AFL service
- [ ] CORS configuration
- [ ] Health check endpoint
- [ ] Add to `docker-compose.yml` / `justfile`

### Vue 3 project setup
- [ ] Scaffold Vue 3 + TypeScript + Vite project
- [ ] Configure Apollo Client pointing at gateway (:8090)
- [ ] Set up Vue Router
- [ ] Install and configure PrimeVue
- [ ] Add to `justfile`

### AFL Match view
- [ ] Match result display with player stats
- [ ] Inline editing of player stats
- [ ] Wire up mutations through Apollo Client

### Tests
- [ ] Playwright setup
- [ ] Match view read tests
- [ ] Match view edit tests
