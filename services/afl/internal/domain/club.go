package domain

import "time"

// Club represents an AFL club entity
type Club struct {
	ID           uint      `json:"id"`
	Name         string    `json:"name"`
	Abbreviation string    `json:"abbreviation"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
