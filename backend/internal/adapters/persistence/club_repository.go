package persistence

import (
	"gffl/internal/domain"
	"gorm.io/gorm"
)

// ClubRepositoryImpl implements the ClubRepository interface
type ClubRepositoryImpl struct {
	db *gorm.DB
}

// NewClubRepository creates a new ClubRepositoryImpl
func NewClubRepository(db *gorm.DB) *ClubRepositoryImpl {
	return &ClubRepositoryImpl{
		db: db,
	}
}

// FindAll retrieves all clubs from the database
func (r *ClubRepositoryImpl) FindAll() ([]domain.Club, error) {
	var fflClubs []FFLClub
	err := r.db.Preload("Players").Find(&fflClubs).Error
	if err != nil {
		return nil, err
	}
	
	clubs := make([]domain.Club, len(fflClubs))
	for i, fflClub := range fflClubs {
		clubs[i] = fflClub.ToDomain()
	}
	
	return clubs, nil
}

// FindByID retrieves a club by its ID
func (r *ClubRepositoryImpl) FindByID(id uint) (*domain.Club, error) {
	var fflClub FFLClub
	err := r.db.Preload("Players").First(&fflClub, id).Error
	if err != nil {
		return nil, err
	}
	
	club := fflClub.ToDomain()
	return &club, nil
}

// Create creates a new club in the database
func (r *ClubRepositoryImpl) Create(club *domain.Club) error {
	var fflClub FFLClub
	fflClub.FromDomain(club)
	
	err := r.db.Create(&fflClub).Error
	if err != nil {
		return err
	}
	
	// Update the domain entity with the generated ID
	club.ID = fflClub.ID
	club.CreatedAt = fflClub.CreatedAt
	club.UpdatedAt = fflClub.UpdatedAt
	
	return nil
}

// Update updates an existing club in the database
func (r *ClubRepositoryImpl) Update(club *domain.Club) error {
	var fflClub FFLClub
	fflClub.FromDomain(club)
	
	return r.db.Save(&fflClub).Error
}

// Delete deletes a club by its ID
func (r *ClubRepositoryImpl) Delete(id uint) error {
	return r.db.Delete(&FFLClub{}, id).Error
}