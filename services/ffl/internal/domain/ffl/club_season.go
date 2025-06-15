package ffl

import (
	"time"
)

type ClubSeason struct {
	ID                uint       `json:"id"`
	CreatedAt         time.Time  `json:"createdAt"`
	UpdatedAt         time.Time  `json:"updatedAt"`
	DeletedAt         *time.Time `json:"deletedAt,omitempty"`
	ClubID            uint       `json:"clubId"`
	SeasonID          uint       `json:"seasonId"`
	Club              Club       `json:"club"`
	Played            int        `json:"played"`
	Won               int        `json:"won"`
	Lost              int        `json:"lost"`
	Drawn             int        `json:"drawn"`
	PointsFor         int        `json:"pointsFor"`
	PointsAgainst     int        `json:"pointsAgainst"`
	ExtraPoints       int        `json:"extraPoints"`
	PremiershipPoints int        `json:"premiershipPoints"`
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