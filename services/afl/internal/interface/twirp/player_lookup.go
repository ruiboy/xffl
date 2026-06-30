package twirp

import (
	"context"

	aflv1 "xffl/contracts/gen/afl/v1"
	"xffl/services/afl/internal/domain"
)

type playerLookupServer struct {
	players       domain.PlayerRepository
	playerSeasons domain.PlayerSeasonRepository
	playerMatches domain.PlayerMatchRepository
	byes          domain.ByeRepository
}

func NewPlayerLookupServer(players domain.PlayerRepository, playerSeasons domain.PlayerSeasonRepository, playerMatches domain.PlayerMatchRepository, byes domain.ByeRepository) aflv1.PlayerLookup {
	return &playerLookupServer{players: players, playerSeasons: playerSeasons, playerMatches: playerMatches, byes: byes}
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

func (s *playerLookupServer) LookupByeInfo(ctx context.Context, req *aflv1.LookupByeInfoRequest) (*aflv1.LookupByeInfoResponse, error) {
	psIDs := make([]int, len(req.PlayerSeasonIds))
	for i, id := range req.PlayerSeasonIds {
		psIDs[i] = int(id)
	}
	roundID := int(req.RoundId)

	byeStatus, err := s.playerMatches.FindByeStatusBatch(ctx, psIDs, roundID)
	if err != nil {
		return nil, err
	}

	// Collect player_season_ids that have a bye — only fetch averages for those.
	var byePSIDs []int
	for _, bs := range byeStatus {
		if bs.HasBye {
			byePSIDs = append(byePSIDs, bs.PlayerSeasonID)
		}
	}

	avgsByPS := make(map[int]domain.PlayerSeasonAverages)
	if len(byePSIDs) > 0 {
		avgs, err := s.playerMatches.GetSeasonAveragesBatch(ctx, byePSIDs, roundID)
		if err != nil {
			return nil, err
		}
		for _, a := range avgs {
			avgsByPS[a.PlayerSeasonID] = a
		}
	}

	infos := make([]*aflv1.ByePlayerInfo, len(byeStatus))
	for i, bs := range byeStatus {
		info := &aflv1.ByePlayerInfo{
			PlayerSeasonId: int32(bs.PlayerSeasonID),
			HasBye:         bs.HasBye,
			PlayedLast:     bs.PlayedLast,
		}
		if a, ok := avgsByPS[bs.PlayerSeasonID]; ok {
			info.AvgGoals = a.Goals
			info.AvgKicks = a.Kicks
			info.AvgHandballs = a.Handballs
			info.AvgMarks = a.Marks
			info.AvgTackles = a.Tackles
			info.AvgHitouts = a.Hitouts
		}
		infos[i] = info
	}

	return &aflv1.LookupByeInfoResponse{Players: infos}, nil
}

func toProtoStats(pms []domain.PlayerMatch) []*aflv1.PlayerMatchStats {
	stats := make([]*aflv1.PlayerMatchStats, len(pms))
	for i, pm := range pms {
		stats[i] = &aflv1.PlayerMatchStats{
			Id:             int32(pm.ID),
			Status:         pm.AFLPlayerMatchStatus(),
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
