package application

import (
	"xffl/services/afl/internal/domain/afl"
	"xffl/services/afl/internal/ports/in"
	"xffl/services/afl/internal/ports/out"
)

// ClubService implements the ClubUseCase interface
type ClubService struct {
	clubRepo out.ClubRepository
}

// NewClubService creates a new ClubService
func NewClubService(clubRepo out.ClubRepository) in.ClubUseCase {
	return &ClubService{
		clubRepo: clubRepo,
	}
}

// GetAllClubs retrieves all clubs
func (s *ClubService) GetAllClubs() ([]afl.Club, error) {
	return s.clubRepo.FindAll()
}