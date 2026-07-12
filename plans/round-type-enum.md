# Round Type — typed finals classification

Referenced from `ideas.md` REV-11. Replace name-based finals detection with a typed
`round_type` on `afl.round` and `ffl.round`. Closes the ladder finals-leak bug and is
the foundation for per-week finals navigation.

## Why

- Finals are currently detected by string-matching the round name (`%Final%`), which is brittle.
- **Bug (REV-11):** `domain.CalculateLadder` / `FindFinalBySeasonID` filter on `data_status='final'` (completeness), *not* round type, so finals rounds count toward the home-and-away ladder. Latent until the Phase 24 backfill flipped 2024/2025 finals to `final`; the derivation SQL excludes them via `NOT ILIKE '%Final%'`, so app and SQL now disagree.
- Blocks per-week finals navigation in `RoundNav`.

## Model

- `round_type` lives on `round`, not `match` — rounds are homogeneous (every match in a round shares a type).
- Stored as `VARCHAR(50) NOT NULL DEFAULT 'MINOR'`, mirroring `match.data_status` / `club_match.side`. Validity + an `IsFinal()` helper live in the domain layer; **no DB CHECK** (project rule — enum logic in domain).
- Values (AFL): `MINOR`, `WILDCARD_FINAL`, `QUALIFYING_FINAL`, `ELIMINATION_FINAL`, `SEMI_FINAL`, `PRELIMINARY_FINAL`, `GRAND_FINAL`. `IsFinal()` = any of the finals values. "Opening Round" is `MINOR`.
- **FFL mirror:** its own enum (service isolation — no cross-service import), but only **two values** — `MINOR` and `GRAND_FINAL`. FFL models no finals lead-up rounds by type; everything but the grand final is `MINOR` (including the SUPERBYE round), and `IsFinal()` on the FFL side is just `== GRAND_FINAL`. Set it independently at FFL season-setup; do **not** couple it to the linked `afl_round_id`'s type at query time.

## Finals week is derived, not stored

Fixed canonical order of the finals tiers:

```
0  WILDCARD_FINAL
1  QUALIFYING_FINAL / ELIMINATION_FINAL   (same tier → same week)
2  SEMI_FINAL
3  PRELIMINARY_FINAL
4  GRAND_FINAL
```

`finalsWeek` = dense-rank of the distinct tiers *present in that season*, up to this round's tier. QF/EF share tier 1 → collapse into one week. Ranking (not a static `type → week` map) is required so week numbers self-correct when a wildcard round exists:

- Normal year — tiers `{1,2,3,4}` → QF/EF = **W1**, SF = W2, PF = W3, GF = W4.
- Wildcard year — tiers `{0,1,2,3,4}` → Wildcard = **W1**, QF/EF = **W2**, SF = W3, PF = W4, GF = W5.

The season's own set of finals rounds carries the "is there a wildcard" signal, so a new format adds one enum value + one ordering entry — no renumbering, no stored-data churn. Week **labels** also derive from the tiers present (tier-0 → "Wildcard", GF → "Grand Final", else "Finals Week N").

## Schema migration

Current DDL (both are identical apart from FFL's `afl_round_id`):

```
afl.round (… season_id, name)
ffl.round (… season_id, name, afl_round_id)
```

### Live DB — DDL + backfill

```sql
ALTER TABLE afl.round ADD COLUMN round_type VARCHAR(50) NOT NULL DEFAULT 'MINOR';
ALTER TABLE ffl.round ADD COLUMN round_type VARCHAR(50) NOT NULL DEFAULT 'MINOR';

-- Backfill from round names (the last time a name is parsed). The live DB carries
-- finals rounds (1998–2023, 2024/2025) that the seed files do not.
UPDATE afl.round SET round_type = CASE
    WHEN name ILIKE 'Grand Final%'       THEN 'GRAND_FINAL'
    WHEN name ILIKE 'Preliminary Final%' THEN 'PRELIMINARY_FINAL'
    WHEN name ILIKE 'Semi Final%'        THEN 'SEMI_FINAL'
    WHEN name ILIKE 'Elimination Final%' THEN 'ELIMINATION_FINAL'
    WHEN name ILIKE 'Qualifying Final%'  THEN 'QUALIFYING_FINAL'
    WHEN name ILIKE 'Wildcard%'          THEN 'WILDCARD_FINAL'
    ELSE 'MINOR'
END;
-- FFL: only two values, so the CASE is just
--   WHEN name ILIKE 'Grand Final%' THEN 'GRAND_FINAL' ELSE 'MINOR'.
-- Live/seed FFL rounds are MINOR + SUPERBYE today, so all resolve to MINOR until
-- an FFL grand final is added.
```

### `dev/postgres/init` (fresh-DB schema)

- `01_afl_schema.sql` — add `round_type VARCHAR(50) NOT NULL DEFAULT 'MINOR'` to `afl.round` (after `name`).
- `02_ffl_schema.sql` — same on `ffl.round` (after `afl_round_id`).

### `dev/postgres/seed` + `dev/postgres/test-e2e` (fixtures)

- Seeds are **home-and-away only** today — AFL round names are `Round N` / `Opening Round`; FFL adds `SUPERBYE`; no finals rounds exist in any seed file. So the column `DEFAULT` already types every seeded round correctly; the `INSERT`s need no per-row change.
- Add the name→type `UPDATE` (same CASE as above) at the end of the AFL and FFL seed runs (e.g. `06_seed_finalize.sql`) so any *future* finals rounds added to a seed get typed automatically. Mirror into the e2e seeds (`test-e2e/10_afl_seed.sql`, `11_ffl_seed.sql`) for parity.
- **Divergence note:** the live DB has finals rounds the seed files don't. The seed change is cosmetic (default covers it); the live `ALTER` backfill is where finals actually get typed.

## Domain + application

- `RoundType` enum + `IsFinal()` in the AFL domain, and a separate one in the FFL domain.
- Regenerate sqlc + gqlgen to surface `round_type`.
- Swap `CalculateLadder` / `FindFinalBySeasonID` to filter on `NOT IsFinal()`; delete the `%Final%` matching → **closes REV-11**.
- `finalsWeek` derivation (dense-rank helper) — domain method, or a shared frontend util.

## Frontend

- Expose `roundType` on the round GraphQL type; add to fragments.
- `RoundNav` (`features/ffl/components/RoundNav.vue`): **per-week** finals nav. It already receives the full season round list, so it (or the shared `finalsWeek` util) orders the finals rounds present and dense-ranks them into weeks. A week can span multiple rounds (Week 1 = Qualifying + Elimination are separate DB rounds), so a week pill needs a **week-scoped destination** (a "show the whole finals week" view) rather than a single `roundId`. H&A pills stay per-round numbers; the current `name.replace(/^Round\s+/i,'')` label (which overflows the 32px pill on finals/Opening names) is replaced by an enum-keyed short label.

## Suggested rollout (scope TBD)

1. **Bug-fix slice (do-now candidate):** DDL + init + seed + domain `RoundType`/`IsFinal()` + swap the ladder filter off `%Final%`. Low risk, closes REV-11.
2. **Nav slice (deferrable):** `finalsWeek` derivation + `RoundNav` per-week grouping + the finals-week view.
3. **FFL mirror:** follows AFL; lands with FFL finals / super-bye work.

## Open decisions

- `finalsWeek` derived client-side vs exposed as a server-computed round field.
- No ADR — schema+domain change following the existing typed-varchar precedent.
