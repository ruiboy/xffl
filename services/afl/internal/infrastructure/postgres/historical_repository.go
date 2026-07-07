package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"xffl/services/afl/internal/application"
	"xffl/services/afl/internal/infrastructure/postgres/sqlcgen"
)

// HistoricalRepository implements application.HistoricalRepo for the one-time
// afltables backfill (cmd/afltables-import). It uses idempotent get-or-create
// semantics so a run can be safely repeated.
type HistoricalRepository struct{ q *sqlcgen.Queries }

func NewHistoricalRepository(pool *pgxpool.Pool) *HistoricalRepository {
	return &HistoricalRepository{q: sqlcgen.New(pool)}
}

func (r *HistoricalRepository) UpsertLeague(ctx context.Context, name string) (int, error) {
	id, err := r.q.UpsertLeagueByName(ctx, name)
	return int(id), err
}

func (r *HistoricalRepository) GetOrCreateSeason(ctx context.Context, leagueID int, name string) (int, error) {
	id, err := r.q.FindSeasonByLeagueAndName(ctx, sqlcgen.FindSeasonByLeagueAndNameParams{
		LeagueID: int32(leagueID), Name: name,
	})
	if err == nil {
		return int(id), nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return 0, err
	}
	newID, err := r.q.InsertSeasonReturningID(ctx, sqlcgen.InsertSeasonReturningIDParams{
		LeagueID: int32(leagueID), Name: name,
	})
	return int(newID), err
}

func (r *HistoricalRepository) GetOrCreateRound(ctx context.Context, seasonID int, name string) (int, error) {
	id, err := r.q.FindRoundBySeasonAndName(ctx, sqlcgen.FindRoundBySeasonAndNameParams{
		SeasonID: int32(seasonID), Name: name,
	})
	if err == nil {
		return int(id), nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return 0, err
	}
	newID, err := r.q.InsertRoundReturningID(ctx, sqlcgen.InsertRoundReturningIDParams{
		SeasonID: int32(seasonID), Name: name,
	})
	return int(newID), err
}

func (r *HistoricalRepository) UpsertClub(ctx context.Context, name string) (int, error) {
	id, err := r.q.UpsertClubByName(ctx, name)
	return int(id), err
}

func (r *HistoricalRepository) GetOrCreateClubSeason(ctx context.Context, clubID, seasonID int) (int, error) {
	id, err := r.q.UpsertClubSeasonReturningID(ctx, sqlcgen.UpsertClubSeasonReturningIDParams{
		ClubID: int32(clubID), SeasonID: int32(seasonID),
	})
	return int(id), err
}

func (r *HistoricalRepository) FindMatchByRoundAndHomeClubSeason(ctx context.Context, roundID, homeClubSeasonID int) (int, bool, error) {
	id, err := r.q.FindMatchByRoundAndHomeClubSeason(ctx, sqlcgen.FindMatchByRoundAndHomeClubSeasonParams{
		RoundID: int32(roundID), ClubSeasonID: int32(homeClubSeasonID),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, err
	}
	return int(id), true, nil
}

func (r *HistoricalRepository) InsertMatch(ctx context.Context, roundID int, venue string, startDt time.Time) (int, error) {
	ts := pgtype.Timestamptz{}
	if !startDt.IsZero() {
		ts = pgtype.Timestamptz{Time: startDt, Valid: true}
	}
	var venuePtr *string
	if venue != "" {
		venuePtr = &venue
	}
	id, err := r.q.InsertHistoricalMatch(ctx, sqlcgen.InsertHistoricalMatchParams{
		RoundID: int32(roundID),
		Venue:   venuePtr,
		StartDt: ts,
	})
	return int(id), err
}

func (r *HistoricalRepository) UpsertClubMatch(ctx context.Context, matchID, clubSeasonID int, side string) (int, error) {
	id, err := r.q.UpsertClubMatchReturningID(ctx, sqlcgen.UpsertClubMatchReturningIDParams{
		MatchID: int32(matchID), ClubSeasonID: int32(clubSeasonID), Side: side,
	})
	return int(id), err
}

func (r *HistoricalRepository) FindPlayersByExactName(ctx context.Context, name string) ([]application.PlayerRef, error) {
	rows, err := r.q.FindPlayersByExactName(ctx, name)
	if err != nil {
		return nil, err
	}
	out := make([]application.PlayerRef, len(rows))
	for i, row := range rows {
		out[i] = application.PlayerRef{ID: int(row.ID), Name: row.Name}
	}
	return out, nil
}

func (r *HistoricalRepository) AllPlayers(ctx context.Context) ([]application.PlayerRef, error) {
	rows, err := r.q.FindAllPlayers(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]application.PlayerRef, len(rows))
	for i, row := range rows {
		out[i] = application.PlayerRef{ID: int(row.ID), Name: row.Name}
	}
	return out, nil
}

func (r *HistoricalRepository) CreatePlayer(ctx context.Context, name string) (int, error) {
	p, err := r.q.InsertPlayer(ctx, name)
	return int(p.ID), err
}

func (r *HistoricalRepository) GetOrCreatePlayerSeason(ctx context.Context, playerID, clubSeasonID int) (int, error) {
	ps, err := r.q.UpsertPlayerSeason(ctx, sqlcgen.UpsertPlayerSeasonParams{
		PlayerID: int32(playerID), ClubSeasonID: int32(clubSeasonID),
	})
	return int(ps.ID), err
}

func (r *HistoricalRepository) UpsertPlayerMatch(ctx context.Context, p application.PlayerMatchInput) error {
	k, hb, mk, ho, tk, gl, bh := int32(p.Kicks), int32(p.Handballs), int32(p.Marks), int32(p.Hitouts), int32(p.Tackles), int32(p.Goals), int32(p.Behinds)
	_, err := r.q.UpsertPlayerMatch(ctx, sqlcgen.UpsertPlayerMatchParams{
		ClubMatchID:    int32(p.ClubMatchID),
		PlayerSeasonID: int32(p.PlayerSeasonID),
		Kicks:          &k,
		Handballs:      &hb,
		Marks:          &mk,
		Hitouts:        &ho,
		Tackles:        &tk,
		Goals:          &gl,
		Behinds:        &bh,
	})
	return err
}

func (r *HistoricalRepository) FindXref(ctx context.Context, source, season, club, player string) (int, bool, error) {
	id, err := r.q.FindDataopsPlayerSource(ctx, sqlcgen.FindDataopsPlayerSourceParams{
		Source: source, ExternalSeason: season, ExternalClub: club, ExternalPlayer: player,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, err
	}
	return int(id), true, nil
}

func (r *HistoricalRepository) StoreXref(ctx context.Context, source, season, club, player string, playerSeasonID int) error {
	return r.q.UpsertDataopsPlayerSource(ctx, sqlcgen.UpsertDataopsPlayerSourceParams{
		Source: source, ExternalSeason: season, ExternalClub: club, ExternalPlayer: player,
		PlayerSeasonID: int32(playerSeasonID),
	})
}

