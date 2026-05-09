package domain

const (
	PremiershipPointsWin  = 4
	PremiershipPointsDraw = 2
)

// CalculateLadder folds a set of final matches into per-ClubSeason standings.
// Matches must have StoredScore set on each ClubMatch; matches with a missing
// ClubSeasonID on either side are skipped.
func CalculateLadder(matches []Match) map[int]ClubSeason {
	standings := make(map[int]ClubSeason)
	for _, m := range matches {
		if m.Home.ClubSeasonID == 0 || m.Away.ClubSeasonID == 0 {
			continue
		}
		home := standings[m.Home.ClubSeasonID]
		away := standings[m.Away.ClubSeasonID]
		home.ID = m.Home.ClubSeasonID
		away.ID = m.Away.ClubSeasonID

		home.Played++
		away.Played++
		home.For += m.Home.StoredScore
		home.Against += m.Away.StoredScore
		away.For += m.Away.StoredScore
		away.Against += m.Home.StoredScore

		switch m.DeriveResult() {
		case MatchResultHomeWin:
			home.Won++
			home.PremiershipPoints += PremiershipPointsWin
			away.Lost++
		case MatchResultAwayWin:
			away.Won++
			away.PremiershipPoints += PremiershipPointsWin
			home.Lost++
		case MatchResultDraw:
			home.Drawn++
			home.PremiershipPoints += PremiershipPointsDraw
			away.Drawn++
			away.PremiershipPoints += PremiershipPointsDraw
		}

		standings[m.Home.ClubSeasonID] = home
		standings[m.Away.ClubSeasonID] = away
	}
	return standings
}
