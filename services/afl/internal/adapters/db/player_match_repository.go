package db

import (
	"time"
	"gorm.io/gorm"
	"xffl/services/afl/internal/domain/afl"
)

// PlayerMatchRepository implements player match database operations
type PlayerMatchRepository struct {
	db *gorm.DB
}

// NewPlayerMatchRepository creates a new player match repository
func NewPlayerMatchRepository(db *gorm.DB) *PlayerMatchRepository {
	return &PlayerMatchRepository{db: db}
}

// PlayerMatchEntity represents the database model for player_match
type PlayerMatchEntity struct {
	ID             uint      `gorm:"primaryKey"`
	PlayerSeasonID uint      `gorm:"column:player_season_id;not null"`
	ClubMatchID    uint      `gorm:"column:club_match_id;not null"`
	Kicks          int       `gorm:"column:kicks;default:0"`
	Handballs      int       `gorm:"column:handballs;default:0"`
	Marks          int       `gorm:"column:marks;default:0"`
	Hitouts        int       `gorm:"column:hitouts;default:0"`
	Tackles        int       `gorm:"column:tackles;default:0"`
	Goals          int       `gorm:"column:goals;default:0"`
	Behinds        int       `gorm:"column:behinds;default:0"`
	CreatedAt      time.Time `gorm:"column:created_at"`
	UpdatedAt      time.Time `gorm:"column:updated_at"`
	DeletedAt      *time.Time `gorm:"column:deleted_at"`
}

// TableName specifies the table name for GORM
func (PlayerMatchEntity) TableName() string {
	return "afl.player_match"
}

// UpdatePlayerMatch updates or creates a player match record using UPSERT
func (r *PlayerMatchRepository) UpdatePlayerMatch(playerSeasonID, clubMatchID uint, stats afl.PlayerMatch) (*afl.PlayerMatch, error) {
	// Use ON CONFLICT DO UPDATE (upsert)
	result := r.db.Exec(`
		INSERT INTO afl.player_match (player_season_id, club_match_id, kicks, handballs, marks, hitouts, tackles, goals, behinds, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW())
		ON CONFLICT (player_season_id, club_match_id)
		DO UPDATE SET
			kicks = EXCLUDED.kicks,
			handballs = EXCLUDED.handballs,
			marks = EXCLUDED.marks,
			hitouts = EXCLUDED.hitouts,
			tackles = EXCLUDED.tackles,
			goals = EXCLUDED.goals,
			behinds = EXCLUDED.behinds,
			updated_at = NOW()
	`, playerSeasonID, clubMatchID, stats.Kicks, stats.Handballs, stats.Marks, stats.Hitouts, stats.Tackles, stats.Goals, stats.Behinds)

	if result.Error != nil {
		return nil, result.Error
	}

	// Fetch the updated/created record
	return r.FindByPlayerSeasonAndClubMatch(playerSeasonID, clubMatchID)
}

// FindByPlayerSeasonAndClubMatch finds a player match by player season and club match IDs
func (r *PlayerMatchRepository) FindByPlayerSeasonAndClubMatch(playerSeasonID, clubMatchID uint) (*afl.PlayerMatch, error) {
	var entity PlayerMatchEntity
	
	result := r.db.Where("player_season_id = ? AND club_match_id = ? AND deleted_at IS NULL", playerSeasonID, clubMatchID).First(&entity)
	if result.Error != nil {
		return nil, result.Error
	}

	return r.entityToDomain(entity), nil
}

// entityToDomain converts database entity to domain model
func (r *PlayerMatchRepository) entityToDomain(entity PlayerMatchEntity) *afl.PlayerMatch {
	return &afl.PlayerMatch{
		ID:             entity.ID,
		PlayerSeasonID: entity.PlayerSeasonID,
		ClubMatchID:    entity.ClubMatchID,
		Kicks:          entity.Kicks,
		Handballs:      entity.Handballs,
		Marks:          entity.Marks,
		Hitouts:        entity.Hitouts,
		Tackles:        entity.Tackles,
		Goals:          entity.Goals,
		Behinds:        entity.Behinds,
		CreatedAt:      entity.CreatedAt,
		UpdatedAt:      entity.UpdatedAt,
	}
}