package services

import (
	"xffl/services/afl/internal/domain/afl"
)

// clubRepository defines the interface for club data operations needed by ClubService
type clubRepository interface {
	FindAll() ([]afl.Club, error)
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
func (s *ClubService) GetAllClubs() ([]afl.Club, error) {
	return s.clubRepo.FindAll()
}