# ADR-009: Database Persistence Layer

**Status:** Accepted
**Date:** 2026-03-15

## Context

Services need to interact with PostgreSQL. First-cut used GORM, which introduced struct-tag magic that fights clean architecture and is not idiomatic Go.

## Options

1. **GORM** — full ORM, struct tags, familiar from first-cut. Magic can be surprising; implicit SQL makes debugging harder.
2. **sqlc + pgx** — write SQL, generate type-safe Go. No magic, compile-time safety, Go-idiomatic. Generated models naturally land in infrastructure layer.
3. **Raw pgx** — maximum control, significant boilerplate for little gain over sqlc.

## Decision

**sqlc + pgx** with a thin helper layer in `shared/database/`.

### Persistence

- SQL-first: all queries are hand-written SQL, sqlc generates type-safe Go code.
- Migrations are plain SQL files.
- Domain entities remain pure — sqlc-generated DB models and converters live in the infrastructure layer (ADR-005).

### sqlc Configuration

```yaml
emit_interface: true              # Querier interface for mocking
emit_empty_slices: true           # [] not null for empty results
emit_pointers_for_null_types: true # *string not sql.NullString
```

### Transaction Boundaries

The **application layer** (use cases) owns the transaction lifecycle. It is the only layer that knows which repository calls belong to a single unit of work.

- Domain layer: repository interfaces with no transaction awareness.
- Application layer: starts/commits/rolls back via `DB.WithTx()`.
- Infrastructure layer: provides the `WithTx` implementation wrapping `pgx.BeginTx`.

### Helper Surface (`shared/database/`)

| Helper | Purpose |
|---|---|
| `DB.WithTx(ctx, func(q) error)` | Transaction lifecycle — begin, commit, rollback |
| `DB.Queries()` | Read path, no transaction |
| `MapPgError(err)` | Translate pgx errors (ErrNoRows, unique/FK violations) to domain errors |

### Error Mapping

`MapPgError` translates infrastructure errors to domain errors so pgx concerns do not leak beyond the infrastructure layer:

- `pgx.ErrNoRows` → `domain.ErrNotFound`
- PG `23505` (unique violation) → `domain.ErrConflict`
- PG `23503` (FK violation) → `domain.ErrInvalidRef`

## Rationale

- **SQL-first** aligns with "no magic" and Go idioms.
- **Compile-time query validation** catches bugs early.
- **Clean architecture by default** — generated DB models stay in infrastructure; domain entities stay pure.
- **Thin helper layer** reduces boilerplate without becoming an abstraction trap.
- **Easy escape hatch** — dropping to raw pgx for specific queries is trivial since sqlc uses pgx underneath.