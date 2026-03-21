---
name: checkarch
description: Validate code against architecture principles and ADRs (clean architecture layers, DDD, service layout)
disable-model-invocation: true
---

Validate code against the architecture principles defined in `ai/architecture/principles.md` and the ADRs in `ai/decisions/`.

## Steps

1. Read `ai/architecture/principles.md` to load the current rules.
2. Read all `ai/decisions/adr-*.md` files. Focus on ADRs with `enforceable: true` in frontmatter. Use the `rules:` field as your primary checklist for that ADR. Still apply judgment — rules are plain language, not exact patterns. Skip ADRs with `enforceable: false`.
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
6. Additionally check against the `rules:` listed in each enforceable ADR's frontmatter. Use the rules as a checklist and apply judgment to verify each one from code structure, imports, and file locations.
7. Present a summary of violations grouped by source (principles vs ADR number), with file paths and line numbers.
8. If violations are found, ask the user: "Want me to fix these violations?"
9. If yes, fix them and show the changes for approval before applying.