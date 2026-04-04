# Current Sprint

**Sprint goal:** Phase 9 — FFL Team Composition Rules

Define and enforce the rules for how an FFL team is structured each round.
Validation lives in the domain. The UI enforces the rules through good UX.

---

## Team Composition Rules (source of truth)

| Position  | Starters | Scoring |
|-----------|----------|---------|
| Goals     | 3        | 5 pts × goals |
| Kicks     | 4        | 1 pt × kicks |
| Handballs | 4        | 1 pt × handballs |
| Marks     | 2        | 2 pts × marks |
| Tackles   | 2        | 4 pts × tackles |
| Hitouts   | 2        | 1 pt × hitouts |
| Star      | 1        | all stats combined |
| **Total** | **18**   | |

**Bench** — up to 4 players:
- 1 backup star (no position pair; backs up the star slot)
- 3 dual-position players: each covers exactly 2 non-star positions; each non-star position may be covered by at most one bench player

**Interchange** — at most 1 position per match: the bench player covering that position may freely substitute the starter if they score higher.

**Substitution** — a bench player steps in only when the starter is DNP (did not play). A player who played but scored 0 cannot be substituted.

Teams need not be full.

---

## Backend Tasks

### B1 — Fix Score() for multi-starter positions
**File:** `services/ffl/internal/domain/club_match.go`

`starters` is currently `map[Position]*PlayerMatch` — only one slot per position, last write wins. Change to `map[Position][]*PlayerMatch` so all starters score. Substitution and interchange logic applies per individual slot within the position group.

- [ ] Change starters map type
- [ ] Update substitution loop (iterate each slot per position)
- [ ] Update interchange loop
- [ ] Update total sum

### B2 — Add PositionSlots constants + ValidateLineup()
**File:** `services/ffl/internal/domain/player_match.go`

- [ ] Add `PositionSlots map[Position]int` (goals:3, kicks:4, handballs:4, marks:2, tackles:2, hitouts:2, star:1)
- [ ] Add `ValidateLineup(entries []UpsertPlayerMatchParams) error` enforcing:
  1. Starter count per position ≤ PositionSlots[pos]
  2. Bench count ≤ 4
  3. At most 1 bench star (backupPositions contains "star")
  4. Non-star bench players have exactly 2 backup positions, none "star"
  5. Each non-star position in at most 1 bench player's backup pair
  6. At most 1 interchange position across all players
  7. Interchange position (if set) must be a recognised Position value

### B3 — Enforce validation in SetLineup command
**File:** `services/ffl/internal/application/commands.go`

- [ ] Call `domain.ValidateLineup(entries)` before opening the transaction; return error on failure

### B4 — Integration tests for all composition rules
**File:** `services/ffl/internal/interface/graphql/integration_test.go`

- [ ] Valid 18-starter lineup saves successfully
- [ ] Too many starters for a position → error
- [ ] 5 bench players → error
- [ ] 2 bench stars → error
- [ ] Bench player with 3 backup positions → error
- [ ] Same position in two bench players' backup pairs → error
- [ ] 2 interchange positions → error
- [ ] Score() sums correctly across all starters in a multi-slot position

---

## Frontend Tasks

### F1 — Fix slot counts + state retention bug
**File:** `frontend/web/src/features/ffl/views/TeamBuilderView.vue`

Current counts are wrong (marks:3, tackles:3, star:3 = 22 total; should be 18).

State retention bug: `watch(clubMatch, …)` resets all local slot state whenever Apollo updates the cache (e.g. after a mutation response). Fix: track `initializedMatchId`; only re-initialise slots when the match ID changes, not when cached data within the same match updates.

- [ ] Fix positions array: marks→2, tackles→2, star→1
- [ ] Add `initializedMatchId = ref<string|null>(null)`; guard the watch with `if (cm.id === initializedMatchId.value) return`
- [ ] Update summary: "18 starters" not "22"

### F2 — Redesign bench UI
**File:** `frontend/web/src/features/ffl/views/TeamBuilderView.vue`

Replace flat 8-slot bench with typed structure:
- `benchStarSlot: { player, interchangeable: boolean }`
- `benchDualSlots: [{ player, positions: [pos|null, pos|null] }] × 3`

UI:
- **Backup Star** row: single slot; "Interchange" toggle (makes star the interchange position)
- **Bench 1/2/3** rows: player name; two position pill selectors (non-star only; each position disabled if already used by another bench dual slot); "Interchange" toggle; Remove
- Only one interchange active across all rows at a time

