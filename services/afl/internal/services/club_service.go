package services

import (
	"xffl/services/afl/internal/domain"
)

// clubRepository defines the interface for club data operations needed by ClubService
type clubRepository interface {
	FindAll() ([]domain.Club, error)
}

// ClubService implements club business logic
type ClubService struct {
	clubRepo clubRepository
}

// NewClubService creates a new ClubService
func NewClubService(clubRepo clubRepository) *ClubService {
	return &ClubService{
		clubRepo: clubRepo,
	}
}

// GetAllClubs retrieves all clubs
func (s *ClubService) GetAllClubs() ([]domain.Club, error) {
	return s.clubRepo.FindAll()
}