package db

import (
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// DatabaseConfig holds the PostgreSQL connection configuration.
type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// DefaultConfig returns the default database configuration.
// Override individual fields before passing to Connect.
func DefaultConfig() *DatabaseConfig {
	return &DatabaseConfig{
		Host:     "postgres",
		Port:     "5432",
		User:     "postgres",
		Password: "postgres",
		DBName:   "app",
		SSLMode:  "disable",
	}
}

// DSN returns the PostgreSQL connection string.
func (c *DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode,
	)
}

// Connect establishes a connection to the PostgreSQL database using GORM.
func Connect(cfg *DatabaseConfig) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	return db, nil
}
