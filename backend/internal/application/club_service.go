package application

import (
	"gffl/internal/domain"
	"gffl/internal/ports/out"
)

// ClubService implements the ClubUseCase interface
type ClubService struct {
	clubRepo out.ClubRepository
}

// NewClubService creates a new ClubService
func NewClubService(clubRepo out.ClubRepository) *ClubService {
	return &ClubService{
		clubRepo: clubRepo,
	}
}

// GetAllClubs retrieves all clubs
func (s *ClubService) GetAllClubs() ([]domain.Club, error) {
	return s.clubRepo.FindAll()
}

// GetClubByID retrieves a club by its ID
func (s *ClubService) GetClubByID(id uint) (*domain.Club, error) {
	return s.clubRepo.FindByID(id)
}

// CreateClub creates a new club
func (s *ClubService) CreateClub(name string) (*domain.Club, error) {
	club := domain.NewClub(name)
	err := s.clubRepo.Create(club)
	if err != nil {
		return nil, err
	}
	return club, nil
}

// UpdateClub updates an existing club
func (s *ClubService) UpdateClub(id uint, name string) (*domain.Club, error) {
	club, err := s.clubRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	
	club.UpdateName(name)
	err = s.clubRepo.Update(club)
	if err != nil {
		return nil, err
	}
	
	return club, nil
}

// DeleteClub deletes a club by its ID
func (s *ClubService) DeleteClub(id uint) error {
	return s.clubRepo.Delete(id)
}