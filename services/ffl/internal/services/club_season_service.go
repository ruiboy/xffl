package services

import (
	"xffl/services/ffl/internal/domain/ffl"
)

// clubSeasonRepository defines the interface for club season data operations needed by ClubSeasonService
type clubSeasonRepository interface {
	FindBySeasonID(seasonID uint) ([]ffl.ClubSeason, error)
}

// ClubSeasonService implements club season business logic
type ClubSeasonService struct {
	clubSeasonRepo clubSeasonRepository
}

// NewClubSeasonService creates a new ClubSeasonService
func NewClubSeasonService(clubSeasonRepo clubSeasonRepository) *ClubSeasonService {
	return &ClubSeasonService{
		clubSeasonRepo: clubSeasonRepo,
	}
}

// GetLadderBySeasonID retrieves the ladder for a given season
func (s *ClubSeasonService) GetLadderBySeasonID(seasonID uint) ([]ffl.ClubSeason, error) {
	return s.clubSeasonRepo.FindBySeasonID(seasonID)
}
