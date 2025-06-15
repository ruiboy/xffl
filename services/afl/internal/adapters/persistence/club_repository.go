package persistence

import (
	"time"
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

// ClubEntity represents the database model for club
type ClubEntity struct {
	ID           uint      `gorm:"primaryKey"`
	Name         string    `gorm:"column:name;not null"`
	Abbreviation string    `gorm:"column:abbreviation;not null;unique"`
	CreatedAt    time.Time `gorm:"column:created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at"`
	DeletedAt    *time.Time `gorm:"column:deleted_at"`
}

// TableName specifies the table name for GORM
func (ClubEntity) TableName() string {
	return "afl.club"
}

// FindAll retrieves all clubs from the database
func (r *ClubRepository) FindAll() ([]afl.Club, error) {
	var entities []ClubEntity
	err := r.db.Where("deleted_at IS NULL").Find(&entities).Error
	if err != nil {
		return nil, err
	}

	clubs := make([]afl.Club, len(entities))
	for i, entity := range entities {
		clubs[i] = *r.entityToDomain(entity)
	}
	return clubs, nil
}

// FindByID retrieves a club by its ID
func (r *ClubRepository) FindByID(id uint) (*afl.Club, error) {
	var entity ClubEntity
	err := r.db.Where("id = ? AND deleted_at IS NULL", id).First(&entity).Error
	if err != nil {
		return nil, err
	}
	return r.entityToDomain(entity), nil
}

// entityToDomain converts database entity to domain model
func (r *ClubRepository) entityToDomain(entity ClubEntity) *afl.Club {
	return &afl.Club{
		ID:           entity.ID,
		Name:         entity.Name,
		Abbreviation: entity.Abbreviation,
		CreatedAt:    entity.CreatedAt,
		UpdatedAt:    entity.UpdatedAt,
	}
}