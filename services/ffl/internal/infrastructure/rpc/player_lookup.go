package rpc

import (
	"context"
	"net/http"

	aflv1 "xffl/contracts/gen/afl/v1"
	"xffl/services/ffl/internal/application"
)

type AFLPlayerLookup struct {
	client aflv1.PlayerLookup
}

func NewAFLPlayerLookup(aflBaseURL string) *AFLPlayerLookup {
	return &AFLPlayerLookup{
		client: aflv1.NewPlayerLookupProtobufClient(aflBaseURL, &http.Client{}),
	}
}

func (a *AFLPlayerLookup) LookupPlayers(ctx context.Context, aflPlayerIDs []int) ([]application.PlayerCandidate, error) {
	ids := make([]int32, len(aflPlayerIDs))
	for i, id := range aflPlayerIDs {
		ids[i] = int32(id)
	}

	resp, err := a.client.LookupPlayers(ctx, &aflv1.LookupPlayersRequest{AflPlayerIds: ids})
	if err != nil {
		return nil, err
	}

	candidates := make([]application.PlayerCandidate, len(resp.Players))
	for i, p := range resp.Players {
		candidates[i] = application.PlayerCandidate{
			AFLPlayerID: int(p.Id),
			Name:        p.Name,
			Club:        p.ClubName,
		}
	}
	return candidates, nil
}

func (a *AFLPlayerLookup) LookupPlayerSeason(ctx context.Context, aflPlayerSeasonID int) (int, error) {
	resp, err := a.client.LookupPlayerSeason(ctx, &aflv1.LookupPlayerSeasonRequest{PlayerSeasonId: int32(aflPlayerSeasonID)})
	if err != nil {
		return 0, err
	}
	return int(resp.PlayerId), nil
}

func (a *AFLPlayerLookup) LookupPlayerMatch(ctx context.Context, aflPlayerMatchIDs []int) ([]application.PlayerMatchStats, error) {
	ids := make([]int32, len(aflPlayerMatchIDs))
	for i, id := range aflPlayerMatchIDs {
		ids[i] = int32(id)
	}
	resp, err := a.client.LookupPlayerMatch(ctx, &aflv1.LookupPlayerMatchRequest{
		Key: &aflv1.LookupPlayerMatchRequest_ByIds{
			ByIds: &aflv1.LookupByIDs{Ids: ids},
		},
	})
	if err != nil {
		return nil, err
	}
	return toPlayerMatchStats(resp.Stats), nil
}

func (a *AFLPlayerLookup) LookupPlayerMatchBySeasonRound(ctx context.Context, aflPlayerSeasonIDs []int, aflRoundID int) ([]application.PlayerMatchStats, error) {
	psIDs := make([]int32, len(aflPlayerSeasonIDs))
	for i, id := range aflPlayerSeasonIDs {
		psIDs[i] = int32(id)
	}
	resp, err := a.client.LookupPlayerMatch(ctx, &aflv1.LookupPlayerMatchRequest{
		Key: &aflv1.LookupPlayerMatchRequest_BySeasonRound{
			BySeasonRound: &aflv1.LookupBySeasonRound{
				PlayerSeasonIds: psIDs,
				RoundId:         int32(aflRoundID),
			},
		},
	})
	if err != nil {
		return nil, err
	}
	return toPlayerMatchStats(resp.Stats), nil
}

func toPlayerMatchStats(stats []*aflv1.PlayerMatchStats) []application.PlayerMatchStats {
	out := make([]application.PlayerMatchStats, len(stats))
	for i, s := range stats {
		out[i] = application.PlayerMatchStats{
			ID:             int(s.Id),
			Status:         s.Status,
			Goals:          int(s.Goals),
			Kicks:          int(s.Kicks),
			Handballs:      int(s.Handballs),
			Marks:          int(s.Marks),
			Tackles:        int(s.Tackles),
			Hitouts:        int(s.Hitouts),
			PlayerSeasonID: int(s.PlayerSeasonId),
		}
	}
	return out
}
