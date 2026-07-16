package domain

const (
	PremiershipPointsWin  = 4
	PremiershipPointsDraw = 2
)

// ByeResult is a club's scoring bye in a home-and-away round. Unlike an AFL bye,
// the club still fields a team, so its score counts toward the season aggregate —
// but a bye has no opponent, so it yields no win/loss and no premiership points.
type ByeResult struct {
	ClubSeasonID int
	Score        int
	RoundType    RoundType
}

// CalculateLadder folds a season's final matches and scoring byes into
// per-ClubSeason standings. Matches must have StoredScore set on each ClubMatch;
// matches with a missing ClubSeasonID on either side are skipped, as is the grand
// final (the ladder is home-and-away only). A bye adds its score to For and counts
// as a played round, but earns no premiership points.
func CalculateLadder(matches []Match, byes []ByeResult) map[int]ClubSeason {
	standings := make(map[int]ClubSeason)
	for _, m := range matches {
		if m.RoundType.IsFinal() {
			continue
		}
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
	for _, b := range byes {
		if b.RoundType.IsFinal() || b.ClubSeasonID == 0 {
			continue
		}
		cs := standings[b.ClubSeasonID]
		cs.ID = b.ClubSeasonID
		cs.Played++
		cs.For += b.Score
		standings[b.ClubSeasonID] = cs
	}
	return standings
}
