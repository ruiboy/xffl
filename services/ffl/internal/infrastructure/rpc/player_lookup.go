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
