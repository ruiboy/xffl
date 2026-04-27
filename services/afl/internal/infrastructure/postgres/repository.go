package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

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

// derefOrStr returns the value pointed to by p, or empty string if p is nil.
func derefOrStr(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

// intToInt32Ptr converts *int to *int32 for sqlc params.
func intToInt32Ptr(p *int) *int32 {
	if p == nil {
		return nil
	}
	v := int32(*p)
	return &v
}

// int32PtrToIntPtr converts *int32 to *int for domain mapping.
func int32PtrToIntPtr(p *int32) *int {
	if p == nil {
		return nil
	}
	v := int(*p)
	return &v
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

func (r *ClubRepository) FindByIDs(ctx context.Context, ids []int) (map[int]domain.Club, error) {
	int32IDs := make([]int32, len(ids))
	for i, id := range ids {
		int32IDs[i] = int32(id)
	}
	rows, err := r.q.FindClubsByIDs(ctx, int32IDs)
	if err != nil {
		return nil, err
	}
	out := make(map[int]domain.Club, len(rows))
	for _, row := range rows {
		out[int(row.ID)] = domain.Club{ID: int(row.ID), Name: row.Name}
	}
	return out, nil
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

type RoundRepository struct {
	q    *sqlcgen.Queries
	pool *pgxpool.Pool
}

func NewRoundRepository(q *sqlcgen.Queries, pool *pgxpool.Pool) *RoundRepository {
	return &RoundRepository{q: q, pool: pool}
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

// findNeighbours returns at most two rows: the most recently started round
// (first_match_dt <= asOf) and the first upcoming round (first_match_dt > asOf).
const findNeighbours = `
WITH round_bounds AS (
    SELECT r.id, r.name, r.season_id, MIN(m.start_dt) AS first_match_dt
    FROM afl.round r
    JOIN afl.match m ON m.round_id = r.id AND m.deleted_at IS NULL
    WHERE r.deleted_at IS NULL
    GROUP BY r.id, r.name, r.season_id
    HAVING MIN(m.start_dt) IS NOT NULL
)
(SELECT id, name, season_id, first_match_dt FROM round_bounds
 WHERE first_match_dt <= $1 ORDER BY first_match_dt DESC LIMIT 1)
UNION ALL
(SELECT id, name, season_id, first_match_dt FROM round_bounds
 WHERE first_match_dt > $1 ORDER BY first_match_dt ASC LIMIT 1)
`

func (r *RoundRepository) FindNeighbours(ctx context.Context, asOf time.Time) ([]domain.RoundWithStart, error) {
	rows, err := r.pool.Query(ctx, findNeighbours, asOf)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.RoundWithStart
	for rows.Next() {
		var id, sid int32
		var name string
		var first pgtype.Timestamptz
		if err := rows.Scan(&id, &name, &sid, &first); err != nil {
			return nil, fmt.Errorf("scan round neighbours: %w", err)
		}
		if !first.Valid {
			continue
		}
		out = append(out, domain.RoundWithStart{
			Round:          domain.Round{ID: int(id), Name: name, SeasonID: int(sid)},
			FirstMatchTime: first.Time,
		})
	}
	return out, rows.Err()
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
			ID:        int(row.ID),
			RoundID:   int(row.RoundID),
			Home:      domain.ClubMatch{ID: int(row.HomeClubMatchID)},
			Away:      domain.ClubMatch{ID: int(row.AwayClubMatchID)},
			Venue:     row.Venue,
			StartTime: row.StartDt.Time,
			Result:    domain.MatchResult(row.DrvResult),
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
		ID:        int(row.ID),
		RoundID:   int(row.RoundID),
		Home:      domain.ClubMatch{ID: int(row.HomeClubMatchID)},
		Away:      domain.ClubMatch{ID: int(row.AwayClubMatchID)},
		Venue:     row.Venue,
		StartTime: row.StartDt.Time,
		Result:    domain.MatchResult(row.DrvResult),
	}, nil
}

func (r *MatchRepository) FindByIDs(ctx context.Context, ids []int) (map[int]domain.Match, error) {
	int32IDs := make([]int32, len(ids))
	for i, id := range ids {
		int32IDs[i] = int32(id)
	}
	rows, err := r.q.FindMatchesByIDs(ctx, int32IDs)
	if err != nil {
		return nil, err
	}
	out := make(map[int]domain.Match, len(rows))
	for _, row := range rows {
		out[int(row.ID)] = domain.Match{
			ID:        int(row.ID),
			RoundID:   int(row.RoundID),
			Home:      domain.ClubMatch{ID: int(row.HomeClubMatchID)},
			Away:      domain.ClubMatch{ID: int(row.AwayClubMatchID)},
			Venue:     row.Venue,
			StartTime: row.StartDt.Time,
			Result:    domain.MatchResult(row.DrvResult),
		}
	}
	return out, nil
}

func (r *MatchRepository) FindByIDWithDetails(ctx context.Context, id int) (domain.Match, error) {
	match, err := r.FindByID(ctx, id)
	if err != nil {
		return domain.Match{}, err
	}

	if err := r.hydrateClubMatch(ctx, &match.Home); err != nil {
		return domain.Match{}, err
	}
	if err := r.hydrateClubMatch(ctx, &match.Away); err != nil {
		return domain.Match{}, err
	}
	return match, nil
}

func (r *MatchRepository) hydrateClubMatch(ctx context.Context, cm *domain.ClubMatch) error {
	if cm.ID == 0 {
		return nil
	}
	row, err := r.q.FindClubMatchByID(ctx, int32(cm.ID))
	if err != nil {
		return err
	}
	cm.MatchID = int(row.MatchID)
	cm.ClubSeasonID = int(row.ClubSeasonID)
	cm.RushedBehinds = derefOr(row.RushedBehinds)
	cm.StoredScore = derefOr(row.DrvScore)

	pmRows, err := r.q.FindPlayerMatchesByClubMatchID(ctx, int32(cm.ID))
	if err != nil {
		return err
	}
	cm.PlayerMatches = make([]domain.PlayerMatch, len(pmRows))
	for i, pmRow := range pmRows {
		cm.PlayerMatches[i] = domain.PlayerMatch{
			ID:             int(pmRow.ID),
			ClubMatchID:    int(pmRow.ClubMatchID),
			PlayerSeasonID: int(pmRow.PlayerSeasonID),
			Kicks:          derefOr(pmRow.Kicks),
			Handballs:      derefOr(pmRow.Handballs),
			Marks:          derefOr(pmRow.Marks),
			Hitouts:        derefOr(pmRow.Hitouts),
			Tackles:        derefOr(pmRow.Tackles),
			Goals:          derefOr(pmRow.Goals),
			Behinds:        derefOr(pmRow.Behinds),
		}
	}
	return nil
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
			StoredScore:   derefOr(row.DrvScore),
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
		StoredScore:   derefOr(row.DrvScore),
	}, nil
}

func (r *ClubMatchRepository) FindRoundID(ctx context.Context, clubMatchID int) (int, error) {
	roundID, err := r.q.FindRoundIDByClubMatchID(ctx, int32(clubMatchID))
	if err != nil {
		return 0, err
	}
	return int(roundID), nil
}

func (r *ClubMatchRepository) UpdateScore(ctx context.Context, id int, score int) error {
	s := int32(score)
	return r.q.UpdateClubMatchScore(ctx, sqlcgen.UpdateClubMatchScoreParams{
		ID:       int32(id),
		DrvScore: &s,
	})
}

// --- Player ---

type PlayerRepository struct{ q *sqlcgen.Queries }

func NewPlayerRepository(q *sqlcgen.Queries) *PlayerRepository {
	return &PlayerRepository{q: q}
}

func (r *PlayerRepository) FindByID(ctx context.Context, id int) (domain.Player, error) {
	row, err := r.q.FindPlayerByID(ctx, int32(id))
	if err != nil {
		return domain.Player{}, err
	}
	return domain.Player{ID: int(row.ID), Name: row.Name}, nil
}

func (r *PlayerRepository) FindByIDs(ctx context.Context, ids []int) ([]domain.Player, error) {
	int32IDs := make([]int32, len(ids))
	for i, id := range ids {
		int32IDs[i] = int32(id)
	}
	rows, err := r.q.FindPlayersByIDs(ctx, int32IDs)
	if err != nil {
		return nil, err
	}
	players := make([]domain.Player, len(rows))
	for i, row := range rows {
		players[i] = domain.Player{ID: int(row.ID), Name: row.Name}
	}
	return players, nil
}

func (r *PlayerRepository) Search(ctx context.Context, query string) ([]domain.Player, error) {
	rows, err := r.q.SearchPlayersByName(ctx, &query)
	if err != nil {
		return nil, err
	}
	players := make([]domain.Player, len(rows))
	for i, row := range rows {
		players[i] = domain.Player{ID: int(row.ID), Name: row.Name}
	}
	return players, nil
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
			Status:         derefOrStr(row.Status),
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
		Status:         derefOrStr(row.Status),
		Kicks:          derefOr(row.Kicks),
		Handballs:      derefOr(row.Handballs),
		Marks:          derefOr(row.Marks),
		Hitouts:        derefOr(row.Hitouts),
		Tackles:        derefOr(row.Tackles),
		Goals:          derefOr(row.Goals),
		Behinds:        derefOr(row.Behinds),
	}, nil
}

func (r *PlayerMatchRepository) Upsert(ctx context.Context, params domain.UpsertPlayerMatchParams) (domain.PlayerMatch, error) {
	row, err := r.q.UpsertPlayerMatch(ctx, sqlcgen.UpsertPlayerMatchParams{
		ClubMatchID:    int32(params.ClubMatchID),
		PlayerSeasonID: int32(params.PlayerSeasonID),
		Status:         params.Status,
		Kicks:          intToInt32Ptr(params.Kicks),
		Handballs:      intToInt32Ptr(params.Handballs),
		Marks:          intToInt32Ptr(params.Marks),
		Hitouts:        intToInt32Ptr(params.Hitouts),
		Tackles:        intToInt32Ptr(params.Tackles),
		Goals:          intToInt32Ptr(params.Goals),
		Behinds:        intToInt32Ptr(params.Behinds),
	})
	if err != nil {
		return domain.PlayerMatch{}, err
	}
	return domain.PlayerMatch{
		ID:             int(row.ID),
		ClubMatchID:    int(row.ClubMatchID),
		PlayerSeasonID: int(row.PlayerSeasonID),
		Status:         derefOrStr(row.Status),
		Kicks:          derefOr(row.Kicks),
		Handballs:      derefOr(row.Handballs),
		Marks:          derefOr(row.Marks),
		Hitouts:        derefOr(row.Hitouts),
		Tackles:        derefOr(row.Tackles),
		Goals:          derefOr(row.Goals),
		Behinds:        derefOr(row.Behinds),
	}, nil
}

func (r *PlayerMatchRepository) FindStatsByPlayerSeasonIDs(ctx context.Context, ids []int) ([]domain.PlayerSeasonStats, error) {
	int32IDs := make([]int32, len(ids))
	for i, id := range ids {
		int32IDs[i] = int32(id)
	}
	rows, err := r.q.FindPlayerMatchStatsByPlayerSeasonIDs(ctx, int32IDs)
	if err != nil {
		return nil, err
	}
	out := make([]domain.PlayerSeasonStats, len(rows))
	for i, row := range rows {
		out[i] = domain.PlayerSeasonStats{
			PlayerSeasonID: int(row.PlayerSeasonID),
			GamesPlayed:    int(row.GamesPlayed),
			AvgKicks:       row.AvgKicks,
			AvgHandballs:   row.AvgHandballs,
			AvgMarks:       row.AvgMarks,
			AvgHitouts:     row.AvgHitouts,
			AvgTackles:     row.AvgTackles,
			AvgGoals:       row.AvgGoals,
			AvgBehinds:     row.AvgBehinds,
		}
	}
	return out, nil
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
		FromRoundID:  int32PtrToIntPtr(row.FromRoundID),
		ToRoundID:    int32PtrToIntPtr(row.ToRoundID),
	}, nil
}

func (r *PlayerSeasonRepository) FindPlayersForPlayerSeasonIDs(ctx context.Context, ids []int) (map[int]domain.Player, error) {
	int32IDs := make([]int32, len(ids))
	for i, id := range ids {
		int32IDs[i] = int32(id)
	}
	rows, err := r.q.FindPlayersByPlayerSeasonIDs(ctx, int32IDs)
	if err != nil {
		return nil, err
	}
	out := make(map[int]domain.Player, len(rows))
	for _, row := range rows {
		out[int(row.PlayerSeasonID)] = domain.Player{ID: int(row.PlayerID), Name: row.PlayerName}
	}
	return out, nil
}
