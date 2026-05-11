package twirp

import (
	"context"

	aflv1 "xffl/contracts/gen/afl/v1"
	"xffl/services/afl/internal/domain"
)

type playerLookupServer struct {
	players        domain.PlayerRepository
	playerSeasons  domain.PlayerSeasonRepository
	playerMatches  domain.PlayerMatchRepository
}

func NewPlayerLookupServer(players domain.PlayerRepository, playerSeasons domain.PlayerSeasonRepository, playerMatches domain.PlayerMatchRepository) aflv1.PlayerLookup {
	return &playerLookupServer{players: players, playerSeasons: playerSeasons, playerMatches: playerMatches}
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

func (s *playerLookupServer) LookupPlayerMatch(ctx context.Context, req *aflv1.LookupPlayerMatchRequest) (*aflv1.LookupPlayerMatchResponse, error) {
	switch k := req.Key.(type) {
	case *aflv1.LookupPlayerMatchRequest_ByIds:
		ids := make([]int, len(k.ByIds.Ids))
		for i, id := range k.ByIds.Ids {
			ids[i] = int(id)
		}
		pms, err := s.playerMatches.FindByIDs(ctx, ids)
		if err != nil {
			return nil, err
		}
		return &aflv1.LookupPlayerMatchResponse{Stats: toProtoStats(pms)}, nil

	case *aflv1.LookupPlayerMatchRequest_BySeasonRound:
		psIDs := make([]int, len(k.BySeasonRound.PlayerSeasonIds))
		for i, id := range k.BySeasonRound.PlayerSeasonIds {
			psIDs[i] = int(id)
		}
		pms, err := s.playerMatches.FindByPlayerSeasonIDsAndRoundID(ctx, psIDs, int(k.BySeasonRound.RoundId))
		if err != nil {
			return nil, err
		}
		return &aflv1.LookupPlayerMatchResponse{Stats: toProtoStats(pms)}, nil

	default:
		return &aflv1.LookupPlayerMatchResponse{}, nil
	}
}

func toProtoStats(pms []domain.PlayerMatch) []*aflv1.PlayerMatchStats {
	stats := make([]*aflv1.PlayerMatchStats, len(pms))
	for i, pm := range pms {
		stats[i] = &aflv1.PlayerMatchStats{
			Id:             int32(pm.ID),
			Status:         domain.ComputeAFLPlayerMatchStatus(domain.MatchDataStatus(pm.MatchDataStatus)),
			Goals:          int32(pm.Goals),
			Kicks:          int32(pm.Kicks),
			Handballs:      int32(pm.Handballs),
			Marks:          int32(pm.Marks),
			Tackles:        int32(pm.Tackles),
			Hitouts:        int32(pm.Hitouts),
			PlayerSeasonId: int32(pm.PlayerSeasonID),
		}
	}
	return stats
}
