# Data Operations Workflow

Repeating operational cycle for each FFL season. Steps run broadly in sequence; later steps
depend on earlier ones being complete.

> This workflow is the source of truth for the Data Ops section of the UI. Future seasons
> follow the same steps.

---

## Season setup *(once per season)*

### Step 1 — AFL season player import

Import the AFL player roster for the new season, matching to existing players where possible.

- **When**: before the season starts, once player registrations are announced
- **Input**: AFL registration data (source TBD; afltables or equivalent)
- **Output**: `afl.player` + `afl.player_season` rows for the new season
- **Notes**: fuzzy name matching to existing players; new and retiring players flagged for accept/reject review

### Step 2 — FFL squad import

Resolve each FFL club's squad to AFL player IDs for the new season.

- **When**: after Step 1; before Round 1
- **Precondition**: AFL player roster complete
- **Input**: FFL club squad lists
- **Output**: `ffl.player` + `ffl.player_season` rows linked to `afl.player_season`

---

## Ad-hoc *(during season)*

### Step 3 — In-season player trades

Add or remove a player from an FFL club's squad mid-season.

- **When**: whenever a trade occurs (delistings, mid-season signings, rule changes)
- **Entry point**: Squad view → Manage mode
- **Output**: `ffl.player_season` row updated with `from_round_id` / `to_round_id`

---

## Round cycle *(every round)*

### Step 4 — FFL team submission

Import each FFL club's team for the round from the forum post.

- **When**: after team sheets close, before AFL matches begin
- **Input**: forum post with player names and positions (pasted into Data Ops UI)
- **Output**: `ffl.club_match` + `ffl.player_match` rows; `ffl.club_match.data_status → submitted`
- **Notes**: unrecognised player names surface for manual resolution via player search

### Step 5 — AFL stats import

Scrape and import AFL player match statistics for the round.

- **When**: as AFL matches complete; repeated until all round matches are marked final
- **Input**: FootyWire match page (scraped via Data Ops UI)
- **Output**: `afl.player_match` rows with stats; `afl.match.data_status → partial` then `final`
- **Notes**: previously-resolved name mismatches auto-apply via player source map

### Step 6 — Score reconciliation

Compare submitted FFL team scores against AFL-derived scores; surface and document discrepancies.

- **When**: after Steps 4 and 5 are both final for the match
- **Precondition**: `afl.match.data_status = final` AND `ffl.club_match.data_status = final`
- **Input**: finalised AFL stats + confirmed FFL team
- **Output**: structured diff; copy-pasteable forum summary for Team Managers
