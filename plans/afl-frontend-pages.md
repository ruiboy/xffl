# AFL Frontend — Page Inventory

## Routes

| Page   | Route                                          | Description                                |
| ------ | ---------------------------------------------- | ------------------------------------------ |
| Home   | `/`                                            | Latest season + round, ladder, matches     |
| Round  | `/afl/seasons/:seasonId/rounds/:roundId`       | Round matches, top players per stat        |
| Match  | `/afl/seasons/:seasonId/matches/:matchId`      | Teams, scores, full player stats           |
| Admin  | `/admin/afl/seasons/:seasonId/matches/:matchId`| Match stats editing                        |

## Pages

### Home (`/`)

Determines the current context automatically:
- **Latest season** = last season returned by `aflSeasons`
- **Latest round** = last round in that season

Displays:
- Season name and round name
- **Ladder** — full season standings
- **Round matches** — each match showing team names, scores, winner. Links to Match page.
- **Round navigation** — links to other rounds, each links to Round page

### Round (`/afl/seasons/:seasonId/rounds/:roundId`)

- Round name (e.g. "Round 14")
- **Round matches** — same presentation as Home. Links to Match page.
- **Top 5 players** in each stat category for the round:
  1. Kicks
  2. Handballs
  3. Marks
  4. Hitouts
  5. Tackles
  6. Goals
- **Round navigation** — links to other rounds

### Match (`/afl/seasons/:seasonId/matches/:matchId`)

- Team names, scores, winner
- Venue, start time
- Full player stats for both teams (read-only)
- Stat columns: K, HB, M, HO, T, G, B, Disposals, Score

### Admin Match (`/admin/afl/seasons/:seasonId/matches/:matchId`)

- Same layout as Match but with editable player stats
- Uses existing `UpdateAFLPlayerMatch` mutation

## Reusable Components

| Component    | Used on      | Description                                              |
| ------------ | ------------ | -------------------------------------------------------- |
| `NavBar`     | All pages    | Shared navigation bar                                    |
| `MatchSummary` | Home, Round | Team names, scores, winner indicator. Links to Match.  |
| `LadderTable`  | Home        | Season standings: pos, club, P, W, L, D, F, A, Pts    |
| `RoundNav`     | Home, Round | Links to all rounds in the season, highlights current  |
| `TopPlayers`   | Round       | Top 5 players in a single stat category                |
| `PlayerStatsTable` | Match, Admin | Full player stats table (read-only or editable)   |

## Navigation Bar

- **Home** — `/`
- **Seasons** — dropdown to switch seasons (future: links to season-specific Home)
- **Clubs** — placeholder for future club pages
- **About** — placeholder

## API Considerations

No new backend queries needed initially. The frontend can:
1. Fetch `aflSeasons` → pick the last one
2. Fetch that season's rounds → pick the last one
3. All match/round/ladder data is nested under the season query

If performance becomes an issue, a dedicated `latestAFLSeason` query could be added later.
