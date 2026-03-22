package domain

import "context"

// PointsPerGoal is the number of points a goal is worth in AFL.
const PointsPerGoal = 6

type PlayerMatch struct {
	ID             int
	ClubMatchID    int
	PlayerSeasonID int
	Kicks          int
	Handballs      int
	Marks          int
	Hitouts        int
	Tackles        int
	Goals          int
	Behinds        int
}

// Disposals returns the total disposals (kicks + handballs).
func (pm PlayerMatch) Disposals() int {
	return pm.Kicks + pm.Handballs
}

// Score returns the player's scoring contribution in points (goals * 6 + behinds).
func (pm PlayerMatch) Score() int {
	return pm.Goals*PointsPerGoal + pm.Behinds
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
	Upsert(ctx context.Context, params UpsertPlayerMatchParams) (PlayerMatch, error)
}
