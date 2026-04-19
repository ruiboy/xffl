# Historical Data Import

Historical data imports are one-time (or rare) migration tools that load large datasets from external sources into the domain. This document covers the pattern, tooling, and identity reconciliation conventions shared across all historical imports — AFL and FFL.

## Key differences from production integrations

| Concern | Production adapter | Historical import |
|---|---|---|
| Entry point | `cmd/ingest/main.go` | `dev/import/<source>/main.go` |
| Run frequency | Recurring (scheduled) | Once; archived or kept for idempotent re-runs |
| Fetch policy | Cache-aware; respect host rate limits | Fetch once per run; no caching needed |
| Events | Fires domain events downstream | No events — bulk load only |
| Tests | Unit + integration | Manual verification against DB |

## The identity problem

Every external source uses its own identifiers and name spellings. The same player (e.g. Patrick Dangerfield, career 2007–present) may appear under different names or IDs across sources and across years. Without a durable mapping, re-importing or adding a second source creates duplicate players.

The solution is an **xref table per source per entity**, exactly as described in `integrations.md`:

```sql
afl.xref_<source>_player   -- external player identifier → afl.player.id
```

The xref is populated during import and persists permanently. On subsequent runs from the same source, already-mapped players are resolved instantly via xref — no re-matching needed.

## Two-phase import

Every historical import runs in two phases.

### Phase 1 — Reconcile (`--reconcile` flag)

Fetch all data. For each entity seen (player, etc.) that is not yet in the xref:

1. **Exact match** against the domain table (`afl.player.name`) → auto-accept; write xref row
2. **Fuzzy match** (Levenshtein, see below) → auto-accept above threshold; flag below threshold
3. **No viable match** → flag as likely-new entity

Write results to `dev/import/<source>/reconcile.csv`:

```
external_id, source_name, candidate_id, candidate_name, similarity, resolution
```

The `resolution` column is blank — **the human fills it in**:
- A domain `player_id` to link to an existing player
- `new` to create a new player from `source_name`

### Phase 2 — Import (default run)

Read `reconcile.csv` (must be fully resolved — no blank `resolution` entries):

1. For `new` rows: create the domain entity; record its ID back into the mapping
2. Populate xref rows (idempotent — `ON CONFLICT DO NOTHING`)
3. Upsert all data rows using xref-resolved IDs
4. Log counts and any skipped rows

**`reconcile.csv` is committed to the repo.** It is the permanent audit trail of every identity decision made at import time and ensures the import is fully reproducible.

## Fuzzy matching strategy

Use Levenshtein edit distance, normalised by the longer string length:

```
similarity = 1.0 − (edit_distance / max(len(a), len(b)))
```

| Similarity | Behaviour |
|---|---|
| ≥ 0.90 | Auto-accept; log match for review |
| 0.60–0.89 | Write top-3 candidates to `reconcile.csv`; human chooses |
| < 0.60 | Flag as `new` candidate; human confirms or corrects |

Common edge cases to handle correctly:
- Shortened given names: "Nick" vs "Nic", "Tom" vs "Thomas"
- Maiden vs married names
- Apostrophe variants: `O'Brien` vs `O'Brien` (typographic vs ASCII)
- Initials: "B. Smith" vs "Brad Smith"
- Name changes (gender transition, legal change)

When in doubt, prefer flagging over auto-accepting — a human confirmation is cheap; a wrong merge is hard to undo.

## Source conventions

Each import source gets the following layout:

| Artifact | Path |
|---|---|
| Import entry point | `dev/import/<source>/main.go` |
| Reconciliation artifact | `dev/import/<source>/reconcile.csv` |
| xref table SQL | `dev/postgres/init/<n>_<service>_integrations.sql` |
| Parser / fetcher (if reusable across runs) | `services/<svc>/internal/infrastructure/<source>/` |

If the parser is unlikely to be reused (truly one-off scrape), keep it in `dev/import/<source>/` rather than polluting the service infrastructure.

## No events on import

Historical imports do not fire domain events. Events signal live changes for downstream services to react to (e.g. FFL recalculating a fantasy score). Bulk historical loads bypass this path entirely — downstream data is seeded separately when needed.

## FFL historical data — the same pattern

FFL historical match imports (team submissions, scores, player season records) follow the exact same two-phase reconciliation approach. Additional consideration: FFL player seasons must eventually be **linked to AFL player seasons** so that fantasy scoring can reference factual AFL stats. This linkage is itself a reconciliation step — FFL player names against AFL player names — run after both AFL and FFL data are loaded.

See `plans/roadmap.md` for the phased plan covering all historical sources.
