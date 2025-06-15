package ffl

import (
	"time"
)

type ClubSeason struct {
	ID                 uint      `json:"id" gorm:"primaryKey"`
	CreatedAt          time.Time `json:"createdAt"`
	UpdatedAt          time.Time `json:"updatedAt"`
	DeletedAt          *time.Time `json:"deletedAt,omitempty" gorm:"index"`
	ClubID             uint      `json:"clubId" gorm:"not null"`
	SeasonID           uint      `json:"seasonId" gorm:"not null"`
	Club               Club      `json:"club" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Played             int       `json:"played" gorm:"column:drv_played;default:0"`
	Won                int       `json:"won" gorm:"column:drv_won;default:0"`
	Lost               int       `json:"lost" gorm:"column:drv_lost;default:0"`
	Drawn              int       `json:"drawn" gorm:"column:drv_drawn;default:0"`
	PointsFor          int       `json:"pointsFor" gorm:"column:drv_for;default:0"`
	PointsAgainst      int       `json:"pointsAgainst" gorm:"column:drv_against;default:0"`
	ExtraPoints        int       `json:"extraPoints" gorm:"column:drv_extra_points;default:0"`
	PremiershipPoints  int       `json:"premiershipPoints" gorm:"column:drv_premiership_points;default:0"`
}

func (cs *ClubSeason) TableName() string {
	return "ffl.club_season"
}

func (cs *ClubSeason) Percentage() float64 {
	if cs.PointsAgainst == 0 {
		if cs.PointsFor > 0 {
			return 999.9 // Maximum percentage when no points against
		}
		return 0.0
	}
	return float64(cs.PointsFor) / float64(cs.PointsAgainst) * 100.0
}