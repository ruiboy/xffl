package domain

import "context"

// PointsPerGoal is the number of points a goal is worth in AFL.
const PointsPerGoal = 6

// PlayerMatch holds AFL player match data. MatchDataStatus is populated when loaded via
// a query that joins afl.match; it is empty for Upsert return values.
type PlayerMatch struct {
	ID              int
	ClubMatchID     int
	PlayerSeasonID  int
	MatchDataStatus string
	Kicks           int
	Handballs       int
	Marks           int
	Hitouts         int
	Tackles         int
	Goals           int
	Behinds         int
}

// Disposals returns the total disposals (kicks + handballs).
func (pm PlayerMatch) Disposals() int {
	return pm.Kicks + pm.Handballs
}

// Score returns the player's scoring contribution in points (goals * 6 + behinds).
func (pm PlayerMatch) Score() int {
	return pm.Goals*PointsPerGoal + pm.Behinds
}

// AFLPlayerMatchStatus derives the AFL player match status from the match's data_status.
// A player_match row existing means the player has stats; match finality determines whether
// they are playing, played, or merely named (pre-match / no data yet).
func (pm PlayerMatch) AFLPlayerMatchStatus() string {
	switch MatchDataStatus(pm.MatchDataStatus) {
	case MatchDataFinal:
		return "played"
	case MatchDataPartial:
		return "playing"
	default:
		return "named"
	}
}

// UpsertPlayerMatchParams holds optional fields for creating or updating a PlayerMatch.
// Nil fields are left unchanged on update.
type UpsertPlayerMatchParams struct {
	ClubMatchID    int
	PlayerSeasonID int
	Kicks          *int
	Handballs      *int
	Marks          *int
	Hitouts        *int
	Tackles        *int
	Goals          *int
	Behinds        *int
}

type PlayerMatchRepository interface {
	FindByClubMatchID(ctx context.Context, clubMatchID int) ([]PlayerMatch, error)
	FindByID(ctx context.Context, id int) (PlayerMatch, error)
	FindByIDs(ctx context.Context, ids []int) ([]PlayerMatch, error)
	FindByPlayerSeasonID(ctx context.Context, playerSeasonID int) ([]PlayerMatch, error)
	FindByPlayerSeasonIDsAndRoundID(ctx context.Context, playerSeasonIDs []int, roundID int) ([]PlayerMatch, error)
	Upsert(ctx context.Context, params UpsertPlayerMatchParams) (PlayerMatch, error)
}
