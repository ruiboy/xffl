package domain

const (
	PremiershipPointsWin  = 4
	PremiershipPointsDraw = 2
)

// ScoredClubMatch is one club's final club_match, tagged with its match's style
// and round type — the single input to the ladder. Versus, bye and superbye all
// arrive as these rows; the ladder groups them by match and applies the rules for
// that style. It carries the club_match id so per-club_match premiership points
// can be written back.
type ScoredClubMatch struct {
	MatchID      int
	ClubMatchID  int
	ClubSeasonID int
	Score        int
	Style        MatchStyle
	RoundType    RoundType
}

// CalculateLadder folds a season's final club_matches into per-ClubSeason
// standings. Grand-final rounds are excluded — the ladder is home-and-away only.
// By style:
//   - versus: the two sides give win/loss/draw, premiership points, and for/against;
//   - bye: the club's score adds to For only (no game played, no points);
//   - superbye: every score adds to For, and the top scorer(s) earn one extra point.
func CalculateLadder(results []ScoredClubMatch) map[int]ClubSeason {
	standings := make(map[int]ClubSeason)
	for _, group := range groupByMatch(results) {
		switch group[0].Style {
		case MatchStyleVersus:
			if len(group) != 2 {
				continue // a versus match counts only once both sides are final
			}
			foldVersus(standings, group[0], group[1])
		case MatchStyleBye:
			for _, e := range group {
				cs := standings[e.ClubSeasonID]
				cs.ID = e.ClubSeasonID
				cs.For += e.Score
				standings[e.ClubSeasonID] = cs
			}
		case MatchStyleSuperbye:
			foldSuperbye(standings, group)
		}
	}
	return standings
}

func foldVersus(standings map[int]ClubSeason, a, b ScoredClubMatch) {
	csA := standings[a.ClubSeasonID]
	csB := standings[b.ClubSeasonID]
	csA.ID, csB.ID = a.ClubSeasonID, b.ClubSeasonID
	csA.Played++
	csB.Played++
	csA.For += a.Score
	csA.Against += b.Score
	csB.For += b.Score
	csB.Against += a.Score
	switch {
	case a.Score > b.Score:
		csA.Won++
		csA.PremiershipPoints += PremiershipPointsWin
		csB.Lost++
	case b.Score > a.Score:
		csB.Won++
		csB.PremiershipPoints += PremiershipPointsWin
		csA.Lost++
	default:
		csA.Drawn++
		csB.Drawn++
		csA.PremiershipPoints += PremiershipPointsDraw
		csB.PremiershipPoints += PremiershipPointsDraw
	}
	standings[a.ClubSeasonID] = csA
	standings[b.ClubSeasonID] = csB
}

func foldSuperbye(standings map[int]ClubSeason, group []ScoredClubMatch) {
	top := topScore(group)
	for _, e := range group {
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

// ClubMatchPremiershipPoints returns the premiership points earned on each final
// club_match: a versus winner gets PremiershipPointsWin (both get PremiershipPointsDraw
// on a draw), a superbye's top scorer(s) get 1, and everything else 0.
func ClubMatchPremiershipPoints(results []ScoredClubMatch) map[int]int {
	points := make(map[int]int)
	for _, group := range groupByMatch(results) {
		switch group[0].Style {
		case MatchStyleVersus:
			if len(group) != 2 {
				continue
			}
			a, b := group[0], group[1]
			switch {
			case a.Score > b.Score:
				points[a.ClubMatchID], points[b.ClubMatchID] = PremiershipPointsWin, 0
			case b.Score > a.Score:
				points[a.ClubMatchID], points[b.ClubMatchID] = 0, PremiershipPointsWin
			default:
				points[a.ClubMatchID], points[b.ClubMatchID] = PremiershipPointsDraw, PremiershipPointsDraw
			}
		case MatchStyleSuperbye:
			top := topScore(group)
			for _, e := range group {
				if top > 0 && e.Score == top {
					points[e.ClubMatchID] = 1
				} else {
					points[e.ClubMatchID] = 0
				}
			}
		case MatchStyleBye:
			for _, e := range group {
				points[e.ClubMatchID] = 0
			}
		}
	}
	return points
}

func topScore(group []ScoredClubMatch) int {
	top := 0
	for _, e := range group {
		if e.Score > top {
			top = e.Score
		}
	}
	return top
}

// groupByMatch groups scored club_matches by their match, skipping grand finals
// (excluded from the home-and-away ladder). Match order is preserved.
func groupByMatch(results []ScoredClubMatch) [][]ScoredClubMatch {
	index := make(map[int]int) // matchID → position in out
	var out [][]ScoredClubMatch
	for _, r := range results {
		if r.RoundType.IsFinal() {
			continue
		}
		if i, ok := index[r.MatchID]; ok {
			out[i] = append(out[i], r)
			continue
		}
		index[r.MatchID] = len(out)
		out = append(out, []ScoredClubMatch{r})
	}
	return out
}
