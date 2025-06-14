package application

import (
	"xffl/services/ffl/internal/domain/ffl"
	"xffl/services/ffl/internal/ports/in"
	"xffl/services/ffl/internal/ports/out"
)

type clubSeasonService struct {
	clubSeasonRepo out.ClubSeasonRepository
}

func NewClubSeasonService(clubSeasonRepo out.ClubSeasonRepository) in.ClubSeasonUseCase {
	return &clubSeasonService{
		clubSeasonRepo: clubSeasonRepo,
	}
}

func (s *clubSeasonService) GetLadderBySeasonID(seasonID uint) ([]ffl.ClubSeason, error) {
	return s.clubSeasonRepo.FindBySeasonID(seasonID)
}
