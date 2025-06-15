package persistence

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

// FFLPlayer represents the database model for Player
type FFLPlayer struct {
	gorm.Model
	Name   string  `gorm:"not null"`
	ClubID uint    `gorm:"not null"`
	Club   FFLClub `gorm:"foreignKey:ClubID"`
}

// TableName specifies the table name for FFLPlayer
func (*FFLPlayer) TableName() string {
	return "ffl.player"
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

// ToDomain converts FFLPlayer to ffl.Player
func (p *FFLPlayer) ToDomain() ffl.Player {
	var deletedAt *time.Time
	if p.DeletedAt.Valid {
		deletedAt = &p.DeletedAt.Time
	}
	
	return ffl.Player{
		ID:        p.ID,
		Name:      p.Name,
		ClubID:    p.ClubID,
		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
		DeletedAt: deletedAt,
	}
}

// FromDomain converts ffl.Player to FFLPlayer
func (p *FFLPlayer) FromDomain(player *ffl.Player) {
	p.ID = player.ID
	p.Name = player.Name
	p.ClubID = player.ClubID
	p.CreatedAt = player.CreatedAt
	p.UpdatedAt = player.UpdatedAt
	if player.DeletedAt != nil {
		p.DeletedAt = gorm.DeletedAt{Time: *player.DeletedAt, Valid: true}
	}
}

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
