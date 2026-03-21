package postgres

import (
	"context"

	"xffl/services/afl/internal/domain"
	"xffl/services/afl/internal/infrastructure/postgres/sqlcgen"
)

// derefOr returns the value pointed to by p, or zero if p is nil.
func derefOr(p *int32) int {
	if p == nil {
		return 0
	}
	return int(*p)
}

// --- Club ---

type ClubRepository struct{ q *sqlcgen.Queries }

func NewClubRepository(q *sqlcgen.Queries) *ClubRepository {
	return &ClubRepository{q: q}
}

func (r *ClubRepository) FindAll(ctx context.Context) ([]domain.Club, error) {
	rows, err := r.q.FindAllClubs(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]domain.Club, len(rows))
	for i, row := range rows {
		out[i] = domain.Club{ID: int(row.ID), Name: row.Name}
	}
	return out, nil
}

func (r *ClubRepository) FindByID(ctx context.Context, id int) (domain.Club, error) {
	row, err := r.q.FindClubByID(ctx, int32(id))
	if err != nil {
		return domain.Club{}, err
	}
	return domain.Club{ID: int(row.ID), Name: row.Name}, nil
}

// --- Season ---

type SeasonRepository struct{ q *sqlcgen.Queries }

func NewSeasonRepository(q *sqlcgen.Queries) *SeasonRepository {
	return &SeasonRepository{q: q}
}

func (r *SeasonRepository) FindAll(ctx context.Context) ([]domain.Season, error) {
	rows, err := r.q.FindAllSeasons(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]domain.Season, len(rows))
	for i, row := range rows {
		out[i] = domain.Season{ID: int(row.ID), Name: row.Name, LeagueID: int(row.LeagueID)}
	}
	return out, nil
}

func (r *SeasonRepository) FindByID(ctx context.Context, id int) (domain.Season, error) {
	row, err := r.q.FindSeasonByID(ctx, int32(id))
	if err != nil {
		return domain.Season{}, err
	}
	return domain.Season{ID: int(row.ID), Name: row.Name, LeagueID: int(row.LeagueID)}, nil
}

// --- Round ---

type RoundRepository struct{ q *sqlcgen.Queries }

func NewRoundRepository(q *sqlcgen.Queries) *RoundRepository {
	return &RoundRepository{q: q}
}

func (r *RoundRepository) FindBySeasonID(ctx context.Context, seasonID int) ([]domain.Round, error) {
	rows, err := r.q.FindRoundsBySeasonID(ctx, int32(seasonID))
	if err != nil {
		return nil, err
	}
	out := make([]domain.Round, len(rows))
	for i, row := range rows {
		out[i] = domain.Round{ID: int(row.ID), Name: row.Name, SeasonID: int(row.SeasonID)}
	}
	return out, nil
}

func (r *RoundRepository) FindByID(ctx context.Context, id int) (domain.Round, error) {
	row, err := r.q.FindRoundByID(ctx, int32(id))
	if err != nil {
		return domain.Round{}, err
	}
	return domain.Round{ID: int(row.ID), Name: row.Name, SeasonID: int(row.SeasonID)}, nil
}

// --- Match ---

type MatchRepository struct{ q *sqlcgen.Queries }

func NewMatchRepository(q *sqlcgen.Queries) *MatchRepository {
	return &MatchRepository{q: q}
}

func (r *MatchRepository) FindByRoundID(ctx context.Context, roundID int) ([]domain.Match, error) {
	rows, err := r.q.FindMatchesByRoundID(ctx, int32(roundID))
	if err != nil {
		return nil, err
	}
	out := make([]domain.Match, len(rows))
	for i, row := range rows {
		out[i] = domain.Match{
			ID:              int(row.ID),
			RoundID:         int(row.RoundID),
			HomeClubMatchID: int(row.HomeClubMatchID),
			AwayClubMatchID: int(row.AwayClubMatchID),
			Venue:           row.Venue,
			StartTime:       row.StartDt.Time,
			Result:          domain.MatchResult(row.DrvResult),
		}
	}
	return out, nil
}

