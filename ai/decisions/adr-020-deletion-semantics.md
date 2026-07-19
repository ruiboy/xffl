---
status: proposed
date: 2026-07-19
scope: infrastructure
enforceable: true
rules:
  - "repositories delete aggregate children explicitly, in dependency order, inside the aggregate's transaction — deletion order lives in code, not DDL"
  - "foreign keys use RESTRICT (NO ACTION); ON DELETE CASCADE is not used"
  - "deleted_at is opt-in per entity with a stated reason, not a default column on every table"
  - "any table retaining deleted_at must scope its unique constraints with WHERE deleted_at IS NULL"
---

# ADR-020: Deletion Semantics — Explicit Aggregate Deletes, RESTRICT over CASCADE, Soft Delete by Exception

## Context

Phase 25 work on the FFL fixture builder exposed three related problems that all trace back to deletion policy being implicit rather than designed.

**1. Soft-delete tombstones collide with unique constraints.** `SaveFixtures` previously rebuilt a round by soft-deleting every match and club_match and reinserting them. Moving to in-place reconciliation surfaced the reason that churn had "worked": `uni_ffl_club_match UNIQUE (club_season_id, match_id)` does not exclude `deleted_at`, so a tombstoned row permanently blocks that club from rejoining that match. The old code dodged this only by minting a fresh `match_id` on every save. A survey found this is schema-wide, not a one-off:

- 6 unique constraints in `ffl`, **none** scoped to live rows
- **0** partial (`WHERE deleted_at IS NULL`) indexes in the schema
- `deleted_at` present on essentially every table

**2. `ON DELETE CASCADE` encodes policy that contradicts the domain.** `player_match.club_match_id` cascades. The fixture builder's domain guard explicitly refuses to modify or delete a round that has any `player_match` (a submitted team). So the DDL says "teams evaporate with their fixture" while the domain says "a fixture with teams is immutable." Today the guard runs first and the cascade is inert, but the two are contradictory authorities and the DB wins silently. A future guard bug becomes silent data loss rather than an error. There are 24 such cascades (13 `afl`, 11 `ffl`).

**3. Deletion behaviour is invisible at the call site.** With cascades doing the work, "what happens when a match is deleted" is discoverable only by reading DDL, not by reading the repository. That fights ADR-009's existing rule that *SQL reads and writes data only and must not contain business logic* — a referential action is behaviour, not a read or a write.

Worth noting: soft delete is already not universal in practice. `player_match` has hard deletes (`DeletePlayerMatchesByClubMatchID`, used when a resubmitted team replaces the previous one). The convention is already applied inconsistently, which is itself a signal it was never a deliberate decision.

## Options

1. **Status quo** — keep blanket `deleted_at` plus cascades. Rejected: the tombstone/unique-constraint trap is latent in five more places, and the cascade/guard contradiction stays.

2. **Keep cascades, add soft-delete orchestration** — soft-delete children whenever a parent is soft-deleted, in the repository. Rejected: this is the expensive path. It needs per-table cascade logic, and every read path must then walk ancestors to check none is tombstoned, or orphans surface inside aggregates. Cost lands on reads, permanently, and grows combinatorially with the schema.

3. **Explicit repository-orchestrated deletes, RESTRICT foreign keys, soft delete by exception.** Chosen.

## Decision

### Repositories own deletion

A repository method that deletes an aggregate deletes its children explicitly, in dependency order, within the caller's transaction. Deletion order is expressed in Go, greppable and unit-testable, not inferred from DDL.

### Foreign keys are RESTRICT

`ON DELETE CASCADE` is removed. Foreign keys fall back to the default `NO ACTION`/`RESTRICT`. The database stops being an actor and becomes a safety net: if a repository forgets a child, the delete fails loudly at the exact call site instead of silently destroying data.

This is the operative test for whether a child should block its parent's deletion:

- **Child is an unconditional structural dependent** (a `club_match` has no meaning without its `match`) — the repository deletes it first, then the parent. RESTRICT never fires in correct code.
- **Child's presence is a reason to refuse deletion** (a `player_match` means a team was submitted) — the domain guard refuses first, and RESTRICT enforces the same rule at the storage layer as a backstop.

Both cases want RESTRICT. The difference is whose job it is to clear the child, not whether the DB should act.

