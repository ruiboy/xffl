package services

import (
	"xffl/services/ffl/internal/domain/ffl"
)

// clubRepository defines the interface for club data operations needed by ClubService
type clubRepository interface {
	FindAll() ([]ffl.Club, error)
	FindByID(id uint) (*ffl.Club, error)
	Create(club *ffl.Club) error
	Update(club *ffl.Club) error
	Delete(id uint) error
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
func (s *ClubService) GetAllClubs() ([]ffl.Club, error) {
	return s.clubRepo.FindAll()
}

// GetClubByID retrieves a club by its ID
func (s *ClubService) GetClubByID(id uint) (*ffl.Club, error) {
	return s.clubRepo.FindByID(id)
}

// CreateClub creates a new club
func (s *ClubService) CreateClub(name string) (*ffl.Club, error) {
	club := ffl.NewClub(name)
	err := s.clubRepo.Create(club)
	if err != nil {
		return nil, err
	}
	return club, nil
}

// UpdateClub updates an existing club
func (s *ClubService) UpdateClub(id uint, name string) (*ffl.Club, error) {
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
