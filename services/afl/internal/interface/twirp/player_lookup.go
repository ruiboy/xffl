package twirp

import (
	"context"

	aflv1 "xffl/contracts/gen/afl/v1"
	"xffl/services/afl/internal/domain"
)

type playerLookupServer struct {
	players domain.PlayerRepository
}

func NewPlayerLookupServer(players domain.PlayerRepository) aflv1.PlayerLookup {
	return &playerLookupServer{players: players}
}

func (s *playerLookupServer) LookupPlayers(ctx context.Context, req *aflv1.LookupPlayersRequest) (*aflv1.LookupPlayersResponse, error) {
	ids := make([]int, len(req.AflPlayerIds))
	for i, id := range req.AflPlayerIds {
		ids[i] = int(id)
	}

	players, err := s.players.FindByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}

	infos := make([]*aflv1.PlayerInfo, len(players))
	for i, p := range players {
		infos[i] = &aflv1.PlayerInfo{
			Id:   int32(p.ID),
			Name: p.Name,
		}
	}

	return &aflv1.LookupPlayersResponse{Players: infos}, nil
}