### Soft delete is opt-in, per entity, with a reason

`deleted_at` stops being a default column. An entity carries it only where there is a stated need, and that need must be a domain one — not "we soft-delete everything." The questions to answer before adding it:

- Is this *audit/history*? Then it belongs in an event log, not a tombstone in an operational table.
- Is this *undo* or a real lifecycle state? Then model it in the domain (`archived`, `withdrawn`, `cancelled`) with explicit rules, not an infrastructure flag.
- Is this *oops-protection*? That is what backups are for.

The tell that a blanket `deleted_at` is a technical reflex rather than a domain concept: "should the children be soft-deleted too?" is a domain question, and a blanket infrastructure column has no answer to it. Entities that genuinely need soft delete each have their own answer.

### Unique constraints on soft-deleted tables must be partial

Any table that keeps `deleted_at` must scope its unique constraints to live rows (`UNIQUE ... WHERE deleted_at IS NULL`), or tombstones will block legitimate re-creation. This is the concrete bug found above.

## Rationale

The decisive argument is the failure mode, not purity:

| | Repository forgets a child |
|---|---|
| CASCADE | Silent data loss |
| RESTRICT | Loud FK error at the call site |

An earlier draft of this position allowed CASCADE *within* an aggregate boundary, on the grounds that a `club_match` has no identity outside its `match`, so removing it with the parent is definitional rather than a business decision — and hiding that inside the repository contract is legitimate. That reasoning still holds in isolation, but it is dominated: once RESTRICT plus explicit repository deletion is on the table, it achieves the same outcome *and* converts every future mistake from silent to loud. There is no case where CASCADE's convenience outweighs that.

Secondary benefits:

- **Cohesion.** "What happens when a match dies" reads in one place, in Go, in the team's language — not spread across FK declarations in a schema file invisible from the call site.
- **Consistency with ADR-009**, which already forbids business logic in SQL. Referential actions are behaviour.
- **Consistency with ADR-012**, which defines aggregate boundaries by consistency requirements rather than table structure. Deletion policy should follow the same boundaries.

The cost is real: repositories grow explicit child-deletion code, and adding a child table means remembering to extend it. RESTRICT makes forgetting loud, which is what makes the cost acceptable.

## Migration Plan

Staged, so each step is independently reviewable. Risk is lower than the cascade count suggests: almost nothing hard-deletes today, so the cascades are largely dormant.

1. **`player_match.club_match_id` → RESTRICT.** Highest value, smallest change. Converts the one live silent-data-loss risk into an error and encodes "submitted teams are not disposable" as an invariant. Worth shipping on its own.
2. **Audit `deleted_at` table by table.** Decide per entity: keep with a documented reason, or drop the column. Expect most to drop.
3. **Partial unique indexes** on whatever retains `deleted_at` after step 2.
4. **Remove remaining cascades** (both schemas) and add explicit child deletion to the repository methods that need it. Do `ffl` first, `afl` after, verifying each existing delete path.

Step 1 is a fast follow-up to the phase-25 fixture work. Steps 2–4 are their own slice and should not block it.

## Consequences

### Positive

- Deletion policy is readable in code and testable without a database.
- Mistakes fail loudly and locally instead of destroying data quietly.
- Removes a latent bug class (tombstone vs. unique constraint) in five remaining places.
- Shrinks the schema's implicit behaviour surface; DDL goes back to describing structure and invariants.

### Negative

- Repository methods get longer; child deletion is boilerplate that must be maintained as the schema grows.
- Forgetting a child yields a runtime FK error rather than a compile error. Mitigated by tests and by the error being immediate and specific.
- Dropping `deleted_at` from a table is irreversible for existing tombstoned rows — step 2 must confirm nothing reads them before dropping. (Nothing in the FFL read paths currently surfaces soft-deleted rows, but this needs verifying per table rather than assuming.)
- Multi-step migration touching two schemas; needs care to avoid a half-migrated state where some FKs cascade and some restrict.

## Open Questions

- Does any reporting, analytics, or historical-import path depend on reading soft-deleted rows? This gates step 2 and has not been verified beyond the FFL read paths.
- Should `round` retain soft delete? `player_season.from_round_id`/`to_round_id` reference it *without* cascade, so a round is already effectively RESTRICT-protected. It may be a genuine keep-with-reason case.
