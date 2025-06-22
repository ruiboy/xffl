package domain

import "time"

// PlayerMatch represents an AFL player's performance in a specific match
type PlayerMatch struct {
	ID             uint      `json:"id"`
	PlayerSeasonID uint      `json:"player_season_id"`
	ClubMatchID    uint      `json:"club_match_id"`
	Kicks          int       `json:"kicks"`
	Handballs      int       `json:"handballs"`
	Marks          int       `json:"marks"`
	Hitouts        int       `json:"hitouts"`
	Tackles        int       `json:"tackles"`
	Goals          int       `json:"goals"`
	Behinds        int       `json:"behinds"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}