func (r *MatchRepository) FindByID(ctx context.Context, id int) (domain.Match, error) {
	row, err := r.q.FindMatchByID(ctx, int32(id))
	if err != nil {
		return domain.Match{}, err
	}
	return domain.Match{
		ID:              int(row.ID),
		RoundID:         int(row.RoundID),
		HomeClubMatchID: int(row.HomeClubMatchID),
		AwayClubMatchID: int(row.AwayClubMatchID),
		Venue:           row.Venue,
		StartTime:       row.StartDt.Time,
		Result:          domain.MatchResult(row.DrvResult),
	}, nil
}

// --- ClubSeason ---

type ClubSeasonRepository struct{ q *sqlcgen.Queries }

func NewClubSeasonRepository(q *sqlcgen.Queries) *ClubSeasonRepository {
	return &ClubSeasonRepository{q: q}
}

func (r *ClubSeasonRepository) FindBySeasonID(ctx context.Context, seasonID int) ([]domain.ClubSeason, error) {
	rows, err := r.q.FindClubSeasonsBySeasonID(ctx, int32(seasonID))
	if err != nil {
		return nil, err
	}
	out := make([]domain.ClubSeason, len(rows))
	for i, row := range rows {
		out[i] = domain.ClubSeason{
			ID:                int(row.ID),
			ClubID:            int(row.ClubID),
			SeasonID:          int(row.SeasonID),
			Played:            derefOr(row.DrvPlayed),
			Won:               derefOr(row.DrvWon),
			Lost:              derefOr(row.DrvLost),
			Drawn:             derefOr(row.DrvDrawn),
			For:               derefOr(row.DrvFor),
			Against:           derefOr(row.DrvAgainst),
			PremiershipPoints: derefOr(row.DrvPremiershipPoints),
		}
	}
	return out, nil
}

func (r *ClubSeasonRepository) FindByID(ctx context.Context, id int) (domain.ClubSeason, error) {
	row, err := r.q.FindClubSeasonByID(ctx, int32(id))
	if err != nil {
		return domain.ClubSeason{}, err
	}
	return domain.ClubSeason{
		ID:                int(row.ID),
		ClubID:            int(row.ClubID),
		SeasonID:          int(row.SeasonID),
		Played:            derefOr(row.DrvPlayed),
		Won:               derefOr(row.DrvWon),
		Lost:              derefOr(row.DrvLost),
		Drawn:             derefOr(row.DrvDrawn),
		For:               derefOr(row.DrvFor),
		Against:           derefOr(row.DrvAgainst),
		PremiershipPoints: derefOr(row.DrvPremiershipPoints),
	}, nil
}

// --- ClubMatch ---

type ClubMatchRepository struct{ q *sqlcgen.Queries }

func NewClubMatchRepository(q *sqlcgen.Queries) *ClubMatchRepository {
	return &ClubMatchRepository{q: q}
}

func (r *ClubMatchRepository) FindByMatchID(ctx context.Context, matchID int) ([]domain.ClubMatch, error) {
	rows, err := r.q.FindClubMatchesByMatchID(ctx, int32(matchID))
	if err != nil {
		return nil, err
	}
	out := make([]domain.ClubMatch, len(rows))
	for i, row := range rows {
		out[i] = domain.ClubMatch{
			ID:            int(row.ID),
			MatchID:       int(row.MatchID),
			ClubSeasonID:  int(row.ClubSeasonID),
			RushedBehinds: derefOr(row.RushedBehinds),
			Score:         derefOr(row.DrvScore),
		}
	}
	return out, nil
}

func (r *ClubMatchRepository) FindByID(ctx context.Context, id int) (domain.ClubMatch, error) {
	row, err := r.q.FindClubMatchByID(ctx, int32(id))
	if err != nil {
		return domain.ClubMatch{}, err
	}
	return domain.ClubMatch{
		ID:            int(row.ID),
		MatchID:       int(row.MatchID),
		ClubSeasonID:  int(row.ClubSeasonID),
		RushedBehinds: derefOr(row.RushedBehinds),
		Score:         derefOr(row.DrvScore),
	}, nil
}

