# Repository Map

```
xffl/
├── CLAUDE.md              → Agent entry point — rules, repo map, commands
├── justfile               → Task runner (just <recipe>)
├── go.work                → Go workspace referencing all modules
│
├── ai/                    → AI control plane (read-only for agents)
│   ├── architecture/      → principles.md, service-map.md, domain.md, cookbook.md
│   ├── decisions/         → ADR index (decisions.md) + individual ADRs
│   └── prompts/           → Development workflow
│
├── plans/                 → Project plans
│   ├── roadmap.md         → Full project phases
│   ├── current-sprint.md  → Active sprint tasks
│   └── revisit.md         → Ideas to reconsider later (not roadmap)
│
├── services/
│   ├── afl/               → AFL service (Go, GraphQL, :8080)
│   │   ├── cmd/           → Entrypoint
│   │   └── internal/      → domain/ → application/ → infrastructure/ → interface/
│   ├── ffl/               → FFL service (Go, GraphQL, :8081)
│   │   ├── cmd/           → Entrypoint
│   │   └── internal/      → domain/ → application/ → infrastructure/ → interface/
│   └── gateway/           → Reverse proxy (:8090)
│
├── frontend/web/          → Vue 3 SPA (TypeScript, Vite, :3000)
│   ├── src/               → Components, views, router, Apollo client
│   └── e2e/               → Playwright tests
│
├── contracts/events/      → Shared event type definitions (Go)
├── shared/                → Shared Go packages
│   ├── database/          → DB connection helper
│   └── events/            → EventDispatcher interface + PG LISTEN/NOTIFY + in-memory
│
└── dev/
    ├── docker-compose.yml → Postgres (:5432) + Typesense (:8108)
    └── postgres/seed/     → SQL seed files
```
