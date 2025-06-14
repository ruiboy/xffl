package persistence

import (
	"xffl/internal/domain/ffl"
	"xffl/internal/ports/out"
	"gorm.io/gorm"
)

type clubSeasonRepository struct {
	db *gorm.DB
}

func NewClubSeasonRepository(db *gorm.DB) out.ClubSeasonRepository {
	return &clubSeasonRepository{db: db}
}

func (r *clubSeasonRepository) FindBySeasonID(seasonID uint) ([]ffl.ClubSeason, error) {
	var clubSeasons []ffl.ClubSeason
	
	// Use raw SQL to properly handle schema prefix and joins
	err := r.db.Table("ffl.club_season cs").
		Select("cs.*, c.name as club_name").
		Joins("JOIN ffl.club c ON cs.club_id = c.id").
		Where("cs.season_id = ? AND cs.deleted_at IS NULL", seasonID).
		Order("cs.drv_premiership_points DESC").
		Scan(&clubSeasons).Error
	
	if err != nil {
		return nil, err
	}
	
	// Load club data manually since Preload doesn't work well with schema prefixes
	for i := range clubSeasons {
		var club ffl.Club
		err := r.db.Table("ffl.club").
			Where("id = ?", clubSeasons[i].ClubID).
			First(&club).Error
		if err != nil {
			return nil, err
		}
		clubSeasons[i].Club = club
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
	var clubSeason ffl.ClubSeason
	err := r.db.Table("ffl.club_season").Where("id = ? AND deleted_at IS NULL", id).First(&clubSeason).Error
	if err != nil {
		return nil, err
	}
	
	// Load club data manually
	var club ffl.Club
	err = r.db.Table("ffl.club").Where("id = ?", clubSeason.ClubID).First(&club).Error
	if err != nil {
		return nil, err
	}
	clubSeason.Club = club
	
	return &clubSeason, nil
}

func (r *clubSeasonRepository) Create(clubSeason *ffl.ClubSeason) error {
	return r.db.Create(clubSeason).Error
}

func (r *clubSeasonRepository) Update(clubSeason *ffl.ClubSeason) error {
	return r.db.Save(clubSeason).Error
}

func (r *clubSeasonRepository) Delete(id uint) error {
	return r.db.Where("id = ?", id).Delete(&ffl.ClubSeason{}).Error
}
