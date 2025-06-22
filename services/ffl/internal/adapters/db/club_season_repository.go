package db

import (
	"time"
	"xffl/services/ffl/internal/domain/ffl"
	"gorm.io/gorm"
)

// FFLClubSeason represents the database model for ClubSeason
type FFLClubSeason struct {
	gorm.Model
	ClubID            uint     `gorm:"column:club_id;not null"`
	SeasonID          uint     `gorm:"column:season_id;not null"`
	Club              FFLClub  `gorm:"foreignKey:ClubID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Played            int      `gorm:"column:drv_played;default:0"`
	Won               int      `gorm:"column:drv_won;default:0"`
	Lost              int      `gorm:"column:drv_lost;default:0"`
	Drawn             int      `gorm:"column:drv_drawn;default:0"`
	PointsFor         int      `gorm:"column:drv_for;default:0"`
	PointsAgainst     int      `gorm:"column:drv_against;default:0"`
	ExtraPoints       int      `gorm:"column:drv_extra_points;default:0"`
	PremiershipPoints int      `gorm:"column:drv_premiership_points;default:0"`
}

// TableName specifies the table name for FFLClubSeason
func (*FFLClubSeason) TableName() string {
	return "ffl.club_season"
}

// ToDomain converts FFLClubSeason to ffl.ClubSeason
func (cs *FFLClubSeason) ToDomain() ffl.ClubSeason {
	var deletedAt *time.Time
	if cs.DeletedAt.Valid {
		deletedAt = &cs.DeletedAt.Time
	}
	
	return ffl.ClubSeason{
		ID:                cs.ID,
		CreatedAt:         cs.CreatedAt,
		UpdatedAt:         cs.UpdatedAt,
		DeletedAt:         deletedAt,
		ClubID:            cs.ClubID,
		SeasonID:          cs.SeasonID,
		Club:              cs.Club.ToDomain(),
		Played:            cs.Played,
		Won:               cs.Won,
		Lost:              cs.Lost,
		Drawn:             cs.Drawn,
		PointsFor:         cs.PointsFor,
		PointsAgainst:     cs.PointsAgainst,
		ExtraPoints:       cs.ExtraPoints,
		PremiershipPoints: cs.PremiershipPoints,
	}
}

// FromDomain converts ffl.ClubSeason to FFLClubSeason
func (cs *FFLClubSeason) FromDomain(clubSeason *ffl.ClubSeason) {
	cs.ID = clubSeason.ID
	cs.CreatedAt = clubSeason.CreatedAt
	cs.UpdatedAt = clubSeason.UpdatedAt
	if clubSeason.DeletedAt != nil {
		cs.DeletedAt = gorm.DeletedAt{Time: *clubSeason.DeletedAt, Valid: true}
	}
	cs.ClubID = clubSeason.ClubID
	cs.SeasonID = clubSeason.SeasonID
	cs.Played = clubSeason.Played
	cs.Won = clubSeason.Won
	cs.Lost = clubSeason.Lost
	cs.Drawn = clubSeason.Drawn
	cs.PointsFor = clubSeason.PointsFor
	cs.PointsAgainst = clubSeason.PointsAgainst
	cs.ExtraPoints = clubSeason.ExtraPoints
	cs.PremiershipPoints = clubSeason.PremiershipPoints
}

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
