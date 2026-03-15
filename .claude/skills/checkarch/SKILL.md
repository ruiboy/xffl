---
name: checkarch
description: Validate code against architecture principles and ADRs (clean architecture layers, DDD, service layout)
disable-model-invocation: true
---

Validate code against the architecture principles defined in `ai/architecture/principles.md` and the ADRs in `ai/decisions/`.

## Steps

1. Read `ai/architecture/principles.md` to load the current rules.
2. Read all `ai/decisions/adr-*.md` files. Extract only rules that are **mechanically verifiable from code** (imports, directory structure, package boundaries, naming). Skip rules that are purely about technology choice or process.
3. Ask the user what scope to check:
   - **Uncommitted changes** — `git diff HEAD`
   - **Current branch** — `git diff main...HEAD` (all commits since branching from main)
   - **A specific service** — all code under `services/<name>/`
   - **Entire codebase** — all code under `services/` and `shared/`
4. Collect all Go files in scope. Skip generated files (e.g. `*_gen.go`, `generated.go`, `models_gen.go`).
5. Check each file against these rules from `principles.md`. Common examples:
   - **Dependency direction** — domain must not import application, infrastructure, or interface packages. Application must not import infrastructure or interface. Infrastructure must not import interface.
   - **Service layout** — each service follows the directory structure in principles (`cmd/`, `internal/domain/`, `internal/application/`, `internal/infrastructure/`, `internal/interface/`).
   - **Domain purity** — domain packages have no framework or external dependencies (no database drivers, HTTP libraries, etc.). Only stdlib and other domain packages.
   - **Repository pattern** — repository interfaces are defined in domain, implementations in infrastructure.
   - **No shared DB** — services do not import another service's internal packages.
   - **GraphQL roots** — query root types correspond to aggregates, not join entities.
   - **Testing rules** — domain tests have no mocks or infrastructure imports. Interface tests are integration tests (not unit tests with mocked dependencies).
6. Additionally check against any verifiable rules extracted from ADRs in step 2. Common examples:
   - **ADR-001 (repo layout)** — services live under `services/`, shared code under `shared/`
   - **ADR-002 (GraphQL per service)** — each service has its own GraphQL schema, no shared schema files
   - **ADR-003 (shared DB schema isolation)** — no cross-service database imports or shared DB connection packages
   - **ADR-005 (clean architecture layers)** — covered by principles checks above, verify no additional constraints from the ADR
   - **ADR-007 (Go workspace)** — `go.work` exists and references all service modules
   - **ADR-009 (DB persistence layer)** — persistence implementations live in infrastructure, not application or domain
   - Other ADRs — extract and check any rule that can be verified from code structure or imports; skip the rest
7. Present a summary of violations grouped by source (principles vs ADR number), with file paths and line numbers.
8. If violations are found, ask the user: "Want me to fix these violations?"
9. If yes, fix them and show the changes for approval before applying.