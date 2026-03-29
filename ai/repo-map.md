# Repository Map

```
xffl/
├── CLAUDE.md              → Agent entry point — rules, repo map, commands
├── justfile               → Task runner (just <recipe>)
├── go.work                → Go workspace referencing all modules
│
├── ai/                    → AI control plane (read-only for agents)
│   ├── architecture/      → principles.md, service-map.md, bounded-contexts.md
│   ├── decisions/         → ADR index (decisions.md) + individual ADRs
│   ├── plans/             → roadmap.md, current-sprint.md
│   └── prompts/           → Development workflow
│
├── ai-runtime/            → Agent working memory (gitignored)
│   └── current-task.md    → Per-task scratchpad
│
├── services/
│   ├── afl/               → AFL service (Go, GraphQL, :8080)
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
├── dev/
│   ├── docker-compose.yml → Postgres (:5432) + Zinc (:4080)
│   └── postgres/seed/     → SQL seed files
│
└── first-cut/             → Legacy prototype (reference only)
```