// --- Player ---

type PlayerRepository struct{ q *sqlcgen.Queries }

func NewPlayerRepository(q *sqlcgen.Queries) *PlayerRepository {
	return &PlayerRepository{q: q}
}

func (r *PlayerRepository) FindByClubID(ctx context.Context, clubID int) ([]domain.Player, error) {
	id := int32(clubID)
	rows, err := r.q.FindPlayersByClubID(ctx, &id)
	if err != nil {
		return nil, err
	}
	out := make([]domain.Player, len(rows))
	for i, row := range rows {
		out[i] = domain.Player{ID: int(row.ID), Name: row.Name, ClubID: int(row.ClubID)}
	}
	return out, nil
}

func (r *PlayerRepository) FindByID(ctx context.Context, id int) (domain.Player, error) {
	row, err := r.q.FindPlayerByID(ctx, int32(id))
	if err != nil {
		return domain.Player{}, err
	}
	return domain.Player{ID: int(row.ID), Name: row.Name, ClubID: int(row.ClubID)}, nil
}

// --- PlayerMatch ---

type PlayerMatchRepository struct{ q *sqlcgen.Queries }

func NewPlayerMatchRepository(q *sqlcgen.Queries) *PlayerMatchRepository {
	return &PlayerMatchRepository{q: q}
}

func (r *PlayerMatchRepository) FindByClubMatchID(ctx context.Context, clubMatchID int) ([]domain.PlayerMatch, error) {
	rows, err := r.q.FindPlayerMatchesByClubMatchID(ctx, int32(clubMatchID))
	if err != nil {
		return nil, err
	}
	out := make([]domain.PlayerMatch, len(rows))
	for i, row := range rows {
		out[i] = domain.PlayerMatch{
			ID:             int(row.ID),
			ClubMatchID:    int(row.ClubMatchID),
			PlayerSeasonID: int(row.PlayerSeasonID),
			Kicks:          derefOr(row.Kicks),
			Handballs:      derefOr(row.Handballs),
			Marks:          derefOr(row.Marks),
			Hitouts:        derefOr(row.Hitouts),
			Tackles:        derefOr(row.Tackles),
			Goals:          derefOr(row.Goals),
			Behinds:        derefOr(row.Behinds),
		}
	}
	return out, nil
}

func (r *PlayerMatchRepository) FindByID(ctx context.Context, id int) (domain.PlayerMatch, error) {
	row, err := r.q.FindPlayerMatchByID(ctx, int32(id))
	if err != nil {
		return domain.PlayerMatch{}, err
	}
	return domain.PlayerMatch{
		ID:             int(row.ID),
		ClubMatchID:    int(row.ClubMatchID),
		PlayerSeasonID: int(row.PlayerSeasonID),
		Kicks:          derefOr(row.Kicks),
		Handballs:      derefOr(row.Handballs),
		Marks:          derefOr(row.Marks),
		Hitouts:        derefOr(row.Hitouts),
		Tackles:        derefOr(row.Tackles),
		Goals:          derefOr(row.Goals),
		Behinds:        derefOr(row.Behinds),
	}, nil
}

// --- PlayerSeason ---

type PlayerSeasonRepository struct{ q *sqlcgen.Queries }

func NewPlayerSeasonRepository(q *sqlcgen.Queries) *PlayerSeasonRepository {
	return &PlayerSeasonRepository{q: q}
}

func (r *PlayerSeasonRepository) FindByID(ctx context.Context, id int) (domain.PlayerSeason, error) {
	row, err := r.q.FindPlayerSeasonByID(ctx, int32(id))
	if err != nil {
		return domain.PlayerSeason{}, err
	}
	return domain.PlayerSeason{
		ID:           int(row.ID),
		PlayerID:     int(row.PlayerID),
		ClubSeasonID: int(row.ClubSeasonID),
	}, nil
}
