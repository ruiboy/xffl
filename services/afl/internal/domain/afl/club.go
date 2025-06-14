package afl

import (
	"time"
)

// Club represents an AFL club entity
type Club struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	Name         string    `json:"name" gorm:"not null"`
	Abbreviation string    `json:"abbreviation" gorm:"not null;unique"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// TableName specifies the table name for this model
func (Club) TableName() string {
	return "afl.club"
}