Submit serialisation:
- Backup star → `{ position: "star", backupPositions: "star", interchangePosition: (if interchange) "star" }`
- Dual bench → `{ backupPositions: "kicks,marks", interchangePosition: (if interchange) "kicks" }`

- [ ] Replace benchSlots ref with benchStarSlot + benchDualSlots
- [ ] Render Backup Star row with interchange toggle
- [ ] Render Bench 1/2/3 rows with dual position selectors and interchange toggle
- [ ] Disable position pills already used across other bench dual slots
- [ ] Allow at most 1 interchange active across all rows
- [ ] Load existing lineup data into new structure in the watch initialiser
- [ ] Update submitLineup() to serialise correctly

### F3 — Squad panel bench actions
**File:** `frontend/web/src/features/ffl/views/TeamBuilderView.vue`

Replace single "B" button with:
- `★` — add to backup star slot (disabled if filled)
- `B` — add to next available dual-position bench slot (disabled if all 3 filled)

- [ ] Replace B button with ★ and B actions
- [ ] Disable ★ if benchStarSlot.player is set
- [ ] Disable B if all 3 benchDualSlots are filled

### F4 — E2E tests for Team Builder
**File:** `frontend/web/e2e/ffl-team-builder.spec.ts`

#### Layout and structure (read-only mode)
- [ ] Club name shown as h1 heading
- [ ] Manage button visible; Done button not visible
- [ ] Position group headings present: Goals, Kicks, Handballs, Marks, Tackles, Hitouts, Star
- [ ] Correct slot counts per position (3/4/4/2/2/2/1 = 18 total)
- [ ] Bench section present with Backup Star row and 3 dual-position rows
- [ ] No Remove buttons, no Squad panel, no position selectors in read-only mode

#### Layout and structure (manage mode)
- [ ] Click Manage → Done button visible, Manage button gone
- [ ] Squad panel visible in manage mode
- [ ] Remove buttons visible on filled slots
- [ ] Interchange toggles visible on bench rows
- [ ] Dual-position selectors visible on bench dual-position rows
- [ ] Click Done → returns to read-only (Manage visible, Done gone, Squad panel gone)

#### Empty team flow
- [ ] Fresh team: all slots show "Empty slot" in read-only mode
- [ ] Enter Manage → all slots still empty, full squad available in panel
- [ ] Done on empty team → returns to read-only with all slots still empty (no crash)

#### Partial edit → view → re-edit (state retention)
- [ ] Enter Manage → add one player to Goals → click Done
- [ ] In read-only mode: that player's name is visible in the Goals section
- [ ] Click Manage again → player still in Goals slot (local state retained, not reset by Apollo)
- [ ] Add a second player to Kicks → click Done
- [ ] Both players visible in read-only mode

#### Navigate away and back (server state persisted)
- [ ] Add a player to Goals → Done (saves to server)
- [ ] Navigate away (e.g. click Squad in nav)
- [ ] Navigate back to Team Builder
- [ ] Player still visible in Goals slot in read-only mode (loaded from server)
- [ ] Enter Manage → player still in slot

#### Continue building across multiple sessions
- [ ] Load page with partially-saved lineup → enter Manage → existing players present
- [ ] Add more players → Done → all players (old + new) visible in read-only mode
- [ ] Enter Manage again → all players still present in correct slots

#### Bench: backup star
- [ ] In manage mode: Squad panel shows ★ button per player
- [ ] Click ★ on a player → appears in Backup Star row
- [ ] ★ button disabled for other players once Backup Star is filled
- [ ] Remove from Backup Star → ★ buttons re-enabled

#### Bench: dual-position
- [ ] Squad panel shows B button per player
- [ ] Click B → player appears in first empty dual-position bench row
- [ ] Dual-position bench row shows two position selectors (empty initially)
- [ ] Select position in first selector → that position disabled in other bench rows' selectors
- [ ] B button disabled once all 3 dual-position slots are filled

#### Interchange
- [ ] Interchange toggle visible on Backup Star row and each dual-position bench row
- [ ] Click interchange toggle on one row → that row marked as interchange
- [ ] Clicking interchange on a second row → first deactivates, second activates (only 1 active)
- [ ] Interchange state preserved through Done → re-open Manage

---

## Execution Order

```
B2 → B1 → B3 → B4    backend: domain rules first, then enforce, then test
F1 → F2 → F3 → F4    frontend: fix basics first, then rebuild bench, then test
```

B2 can start immediately. F1 can start in parallel with B2 (slot counts are known).
F2 depends on F1. B4 depends on B1–B3.
