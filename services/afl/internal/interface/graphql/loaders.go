package graphql

import (
	"context"
	"fmt"

	"github.com/vikstrous/dataloadgen"

	"xffl/services/afl/internal/application"
	"xffl/services/afl/internal/domain"
)

type loadersKey struct{}

// statsKey is the DataLoader key for player season stats: player season ID plus
// the query params that determine which matches are included.
type statsKey struct {
	PlayerSeasonID int
	UpToRoundID    int // 0 = no filter
	LastN          int // 0 = no filter
}

type Loaders struct {
	PlayerByPlayerSeasonID *dataloadgen.Loader[int, *domain.Player]
	ClubByID               *dataloadgen.Loader[int, *domain.Club]
	MatchByID              *dataloadgen.Loader[int, *domain.Match]
	PlayerSeasonByID       *dataloadgen.Loader[int, *domain.PlayerSeason]
	PlayerSeasonStats      *dataloadgen.Loader[statsKey, *domain.PlayerSeasonStats]
}

func NewLoaders(q *application.Queries) *Loaders {
	return &Loaders{
		PlayerByPlayerSeasonID: dataloadgen.NewLoader(func(ctx context.Context, ids []int) ([]*domain.Player, []error) {
			m, err := q.GetPlayersForPlayerSeasonIDs(ctx, ids)
			return mapToSlice(ids, m, err)
		}),
		ClubByID: dataloadgen.NewLoader(func(ctx context.Context, ids []int) ([]*domain.Club, []error) {
			m, err := q.GetClubsByIDs(ctx, ids)
			return mapToSlice(ids, m, err)
		}),
		MatchByID: dataloadgen.NewLoader(func(ctx context.Context, ids []int) ([]*domain.Match, []error) {
			m, err := q.GetMatchesByIDs(ctx, ids)
			return mapToSlice(ids, m, err)
		}),
		PlayerSeasonByID: dataloadgen.NewLoader(func(ctx context.Context, ids []int) ([]*domain.PlayerSeason, []error) {
			m, err := q.GetPlayerSeasonsByIDs(ctx, ids)
			return mapToSlice(ids, m, err)
		}),
		PlayerSeasonStats: dataloadgen.NewLoader(func(ctx context.Context, keys []statsKey) ([]*domain.PlayerSeasonStats, []error) {
			return batchPlayerSeasonStats(ctx, keys, q)
		}),
	}
}

// batchPlayerSeasonStats groups keys by their params, makes one SQL call per
// unique (UpToRoundID, LastN) combination, and returns results positionally.
// Returns nil (not an error) for players with no qualifying matches.
func batchPlayerSeasonStats(ctx context.Context, keys []statsKey, q *application.Queries) ([]*domain.PlayerSeasonStats, []error) {
	type paramGroup struct{ upTo, lastN int }

	groups := make(map[paramGroup][]int)
	for _, k := range keys {
		pg := paramGroup{k.UpToRoundID, k.LastN}
		groups[pg] = append(groups[pg], k.PlayerSeasonID)
	}

	resultMap := make(map[statsKey]*domain.PlayerSeasonStats, len(keys))
	for pg, psIDs := range groups {
		params := domain.PlayerSeasonStatsParams{PlayerSeasonIDs: psIDs}
		if pg.upTo != 0 {
			v := pg.upTo
			params.UpToRoundID = &v
		}
		if pg.lastN != 0 {
			v := pg.lastN
			params.LastN = &v
		}
		rows, err := q.GetPlayerSeasonStats(ctx, params)
		if err != nil {
			// mark all keys in this group as errors
			errs := make([]error, len(keys))
			for i, k := range keys {
				if (paramGroup{k.UpToRoundID, k.LastN}) == pg {
					errs[i] = err
				}
			}
			return nil, errs
		}
		for _, s := range rows {
			cp := s
			resultMap[statsKey{s.PlayerSeasonID, pg.upTo, pg.lastN}] = &cp
		}
	}

	out := make([]*domain.PlayerSeasonStats, len(keys))
	for i, k := range keys {
		out[i] = resultMap[k] // nil = no qualifying matches → GraphQL null
	}
	return out, make([]error, len(keys))
}

func InjectLoaders(ctx context.Context, l *Loaders) context.Context {
	return context.WithValue(ctx, loadersKey{}, l)
}

func LoadersFromCtx(ctx context.Context) *Loaders {
	return ctx.Value(loadersKey{}).(*Loaders)
}

// mapToSlice converts a map[int]V result to a positionally-matched []*V slice,
// one entry per id.
func mapToSlice[V any](ids []int, m map[int]V, batchErr error) ([]*V, []error) {
	out := make([]*V, len(ids))
	errs := make([]error, len(ids))
	if batchErr != nil {
		for i := range errs {
			errs[i] = batchErr
		}
		return out, errs
	}
	for i, id := range ids {
		if v, ok := m[id]; ok {
			cp := v
			out[i] = &cp
		} else {
			errs[i] = fmt.Errorf("id %d: not found", id)
		}
	}
	return out, errs
}
