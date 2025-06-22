package db

import (
	"xffl/services/ffl/internal/domain/ffl"
	"gorm.io/gorm"
)

// ClubSeasonRepository implements club season database operations
type ClubSeasonRepository struct {
	db *gorm.DB
}

// NewClubSeasonRepository creates a new ClubSeasonRepository
func NewClubSeasonRepository(db *gorm.DB) *ClubSeasonRepository {
	return &ClubSeasonRepository{db: db}
}

func (r *ClubSeasonRepository) FindBySeasonID(seasonID uint) ([]ffl.ClubSeason, error) {
	var entities []FFLClubSeason
	
	err := r.db.Preload("Club").
		Where("season_id = ? AND deleted_at IS NULL", seasonID).
		Order("drv_premiership_points DESC").
		Find(&entities).Error
	
	if err != nil {
		return nil, err
	}
	
	// Convert to domain entities
	clubSeasons := make([]ffl.ClubSeason, len(entities))
	for i, entity := range entities {
		clubSeasons[i] = entity.ToDomain()
	}
	
	// Secondary sort by percentage (calculated field)
	// Since we can't easily ORDER BY a calculated field in GORM, we'll sort in Go
	// This is acceptable for ladder data which typically has < 20 teams
	for i := 0; i < len(clubSeasons)-1; i++ {
		for j := i + 1; j < len(clubSeasons); j++ {
			// If premiership points are equal, sort by percentage
			if clubSeasons[i].PremiershipPoints == clubSeasons[j].PremiershipPoints {
				if clubSeasons[i].Percentage() < clubSeasons[j].Percentage() {
					clubSeasons[i], clubSeasons[j] = clubSeasons[j], clubSeasons[i]
				}
			}
		}
	}
	
	return clubSeasons, nil
}

func (r *ClubSeasonRepository) FindByID(id uint) (*ffl.ClubSeason, error) {
	var entity FFLClubSeason
	err := r.db.Preload("Club").Where("id = ? AND deleted_at IS NULL", id).First(&entity).Error
	if err != nil {
		return nil, err
	}
	
	clubSeason := entity.ToDomain()
	return &clubSeason, nil
}

func (r *ClubSeasonRepository) Create(clubSeason *ffl.ClubSeason) error {
	var entity FFLClubSeason
	entity.FromDomain(clubSeason)
	
	err := r.db.Create(&entity).Error
	if err != nil {
		return err
	}
	
	// Update the domain entity with generated values
	clubSeason.ID = entity.ID
	clubSeason.CreatedAt = entity.CreatedAt
	clubSeason.UpdatedAt = entity.UpdatedAt
	
	return nil
}

func (r *ClubSeasonRepository) Update(clubSeason *ffl.ClubSeason) error {
	var entity FFLClubSeason
	entity.FromDomain(clubSeason)
	
	return r.db.Save(&entity).Error
}

func (r *ClubSeasonRepository) Delete(id uint) error {
	return r.db.Where("id = ?", id).Delete(&FFLClubSeason{}).Error
}
