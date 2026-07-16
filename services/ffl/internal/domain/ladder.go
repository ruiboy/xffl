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

// SuperbyeEntry is one club's team in a superbye — its score for that round.
type SuperbyeEntry struct {
	ClubSeasonID int
	Score        int
}

// SuperbyeClubMatch is one loaded superbye entry with its club_match id, used by
// the ladder recompute to fold standings and to award the winner's premiership
// point on the club_match itself.
type SuperbyeClubMatch struct {
	MatchID      int
	ClubMatchID  int
	ClubSeasonID int
	Score        int
	RoundType    RoundType
}

// SuperbyeMatch is a round's superbye: every club fields a team, no head-to-head.
// Like a bye each score counts toward For (no game played), and additionally the
// round's top scorer(s) earn one extra premiership point.
type SuperbyeMatch struct {
	RoundType RoundType
	Entries   []SuperbyeEntry
}

// CalculateLadder folds a season's final matches, scoring byes and superbyes into
// per-ClubSeason standings. Matches must have StoredScore set on each ClubMatch;
// matches with a missing ClubSeasonID on either side are skipped, as is the grand
// final (the ladder is home-and-away only). A bye adds its score to For but does
// not count as a played round and earns no premiership points. A superbye likewise
// adds each score to For (no played round); its top scorer(s) also earn one extra
// point, tracked in ExtraPoints and included in PremiershipPoints.
func CalculateLadder(matches []Match, byes []ByeResult, superbyes []SuperbyeMatch) map[int]ClubSeason {
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
		cs.For += b.Score
		standings[b.ClubSeasonID] = cs
	}
	for _, sb := range superbyes {
		if sb.RoundType.IsFinal() {
			continue
		}
		top := 0
		for _, e := range sb.Entries {
			if e.Score > top {
				top = e.Score
			}
		}
		for _, e := range sb.Entries {
			if e.ClubSeasonID == 0 {
				continue
			}
			cs := standings[e.ClubSeasonID]
			cs.ID = e.ClubSeasonID
			cs.For += e.Score
			if top > 0 && e.Score == top {
				cs.ExtraPoints++
				cs.PremiershipPoints++
			}
			standings[e.ClubSeasonID] = cs
		}
	}
	return standings
}
