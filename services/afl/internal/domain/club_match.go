package domain

import "context"

type ClubMatch struct {
	ID            int
	MatchID       int
	ClubSeasonID  int
	RushedBehinds int
	StoredScore   int
	PlayerMatches []PlayerMatch
}

// Score computes the total score from player contributions and rushed behinds.
func (cm ClubMatch) Score() int {
	total := cm.RushedBehinds
	for _, pm := range cm.PlayerMatches {
		total += pm.Score()
	}
	return total
}

type ClubMatchRepository interface {
	FindByMatchID(ctx context.Context, matchID int) ([]ClubMatch, error)
	FindByID(ctx context.Context, id int) (ClubMatch, error)
	FindRoundID(ctx context.Context, clubMatchID int) (int, error)
	UpdateScore(ctx context.Context, id int, score int) error
}
