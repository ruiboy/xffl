package persistence

import (
	"xffl/services/ffl/internal/domain/ffl"
	"xffl/services/ffl/internal/ports/out"
	"gorm.io/gorm"
)

type clubSeasonRepository struct {
	db *gorm.DB
}

func NewClubSeasonRepository(db *gorm.DB) out.ClubSeasonRepository {
	return &clubSeasonRepository{db: db}
}

func (r *clubSeasonRepository) FindBySeasonID(seasonID uint) ([]ffl.ClubSeason, error) {
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

func (r *clubSeasonRepository) FindByID(id uint) (*ffl.ClubSeason, error) {
	var entity FFLClubSeason
	err := r.db.Preload("Club").Where("id = ? AND deleted_at IS NULL", id).First(&entity).Error
	if err != nil {
		return nil, err
	}
	
	clubSeason := entity.ToDomain()
	return &clubSeason, nil
}

func (r *clubSeasonRepository) Create(clubSeason *ffl.ClubSeason) error {
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

func (r *clubSeasonRepository) Update(clubSeason *ffl.ClubSeason) error {
	var entity FFLClubSeason
	entity.FromDomain(clubSeason)
	
	return r.db.Save(&entity).Error
}

func (r *clubSeasonRepository) Delete(id uint) error {
	return r.db.Where("id = ?", id).Delete(&FFLClubSeason{}).Error
}
