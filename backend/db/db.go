package db

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

// FFLClub represents a fantasy football club
type FFLClub struct {
	gorm.Model
	Name    string      `gorm:"uniqueIndex;not null"`
	Players []FFLPlayer `gorm:"foreignKey:ClubID"`
}

// TableName specifies the table name for FFLClub
func (FFLClub) TableName() string {
	return "ffl.club"
}

// FFLPlayer represents a fantasy football player
type FFLPlayer struct {
	gorm.Model
	Name   string  `gorm:"not null"`
	ClubID uint    `gorm:"not null"`
	Club   FFLClub `gorm:"foreignKey:ClubID"`
}

// TableName specifies the table name for FFLPlayer
func (FFLPlayer) TableName() string {
	return "ffl.player"
}

// InitDB initializes the database connection
func InitDB() {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		getEnvOrDefault("DB_HOST", "localhost"),
		getEnvOrDefault("DB_USER", "postgres"),
		getEnvOrDefault("DB_PASSWORD", "postgres"),
		getEnvOrDefault("DB_NAME", "gffl"),
		getEnvOrDefault("DB_PORT", "5432"),
	)

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
