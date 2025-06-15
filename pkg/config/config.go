package config

import "os"

// Config holds application configuration
type Config struct {
	Port     string
	Database DatabaseConfig
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Host     string
	User     string
	Password string
	Name     string
	Port     string
}

// Load loads configuration from environment variables
func Load() *Config {
	return &Config{
		Port: getEnvOrDefault("PORT", "8080"),
		Database: DatabaseConfig{
			Host:     getEnvOrDefault("DB_HOST", "localhost"),
			User:     getEnvOrDefault("DB_USER", "postgres"),
			Password: getEnvOrDefault("DB_PASSWORD", "postgres"),
			Name:     getEnvOrDefault("DB_NAME", "xffl"),
			Port:     getEnvOrDefault("DB_PORT", "5432"),
		},
	}
}

// getEnvOrDefault gets environment variable or returns default value
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}