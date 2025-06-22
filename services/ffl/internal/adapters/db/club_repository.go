package db

import (
	"time"
	"xffl/services/ffl/internal/domain/ffl"
	"gorm.io/gorm"
)

// FFLClub represents the database model for Club
type FFLClub struct {
	gorm.Model
	Name    string      `gorm:"uniqueIndex;not null"`
	Players []FFLPlayer `gorm:"foreignKey:ClubID"`
}

// TableName specifies the table name for FFLClub
func (*FFLClub) TableName() string {
	return "ffl.club"
}

// ToDomain converts FFLClub to ffl.Club
func (c *FFLClub) ToDomain() ffl.Club {
	players := make([]ffl.Player, len(c.Players))
	for i, p := range c.Players {
		players[i] = p.ToDomain()
	}
	
	var deletedAt *time.Time
	if c.DeletedAt.Valid {
		deletedAt = &c.DeletedAt.Time
	}
	
	return ffl.Club{
		ID:        c.ID,
		Name:      c.Name,
		Players:   players,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
		DeletedAt: deletedAt,
	}
}

// FromDomain converts ffl.Club to FFLClub
func (c *FFLClub) FromDomain(club *ffl.Club) {
	c.ID = club.ID
	c.Name = club.Name
	c.CreatedAt = club.CreatedAt
	c.UpdatedAt = club.UpdatedAt
	if club.DeletedAt != nil {
		c.DeletedAt = gorm.DeletedAt{Time: *club.DeletedAt, Valid: true}
	}
}

// ClubRepository implements club database operations
type ClubRepository struct {
	db *gorm.DB
}

// NewClubRepository creates a new ClubRepository
func NewClubRepository(db *gorm.DB) *ClubRepository {
	return &ClubRepository{
		db: db,
	}
}

// FindAll retrieves all clubs from the database
func (r *ClubRepository) FindAll() ([]ffl.Club, error) {
	var fflClubs []FFLClub
	err := r.db.Preload("Players").Find(&fflClubs).Error
	if err != nil {
		return nil, err
	}
	
	clubs := make([]ffl.Club, len(fflClubs))
	for i, fflClub := range fflClubs {
		clubs[i] = fflClub.ToDomain()
	}
	
	return clubs, nil
}

// FindByID retrieves a club by its ID
func (r *ClubRepository) FindByID(id uint) (*ffl.Club, error) {
	var fflClub FFLClub
	err := r.db.Preload("Players").First(&fflClub, id).Error
	if err != nil {
		return nil, err
	}
	
	club := fflClub.ToDomain()
	return &club, nil
}

// Create creates a new club in the database
func (r *ClubRepository) Create(club *ffl.Club) error {
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
func (r *ClubRepository) Update(club *ffl.Club) error {
	var fflClub FFLClub
	fflClub.FromDomain(club)
	
	return r.db.Save(&fflClub).Error
}

// Delete deletes a club by its ID
func (r *ClubRepository) Delete(id uint) error {
	return r.db.Delete(&FFLClub{}, id).Error
}
