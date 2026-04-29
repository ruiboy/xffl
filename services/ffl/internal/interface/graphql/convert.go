package graphql

import (
	"strconv"

	"xffl/services/ffl/internal/domain"
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

func convertClub(c domain.Club) *FFLClub {
	return &FFLClub{ID: toID(c.ID), Name: c.Name}
}

func convertClubs(clubs []domain.Club) []*FFLClub {
	out := make([]*FFLClub, len(clubs))
	for i, c := range clubs {
		out[i] = convertClub(c)
	}
	return out
}

func convertPlayer(p domain.Player) *FFLPlayer {
	return &FFLPlayer{ID: toID(p.ID), Name: p.Name, AflPlayerID: toID(p.AFLPlayerID)}
}

func convertPlayers(players []domain.Player) []*FFLPlayer {
	out := make([]*FFLPlayer, len(players))
	for i, p := range players {
		out[i] = convertPlayer(p)
	}
	return out
}

func convertSeason(s domain.Season) *FFLSeason {
	return &FFLSeason{ID: toID(s.ID), Name: s.Name}
}

func convertSeasons(seasons []domain.Season) []*FFLSeason {
	out := make([]*FFLSeason, len(seasons))
	for i, s := range seasons {
		out[i] = convertSeason(s)
	}
	return out
}

func convertRound(r domain.Round) *FFLRound {
	round := &FFLRound{ID: toID(r.ID), Name: r.Name}
	if r.AFLRoundID != nil {
		id := toID(*r.AFLRoundID)
		round.AflRoundID = &id
	}
	return round
}

func convertRounds(rounds []domain.Round) []*FFLRound {
	out := make([]*FFLRound, len(rounds))
	for i, r := range rounds {
		out[i] = convertRound(r)
	}
	return out
}

func convertMatch(m domain.Match) *FFLMatch {
	match := &FFLMatch{
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

func convertMatches(matches []domain.Match) []*FFLMatch {
	out := make([]*FFLMatch, len(matches))
	for i, m := range matches {
		out[i] = convertMatch(m)
	}
	return out
}

func convertClubSeason(cs domain.ClubSeason, club domain.Club, season domain.Season) *FFLClubSeason {
	return &FFLClubSeason{
		ID:         toID(cs.ID),
		Club:       convertClub(club),
		Season:     convertSeason(season),
		Played:     cs.Played,
		Won:        cs.Won,
		Lost:       cs.Lost,
		Drawn:      cs.Drawn,
		For:        cs.For,
		Against:    cs.Against,
		Percentage: cs.Percentage(),
	}
}

func convertClubMatch(cm domain.ClubMatch, club domain.Club) *FFLClubMatch {
	return &FFLClubMatch{
		ID:    toID(cm.ID),
		Club:  convertClub(club),
		Score: cm.StoredScore,
	}
}

func convertPlayerMatch(pm domain.PlayerMatch, player domain.Player) *FFLPlayerMatch {
	result := &FFLPlayerMatch{
		ID:                  toID(pm.ID),
		PlayerSeasonID:      toID(pm.PlayerSeasonID),
		Player:              convertPlayer(player),
		BackupPositions:     pm.BackupPositions,
		InterchangePosition: pm.InterchangePosition,
		Score:               pm.Score,
	}
	if pm.Position != nil {
		s := string(*pm.Position)
		result.Position = &s
	}
	if pm.Status != nil {
		s := string(*pm.Status)
		result.Status = &s
	}
	if pm.AFLPlayerMatchID != nil {
		id := toID(*pm.AFLPlayerMatchID)
		result.AflPlayerMatchID = &id
	}
	return result
}

func convertPlayerSeason(ps domain.PlayerSeason, player domain.Player) *FFLPlayerSeason {
	result := &FFLPlayerSeason{
		ID:           toID(ps.ID),
		Player:       convertPlayer(player),
		ClubSeasonID: toID(ps.ClubSeasonID),
	}
	if ps.AFLPlayerSeasonID != nil {
		id := toID(*ps.AFLPlayerSeasonID)
		result.AflPlayerSeasonID = &id
	}
	return result
}
