package application

import (
	"gffl/internal/domain/ffl"
	"gffl/internal/ports/in"
	"gffl/internal/ports/out"
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