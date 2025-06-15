package persistence

import (
	"gorm.io/gorm"
	"xffl/services/afl/internal/domain/afl"
	"xffl/services/afl/internal/ports/out"
)

// ClubRepository implements the ClubRepository interface
type ClubRepository struct {
	db *gorm.DB
}

// NewClubRepository creates a new ClubRepository
func NewClubRepository(db *gorm.DB) out.ClubRepository {
	return &ClubRepository{db: db}
}

// FindAll retrieves all clubs from the database
func (r *ClubRepository) FindAll() ([]afl.Club, error) {
	var clubs []afl.Club
	err := r.db.Find(&clubs).Error
	return clubs, err
}

// FindByID retrieves a club by its ID
func (r *ClubRepository) FindByID(id uint) (*afl.Club, error) {
	var club afl.Club
	err := r.db.First(&club, id).Error
	if err != nil {
		return nil, err
	}
	return &club, nil
}