# Current Sprint

**Sprint goal:** Phase 7 — Data Model Refinements

Clean up AFL/FFL data models so relationships are correct, then propagate changes through domain entities, GraphQL schemas, frontend, seed data, and tests.

## Key decisions

- `afl.player.club_id` removed — players exist independently; attached to clubs via `player_season → club_season`
- `ffl.player.club_id` removed — same logic
- `afl.player_season` gains `from_round_id` / `to_round_id` (nullable; null = start/end of season)
- `afl.player_match` gains `status` (named/played/dnp)
- `ffl.player.name` → `drv_name` (every FFL player has an AFL player; name is derived)
- `ffl.player_match.score` → `drv_score`
- `ffl.player_match.status` stays underived (may be initialised from AFL, but takes its own values)
- `ffl.club_match` gets unique constraint `(club_season_id, match_id)`
- `ffl.match`: `home_club_match_id`/`away_club_match_id` nullable — versus: both set, bye: home only, else: `clubs` JSONB used
- `ffl.match.clubs` JSONB stores club_season_ids: `{"A": {"club_season_id": 2}, "B": {"club_season_id": 1}}`
- `AFLClub.players` GraphQL field dropped entirely
- No migration scripts — just update init SQL + seed data

## Tasks

### 1. AFL schema changes
- [ ] Remove `club_id` from `afl.player`
- [ ] Add `from_round_id` and `to_round_id` to `afl.player_season` (nullable FK → `afl.round`)
- [ ] Add `status` to `afl.player_match` (varchar)
- [ ] Drop index `idx_afl_player_club_id`
- [ ] Add indexes for new columns

### 2. FFL schema changes
- [ ] Remove `club_id` from `ffl.player`
- [ ] Rename `ffl.player.name` → `drv_name`
- [ ] Rename `ffl.player_match.score` → `drv_score`
- [ ] Add unique constraint `(club_season_id, match_id)` on `ffl.club_match`
- [ ] Drop index `idx_player_club_id`

### 3. Seed data
- [ ] Update AFL seed to remove player `club_id` values, add `status` to player_match rows
- [ ] Update FFL seed to match renamed columns

### 4. AFL domain + infrastructure
- [ ] Remove `ClubID` from `Player` entity
- [ ] Add `FromRoundID` / `ToRoundID` to `PlayerSeason` entity
- [ ] Add `Status` to `PlayerMatch` entity
- [ ] Update sqlc queries
- [ ] Update repository implementations

### 5. FFL domain + infrastructure
- [ ] Remove `ClubID` from `Player` entity (if present)
- [ ] Keep `Name` on `Player` entity (DB column is `drv_name` but domain field stays `Name`)
- [ ] Keep `Score` on `PlayerMatch` entity (DB column is `drv_score` but domain field stays `Score`)
- [ ] Update sqlc queries
- [ ] Update repository implementations

### 6. AFL GraphQL
- [ ] Drop `AFLClub.players` field and resolver
- [ ] Add `status` to `AFLPlayerMatch` type
- [ ] Update `updateAFLPlayerMatch` mutation input to include status
- [ ] Regenerate + fix resolvers

### 7. FFL GraphQL
- [ ] Update types if any field names changed (DB renames are transparent to GraphQL)
- [ ] Regenerate + fix resolvers

### 8. Frontend
- [ ] Remove any references to AFL player → club direct relationship
- [ ] Update queries/components if field names changed
- [ ] Verify all views still render correctly

### 9. Tests
- [ ] Update AFL integration tests (remove club_id from player fixtures, add status)
- [ ] Update FFL integration tests (renamed columns)
- [ ] Run full test suite green
- [ ] Run Playwright e2e tests green

## Out of scope
- FFL UX changes (nav bar, pages) — Phase 8
- Team composition rules — Phase 9
- Event integration (AFL→FFL) — Phase 11
