package graphql

import (
	"strconv"

	"xffl/services/afl/internal/domain"
)

func toID(id int) string {
	return strconv.Itoa(id)
}

func fromID(id string) (int, error) {
	return strconv.Atoi(id)
}

func toStringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func convertClub(c domain.Club) *AFLClub {
	return &AFLClub{
		ID:   toID(c.ID),
		Name: c.Name,
	}
}

func convertClubs(clubs []domain.Club) []*AFLClub {
	out := make([]*AFLClub, len(clubs))
	for i, c := range clubs {
		out[i] = convertClub(c)
	}
	return out
}

func convertSeason(s domain.Season) *AFLSeason {
	return &AFLSeason{
		ID:   toID(s.ID),
		Name: s.Name,
	}
}

func convertSeasons(seasons []domain.Season) []*AFLSeason {
	out := make([]*AFLSeason, len(seasons))
	for i, s := range seasons {
		out[i] = convertSeason(s)
	}
	return out
}

func convertRound(r domain.Round) *AFLRound {
	return &AFLRound{
		ID:   toID(r.ID),
		Name: r.Name,
	}
}

func convertRounds(rounds []domain.Round) []*AFLRound {
	out := make([]*AFLRound, len(rounds))
	for i, r := range rounds {
		out[i] = convertRound(r)
	}
	return out
}

func convertMatch(m domain.Match) *AFLMatch {
	match := &AFLMatch{
		ID:    toID(m.ID),
		Venue: toStringPtr(m.Venue),
	}
	if !m.StartTime.IsZero() {
		t := m.StartTime.Format("2006-01-02T15:04:05Z")
		match.StartTime = &t
	}
	if m.Result != "" {
		r := string(m.Result)
		match.Result = &r
	}
	return match
}

func convertMatches(matches []domain.Match) []*AFLMatch {
	out := make([]*AFLMatch, len(matches))
	for i, m := range matches {
		out[i] = convertMatch(m)
	}
	return out
}

func convertClubSeason(cs domain.ClubSeason, club domain.Club) *AFLClubSeason {
	return &AFLClubSeason{
		ID:                toID(cs.ID),
		Club:              convertClub(club),
		Played:            cs.Played,
		Won:               cs.Won,
		Lost:              cs.Lost,
		Drawn:             cs.Drawn,
		For:               cs.For,
		Against:           cs.Against,
		PremiershipPoints: cs.PremiershipPoints,
	}
}

func convertClubMatch(cm domain.ClubMatch, club domain.Club) *AFLClubMatch {
	return &AFLClubMatch{
		ID:            toID(cm.ID),
		Club:          convertClub(club),
		RushedBehinds: cm.RushedBehinds,
		Score:         cm.Score,
	}
}

func convertPlayerMatch(pm domain.PlayerMatch, player domain.Player) *AFLPlayerMatch {
	return &AFLPlayerMatch{
		ID:        toID(pm.ID),
		Player:    convertPlayer(player),
		Kicks:     pm.Kicks,
		Handballs: pm.Handballs,
		Marks:     pm.Marks,
		Hitouts:   pm.Hitouts,
		Tackles:   pm.Tackles,
		Goals:     pm.Goals,
		Behinds:   pm.Behinds,
		Disposals: pm.Disposals(),
		Score:     pm.Score(),
	}
}

func convertPlayer(p domain.Player) *AFLPlayer {
	return &AFLPlayer{
		ID:   toID(p.ID),
		Name: p.Name,
	}
}

func convertPlayers(players []domain.Player) []*AFLPlayer {
	out := make([]*AFLPlayer, len(players))
	for i, p := range players {
		out[i] = convertPlayer(p)
	}
	return out
}
