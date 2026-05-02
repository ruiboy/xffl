package twirp

import (
	"context"

	aflv1 "xffl/contracts/gen/afl/v1"
	"xffl/services/afl/internal/domain"
)

type playerLookupServer struct {
	players       domain.PlayerRepository
	playerSeasons domain.PlayerSeasonRepository
}

func NewPlayerLookupServer(players domain.PlayerRepository, playerSeasons domain.PlayerSeasonRepository) aflv1.PlayerLookup {
	return &playerLookupServer{players: players, playerSeasons: playerSeasons}
}

func (s *playerLookupServer) LookupPlayers(ctx context.Context, req *aflv1.LookupPlayersRequest) (*aflv1.LookupPlayersResponse, error) {
	ids := make([]int, len(req.AflPlayerIds))
	for i, id := range req.AflPlayerIds {
		ids[i] = int(id)
	}

	players, err := s.players.FindByIDsWithClub(ctx, ids)
	if err != nil {
		return nil, err
	}

	infos := make([]*aflv1.PlayerInfo, len(players))
	for i, p := range players {
		infos[i] = &aflv1.PlayerInfo{
			Id:       int32(p.ID),
			Name:     p.Name,
			ClubName: p.ClubName,
		}
	}

	return &aflv1.LookupPlayersResponse{Players: infos}, nil
}

func (s *playerLookupServer) LookupPlayerSeason(ctx context.Context, req *aflv1.LookupPlayerSeasonRequest) (*aflv1.LookupPlayerSeasonResponse, error) {
	ps, err := s.playerSeasons.FindByID(ctx, int(req.PlayerSeasonId))
	if err != nil {
		return nil, err
	}
	return &aflv1.LookupPlayerSeasonResponse{PlayerId: int32(ps.PlayerID)}, nil
}
