package postgres

import (
	"context"

	"xffl/services/ffl/internal/domain"
	"xffl/services/ffl/internal/infrastructure/postgres/sqlcgen"
)

// derefOr returns the value pointed to by p, or zero if p is nil.
func derefOr(p *int32) int {
	if p == nil {
		return 0
	}
	return int(*p)
}

// intToInt32Ptr converts *int to *int32 for sqlc params.
func intToInt32Ptr(p *int) *int32 {
	if p == nil {
		return nil
	}
	v := int32(*p)
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

func (r *RoundRepository) FindLatest(ctx context.Context) (domain.Round, error) {
	row, err := r.q.FindLatestRound(ctx)
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
	cm.StoredScore = derefOr(row.DrvScore)

	pmRows, err := r.q.FindPlayerMatchesByClubMatchID(ctx, int32(cm.ID))
	if err != nil {
		return err
	}
	cm.PlayerMatches = make([]domain.PlayerMatch, len(pmRows))
	for i, pmRow := range pmRows {
		cm.PlayerMatches[i] = toPlayerMatch(pmRow.ID, pmRow.ClubMatchID, pmRow.PlayerSeasonID,
			pmRow.Position, pmRow.Status, pmRow.BackupPositions, pmRow.InterchangePosition, pmRow.Score, pmRow.AflPlayerMatchID)
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
			ID:       int(row.ID),
			ClubID:   int(row.ClubID),
			SeasonID: int(row.SeasonID),
			Played:   derefOr(row.DrvPlayed),
			Won:      derefOr(row.DrvWon),
			Lost:     derefOr(row.DrvLost),
			Drawn:    derefOr(row.DrvDrawn),
			For:      derefOr(row.DrvFor),
			Against:  derefOr(row.DrvAgainst),
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
		ID:       int(row.ID),
		ClubID:   int(row.ClubID),
		SeasonID: int(row.SeasonID),
		Played:   derefOr(row.DrvPlayed),
		Won:      derefOr(row.DrvWon),
		Lost:     derefOr(row.DrvLost),
		Drawn:    derefOr(row.DrvDrawn),
		For:      derefOr(row.DrvFor),
		Against:  derefOr(row.DrvAgainst),
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
			ID:           int(row.ID),
			MatchID:      int(row.MatchID),
			ClubSeasonID: int(row.ClubSeasonID),
			StoredScore:  derefOr(row.DrvScore),
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
		ID:           int(row.ID),
		MatchID:      int(row.MatchID),
		ClubSeasonID: int(row.ClubSeasonID),
		StoredScore:  derefOr(row.DrvScore),
	}, nil
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

func int32PtrToIntPtr(v *int32) *int {
	if v == nil {
		return nil
	}
	i := int(*v)
	return &i
}

func intPtrToInt32Ptr(v *int) *int32 {
	if v == nil {
		return nil
	}
	i := int32(*v)
	return &i
}

func playerFromRow(id int32, name string, aflPlayerID *int32) domain.Player {
	return domain.Player{ID: int(id), Name: name, AFLPlayerID: int32PtrToIntPtr(aflPlayerID)}
}

func (r *PlayerRepository) FindAll(ctx context.Context) ([]domain.Player, error) {
	rows, err := r.q.FindAllPlayers(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]domain.Player, len(rows))
	for i, row := range rows {
		out[i] = playerFromRow(row.ID, row.Name, row.AflPlayerID)
	}
	return out, nil
}

func (r *PlayerRepository) FindByID(ctx context.Context, id int) (domain.Player, error) {
	row, err := r.q.FindPlayerByID(ctx, int32(id))
	if err != nil {
		return domain.Player{}, err
	}
	return playerFromRow(row.ID, row.Name, row.AflPlayerID), nil
}

func (r *PlayerRepository) Create(ctx context.Context, name string) (domain.Player, error) {
	row, err := r.q.CreatePlayer(ctx, sqlcgen.CreatePlayerParams{
		Name:        name,
		AflPlayerID: nil,
	})
	if err != nil {
		return domain.Player{}, err
	}
	return playerFromRow(row.ID, row.Name, row.AflPlayerID), nil
}

func (r *PlayerRepository) Update(ctx context.Context, id int, name string) (domain.Player, error) {
	row, err := r.q.UpdatePlayer(ctx, sqlcgen.UpdatePlayerParams{
		ID:          int32(id),
		Name:        name,
		AflPlayerID: nil,
	})
	if err != nil {
		return domain.Player{}, err
	}
	return playerFromRow(row.ID, row.Name, row.AflPlayerID), nil
}

func (r *PlayerRepository) Delete(ctx context.Context, id int) error {
	return r.q.DeletePlayer(ctx, int32(id))
}

// --- PlayerMatch ---

type PlayerMatchRepository struct{ q *sqlcgen.Queries }

func NewPlayerMatchRepository(q *sqlcgen.Queries) *PlayerMatchRepository {
	return &PlayerMatchRepository{q: q}
}

func positionPtr(s *string) *domain.Position {
	if s == nil {
		return nil
	}
	p := domain.Position(*s)
	return &p
}

func statusPtr(s *string) *domain.PlayerMatchStatus {
	if s == nil {
		return nil
	}
	st := domain.PlayerMatchStatus(*s)
	return &st
}

func toPlayerMatch(id, clubMatchID, playerSeasonID int32, position, status *string, backupPositions, interchangePosition *string, score *int32, aflPlayerMatchID *int32) domain.PlayerMatch {
	return domain.PlayerMatch{
		ID:                  int(id),
		ClubMatchID:         int(clubMatchID),
		PlayerSeasonID:      int(playerSeasonID),
		Position:            positionPtr(position),
		Status:              statusPtr(status),
		BackupPositions:     backupPositions,
		InterchangePosition: interchangePosition,
		Score:               derefOr(score),
		AFLPlayerMatchID:    int32PtrToIntPtr(aflPlayerMatchID),
	}
}

func (r *PlayerMatchRepository) FindByClubMatchID(ctx context.Context, clubMatchID int) ([]domain.PlayerMatch, error) {
	rows, err := r.q.FindPlayerMatchesByClubMatchID(ctx, int32(clubMatchID))
	if err != nil {
		return nil, err
	}
	out := make([]domain.PlayerMatch, len(rows))
	for i, row := range rows {
		out[i] = toPlayerMatch(row.ID, row.ClubMatchID, row.PlayerSeasonID,
			row.Position, row.Status, row.BackupPositions, row.InterchangePosition, row.Score, row.AflPlayerMatchID)
	}
	return out, nil
}

func (r *PlayerMatchRepository) FindByID(ctx context.Context, id int) (domain.PlayerMatch, error) {
	row, err := r.q.FindPlayerMatchByID(ctx, int32(id))
	if err != nil {
		return domain.PlayerMatch{}, err
	}
	return toPlayerMatch(row.ID, row.ClubMatchID, row.PlayerSeasonID,
		row.Position, row.Status, row.BackupPositions, row.InterchangePosition, row.Score, row.AflPlayerMatchID), nil
}

func posToStringPtr(p *domain.Position) *string {
	if p == nil {
		return nil
	}
	s := string(*p)
	return &s
}

func statusToStringPtr(s *domain.PlayerMatchStatus) *string {
	if s == nil {
		return nil
	}
	str := string(*s)
	return &str
}

func (r *PlayerMatchRepository) Upsert(ctx context.Context, params domain.UpsertPlayerMatchParams) (domain.PlayerMatch, error) {
	row, err := r.q.UpsertPlayerMatch(ctx, sqlcgen.UpsertPlayerMatchParams{
		ClubMatchID:         int32(params.ClubMatchID),
		PlayerSeasonID:      int32(params.PlayerSeasonID),
		Position:            posToStringPtr(params.Position),
		Status:              statusToStringPtr(params.Status),
		BackupPositions:     params.BackupPositions,
		InterchangePosition: params.InterchangePosition,
		Score:               intToInt32Ptr(params.Score),
	})
	if err != nil {
		return domain.PlayerMatch{}, err
	}
	return toPlayerMatch(row.ID, row.ClubMatchID, row.PlayerSeasonID,
		row.Position, row.Status, row.BackupPositions, row.InterchangePosition, row.Score, row.AflPlayerMatchID), nil
}

// --- PlayerSeason ---

type PlayerSeasonRepository struct{ q *sqlcgen.Queries }

func NewPlayerSeasonRepository(q *sqlcgen.Queries) *PlayerSeasonRepository {
	return &PlayerSeasonRepository{q: q}
}

func (r *PlayerSeasonRepository) FindByClubSeasonID(ctx context.Context, clubSeasonID int) ([]domain.PlayerSeason, error) {
	rows, err := r.q.FindPlayerSeasonsByClubSeasonID(ctx, int32(clubSeasonID))
	if err != nil {
		return nil, err
	}
	out := make([]domain.PlayerSeason, len(rows))
	for i, row := range rows {
		out[i] = domain.PlayerSeason{ID: int(row.ID), PlayerID: int(row.PlayerID), ClubSeasonID: int(row.ClubSeasonID), AFLPlayerSeasonID: int32PtrToIntPtr(row.AflPlayerSeasonID)}
	}
	return out, nil
}

func (r *PlayerSeasonRepository) FindByID(ctx context.Context, id int) (domain.PlayerSeason, error) {
	row, err := r.q.FindPlayerSeasonByID(ctx, int32(id))
	if err != nil {
		return domain.PlayerSeason{}, err
	}
	return domain.PlayerSeason{ID: int(row.ID), PlayerID: int(row.PlayerID), ClubSeasonID: int(row.ClubSeasonID), AFLPlayerSeasonID: int32PtrToIntPtr(row.AflPlayerSeasonID)}, nil
}

func (r *PlayerSeasonRepository) Create(ctx context.Context, playerID int, clubSeasonID int) (domain.PlayerSeason, error) {
	row, err := r.q.CreatePlayerSeason(ctx, sqlcgen.CreatePlayerSeasonParams{
		PlayerID:     int32(playerID),
		ClubSeasonID: int32(clubSeasonID),
	})
	if err != nil {
		return domain.PlayerSeason{}, err
	}
	return domain.PlayerSeason{ID: int(row.ID), PlayerID: int(row.PlayerID), ClubSeasonID: int(row.ClubSeasonID), AFLPlayerSeasonID: int32PtrToIntPtr(row.AflPlayerSeasonID)}, nil
}

func (r *PlayerSeasonRepository) Delete(ctx context.Context, id int) error {
	return r.q.DeletePlayerSeason(ctx, int32(id))
}
